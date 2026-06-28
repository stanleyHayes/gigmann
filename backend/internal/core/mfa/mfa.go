// Package mfa implements TOTP (RFC 6238, HMAC-SHA1, 30s step, 6 digits) for
// optional two-factor auth. Pure domain code (stdlib crypto only).
package mfa

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" //nolint:gosec // TOTP is defined on HMAC-SHA1 (RFC 6238); not used for hashing secrets
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"time"
)

const (
	period     = 30
	digits     = 6
	secretLen  = 20
	digitsMod  = 1_000_000 // 10^digits
	skewWindow = 1         // accept the adjacent step on each side
)

// NewSecret returns a fresh base32-encoded TOTP secret (authenticator-compatible).
func NewSecret() (string, error) {
	buf := make([]byte, secretLen)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("mfa: generate secret: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf), nil
}

// Code returns the TOTP code for the secret at time t.
func Code(secretB32 string, t time.Time) (string, error) {
	return codeForCounter(secretB32, counter(t))
}

// ValidateAt reports whether code is valid for the secret at time t (±1 step
// skew, constant-time compare) and, when valid, returns the matching step
// counter. Callers enforce single-use (replay protection) by recording the
// returned counter and rejecting any code whose counter is not strictly greater
// than the last one already consumed for that user.
func ValidateAt(secretB32, code string, t time.Time) (uint64, bool) {
	c := counter(t)
	for skew := int64(-skewWindow); skew <= skewWindow; skew++ {
		step := uint64(int64(c) + skew) //nolint:gosec // small positive step index
		want, err := codeForCounter(secretB32, step)
		if err != nil {
			return 0, false
		}
		if subtle.ConstantTimeCompare([]byte(want), []byte(code)) == 1 {
			return step, true
		}
	}
	return 0, false
}

// Validate reports whether code is valid for the secret at time t (±1 step skew).
func Validate(secretB32, code string, t time.Time) bool {
	_, ok := ValidateAt(secretB32, code, t)
	return ok
}

// OTPAuthURI builds the otpauth:// URI an authenticator app scans.
func OTPAuthURI(secretB32, account, issuer string) string {
	v := url.Values{}
	v.Set("secret", secretB32)
	v.Set("issuer", issuer)
	label := url.PathEscape(issuer + ":" + account)
	return "otpauth://totp/" + label + "?" + v.Encode()
}

func counter(t time.Time) uint64 { return uint64(t.Unix()) / period } //nolint:gosec // unix seconds are non-negative

func codeForCounter(secretB32 string, c uint64) (string, error) {
	secret, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secretB32)
	if err != nil {
		return "", fmt.Errorf("mfa: decode secret: %w", err)
	}
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, c)
	h := hmac.New(sha1.New, secret)
	_, _ = h.Write(buf)
	sum := h.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	truncated := (uint32(sum[offset]&0x7f) << 24) |
		(uint32(sum[offset+1]) << 16) |
		(uint32(sum[offset+2]) << 8) |
		uint32(sum[offset+3])
	return fmt.Sprintf("%0*d", digits, truncated%digitsMod), nil
}
