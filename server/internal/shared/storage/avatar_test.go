package storage

import (
	"bytes"
	"errors"
	"testing"
)

var (
	pngHead  = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00")
	gifHead  = []byte("GIF89a\x01\x00\x01\x00\x00\x00\x00;")
	webpHead = []byte("RIFF\x24\x00\x00\x00WEBPVP8 \x18\x00\x00\x00")
)

func TestDetectAvatarType(t *testing.T) {
	cases := []struct {
		name, declared, wantType, wantExt string
		data                              []byte
		err                               bool
	}{
		{"png", "image/png", "image/png", ".png", pngHead, false},
		{"gif no declared type", "", "image/gif", ".gif", gifHead, false},
		{"webp", "image/webp", "image/webp", ".webp", webpHead, false},
		{"html named png", "image/png", "", "", []byte("<html><script>alert(1)</script></html>"), true},
		{"declared mismatch", "image/jpeg", "", "", pngHead, true},
		{"empty", "image/png", "", "", nil, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := bytes.NewReader(c.data)
			ct, ext, err := DetectAvatarType(r, c.declared)
			if c.err {
				if !errors.Is(err, ErrInvalidImage) {
					t.Fatalf("want ErrInvalidImage, got %v", err)
				}
				return
			}
			if err != nil || ct != c.wantType || ext != c.wantExt {
				t.Fatalf("got %q %q %v", ct, ext, err)
			}
			if r.Len() != len(c.data) {
				t.Fatal("reader not rewound")
			}
		})
	}
}
