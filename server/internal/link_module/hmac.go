package link_module

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"time"
)

const (
	HeaderTimestamp = "X-Link-Timestamp"
	HeaderSignature = "X-Link-Signature"
	MaxClockSkew    = 30 * time.Second
)

var ErrBadSignature = errors.New("invalid signature")

// Sign returns hex(HMAC-SHA256(secret, ts + "." + body)).
func Sign(secret, ts string, body []byte) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(ts))
	m.Write([]byte("."))
	m.Write(body)
	return hex.EncodeToString(m.Sum(nil))
}

// VerifySignature checks timestamp skew and the signature against any configured secret.
func VerifySignature(secrets []string, ts string, body []byte, sig string, now time.Time) error {
	sec, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return ErrBadSignature
	}
	d := now.Sub(time.Unix(sec, 0))
	if d < 0 {
		d = -d
	}
	if d > MaxClockSkew {
		return ErrBadSignature
	}
	given, err := hex.DecodeString(sig)
	if err != nil {
		return ErrBadSignature
	}
	ok := false
	for _, s := range secrets {
		expected, _ := hex.DecodeString(Sign(s, ts, body))
		if hmac.Equal(expected, given) {
			ok = true
		}
	}
	if !ok {
		return ErrBadSignature
	}
	return nil
}
