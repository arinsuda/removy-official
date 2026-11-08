package main

import (
	"net/http"
	"testing"
)

func TestMergeForwardedFor(t *testing.T) {
	h := http.Header{}
	h.Add("X-Forwarded-For", "6.6.6.6")
	h.Add("X-Forwarded-For", "1.1.1.1, 2.2.2.2")
	mergeForwardedFor(h)
	if got := h.Values("X-Forwarded-For"); len(got) != 1 || got[0] != "6.6.6.6, 1.1.1.1, 2.2.2.2" {
		t.Fatalf("got %q", got)
	}
	single := http.Header{"X-Forwarded-For": {"3.3.3.3"}}
	mergeForwardedFor(single)
	if single.Get("X-Forwarded-For") != "3.3.3.3" {
		t.Fatal("single value changed")
	}
}
