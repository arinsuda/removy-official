package router

import "testing"

func TestValidShareImageURL(t *testing.T) {
	ok := []string{
		"https://image.tmdb.org/t/p/w500/abc.jpg",
		"https://IMAGE.TMDB.ORG/t/p/original/abc.jpg",
		"https://image.tmdb.org:443/t/p/w500/abc.jpg",
	}
	bad := []string{
		"",
		"http://image.tmdb.org/t/p/w500/abc.jpg",
		"https://image.tmdb.org.evil.com/t/p/w500/abc.jpg",
		"https://evil.com/t/p/w500/abc.jpg",
		"https://image.tmdb.org@127.0.0.1/t/p/w500/abc.jpg",
		"https://user:pw@image.tmdb.org/t/p/w500/abc.jpg",
		"https://image.tmdb.org:8443/t/p/w500/abc.jpg",
		"https://image.tmdb.org/other/abc.jpg",
		"https://image.tmdb.org/t/p/../../admin",
		"http://169.254.169.254/latest/meta-data",
		"http://localhost:8081/api/health",
		"file:///etc/passwd",
	}
	for _, u := range ok {
		if _, valid := validShareImageURL(u); !valid {
			t.Errorf("expected %q to be allowed", u)
		}
	}
	for _, u := range bad {
		if _, valid := validShareImageURL(u); valid {
			t.Errorf("expected %q to be rejected", u)
		}
	}
}
