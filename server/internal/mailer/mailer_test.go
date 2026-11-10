package mailer

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

const recipient = "private.person@example.test"

func TestSendNeverLogsOrReturnsTheRecipient(t *testing.T) {
	cases := map[string]struct {
		enabled bool
		rt      rtFunc
	}{
		"disabled": {false, nil},
		"sent": {true, func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 201, Body: io.NopCloser(strings.NewReader("{}")), Request: r}, nil
		}},
		"transport error": {true, func(*http.Request) (*http.Response, error) { return nil, errors.New("boom") }},
		"brevo 500": {true, func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader("{}")), Request: r}, nil
		}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			t.Cleanup(func() { log.SetOutput(os.Stderr) })
			m := &Mailer{enabled: c.enabled, httpClient: &http.Client{Transport: c.rt}}
			err := m.send(recipient, "Person", "Verify your email", "<p>x</p>")
			out := buf.String()
			if err != nil {
				out += err.Error()
			}
			if strings.Contains(out, recipient) || strings.Contains(out, "@") {
				t.Fatalf("recipient leaked: %q", out)
			}
			if out == "" {
				t.Fatal("expected a log line or error")
			}
		})
	}
}

func TestDevLogPrintsAndSkipsHTTP(t *testing.T) {
	var buf bytes.Buffer
	m := &Mailer{enabled: true, devLog: true, devOut: &buf, httpClient: &http.Client{Transport: rtFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("HTTP call made while dev log active")
		return nil, nil
	})}}
	if !m.IsEnabled() {
		t.Fatal("IsEnabled must be true when dev log active")
	}
	if err := m.SendVerificationEmail(recipient, "P", "http://localhost:3000/auth/verify-email#token=abc"); err != nil {
		t.Fatal(err)
	}
	if err := m.SendLinkCodeEmail(recipient, "P", "123456"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"[DEV MAIL] to=" + recipient + " kind=verify", "token=abc", "kind=link_code", "123456"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %q", want, out)
		}
	}
}

func TestInactiveSendsNormally(t *testing.T) {
	var buf bytes.Buffer
	called := false
	m := &Mailer{enabled: true, devOut: &buf, httpClient: &http.Client{Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
		called = true
		return &http.Response{StatusCode: 201, Body: io.NopCloser(strings.NewReader("{}")), Request: r}, nil
	})}}
	if err := m.SendLinkCodeEmail(recipient, "P", "123456"); err != nil {
		t.Fatal(err)
	}
	if !called || buf.Len() != 0 {
		t.Fatalf("called=%v printed=%q", called, buf.String())
	}
}
