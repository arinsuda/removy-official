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
