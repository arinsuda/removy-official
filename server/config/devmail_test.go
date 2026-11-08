package config

import "testing"

func TestDevMailLogAllowed(t *testing.T) {
	local := []string{"http://localhost:3000", "http://127.0.0.1:5173", "http://[::1]:3000"}
	cases := []struct {
		name    string
		origins []string
		bases   []string
		want    bool
	}{
		{"all local", local, []string{"http://localhost:3000", "http://127.0.0.1:3000"}, true},
		{"non-local origin", append([]string{"https://popcorn.example"}, local...), []string{"http://localhost:3000"}, false},
		{"non-local base", local, []string{"http://localhost:3000", "https://popcorn.example"}, false},
		{"lookalike host", []string{"http://localhost.evil.com"}, []string{"http://localhost:3000"}, false},
		{"empty origins", nil, []string{"http://localhost:3000"}, false},
		{"empty base", local, []string{""}, false},
		{"no bases", local, nil, false},
	}
	for _, c := range cases {
		if got := DevMailLogAllowed(c.origins, c.bases...); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}
