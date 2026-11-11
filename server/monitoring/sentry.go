// Package monitoring wires optional Sentry error reporting. It is off unless SENTRY_DSN is set,
// reports only panics and 5xx responses, and never sends cookies, headers, bodies, query values
// or emails.
package monitoring

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"runtime/debug"
	"strconv"
	"time"

	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v3"
)

const (
	// FlushTimeout bounds how long shutdown waits for queued events.
	FlushTimeout = 2 * time.Second
	// MinReportedStatus is the lowest HTTP status that is reported (5xx only).
	MinReportedStatus = http.StatusInternalServerError
	redactedEmail     = "[email]"
)

var emailPattern = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)

// Options configures Init.
type Options struct {
	DSN              string
	Environment      string
	TracesSampleRate float64
}

// Init starts Sentry and returns whether it is enabled. An empty DSN is a no-op. The DSN is never logged.
func Init(o Options) bool {
	if o.DSN == "" {
		return false
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              o.DSN,
		Environment:      o.Environment,
		TracesSampleRate: o.TracesSampleRate,
		SendDefaultPII:   false,
		BeforeSend: func(e *sentry.Event, _ *sentry.EventHint) *sentry.Event {
			return ScrubEvent(e)
		},
	})
	if err != nil {
		log.Println("sentry init failed (check SENTRY_DSN)")
		return false
	}
	return true
}

// Flush waits up to FlushTimeout for pending events.
func Flush() { sentry.Flush(FlushTimeout) }

// ShouldReport reports whether an HTTP status is worth sending (5xx only).
func ShouldReport(status int) bool { return status >= MinReportedStatus }

// ScrubEvent drops request data and PII from an event and redacts emails in messages.
func ScrubEvent(e *sentry.Event) *sentry.Event {
	if e == nil {
		return nil
	}
	if e.Request != nil {
		e.Request = &sentry.Request{Method: e.Request.Method, URL: pathOnly(e.Request.URL)}
	}
	id := ""
	if e.User.ID != "" {
		id = e.User.ID
	}
	e.User = sentry.User{ID: id}
	e.Message = emailPattern.ReplaceAllString(e.Message, redactedEmail)
	for i := range e.Exception {
		e.Exception[i].Value = emailPattern.ReplaceAllString(e.Exception[i].Value, redactedEmail)
	}
	for _, b := range e.Breadcrumbs {
		b.Message = emailPattern.ReplaceAllString(b.Message, redactedEmail)
		b.Data = nil
	}
	return e
}

// pathOnly keeps scheme://host/path and drops query and fragment.
func pathOnly(raw string) string {
	for i, r := range raw {
		if r == '?' || r == '#' {
			return raw[:i]
		}
	}
	return raw
}

// Middleware captures panics and 5xx responses. A panic becomes a 500 for the client.
// Register it early so it wraps the rest of the chain.
func Middleware() fiber.Handler {
	return func(c fiber.Ctx) (err error) {
		hub := sentry.CurrentHub().Clone()
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic: %v\n%s", r, debug.Stack())
				capture(hub, c, fmt.Errorf("panic: %v", r))
				err = fiber.ErrInternalServerError
			}
		}()
		err = c.Next()

		status := c.Response().StatusCode()
		var fe *fiber.Error
		if errors.As(err, &fe) {
			status = fe.Code
		} else if err != nil {
			status = fiber.StatusInternalServerError
		}
		if ShouldReport(status) {
			if err == nil {
				err = fmt.Errorf("http %d", status)
			}
			capture(hub, c, err)
		}
		return err
	}
}

func capture(hub *sentry.Hub, c fiber.Ctx, err error) {
	hub.WithScope(func(s *sentry.Scope) {
		s.SetTag("http.method", c.Method())
		s.SetTag("http.route", routePath(c))
		if cl := mw.GetClaims(c); cl != nil && cl.UserID != 0 {
			s.SetUser(sentry.User{ID: strconv.FormatUint(uint64(cl.UserID), 10)})
		}
		hub.CaptureException(err)
	})
}

// routePath prefers the route pattern (no ids) over the raw path.
func routePath(c fiber.Ctx) string {
	if r := c.Route(); r != nil && r.Path != "" {
		return r.Path
	}
	return c.Path()
}
