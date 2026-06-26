package mfa_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/core/mfa"
)

func TestRoundTrip(t *testing.T) {
	secret, err := mfa.NewSecret()
	require.NoError(t, err)
	now := time.Unix(1_700_000_000, 0)

	code, err := mfa.Code(secret, now)
	require.NoError(t, err)
	assert.Len(t, code, 6)
	assert.True(t, mfa.Validate(secret, code, now), "fresh code must validate")
}

func TestValidateAcceptsAdjacentStep(t *testing.T) {
	secret, _ := mfa.NewSecret()
	now := time.Unix(1_700_000_000, 0)
	prev, err := mfa.Code(secret, now.Add(-30*time.Second))
	require.NoError(t, err)
	assert.True(t, mfa.Validate(secret, prev, now), "previous 30s step is within the skew window")
}

func TestValidateRejectsWrongAndStaleCodes(t *testing.T) {
	secret, _ := mfa.NewSecret()
	now := time.Unix(1_700_000_000, 0)
	assert.False(t, mfa.Validate(secret, "000000", now.Add(10*time.Minute)))
	old, _ := mfa.Code(secret, now.Add(-10*time.Minute)) // far outside the window
	assert.False(t, mfa.Validate(secret, old, now))
}

func TestKnownVector(t *testing.T) {
	// RFC 6238 SHA1 test secret "12345678901234567890" at T=59 → 8-digit 94287082;
	// 6 digits = 94287082 mod 1e6 = 287082.
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ" // base32 of the ASCII test secret
	code, err := mfa.Code(secret, time.Unix(59, 0))
	require.NoError(t, err)
	assert.Equal(t, "287082", code)
}

func TestOTPAuthURI(t *testing.T) {
	uri := mfa.OTPAuthURI("ABC234", "sammy@gigmann.health", "Gigmann")
	assert.Contains(t, uri, "otpauth://totp/")
	assert.Contains(t, uri, "secret=ABC234")
	assert.Contains(t, uri, "issuer=Gigmann")
}
