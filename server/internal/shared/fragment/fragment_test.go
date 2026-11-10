package fragment

import (
	"net/url"
	"strings"
	"testing"
)

func TestURL(t *testing.T) {
	got := URL("http://x.test", "/reset-password", url.Values{"token": {"a b&c"}, "uid": {"7"}})
	if got != "http://x.test/reset-password#token=a+b%26c&uid=7" {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(got, "?") {
		t.Fatal("must not contain a query string")
	}
}
