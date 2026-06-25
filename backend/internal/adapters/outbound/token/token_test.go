package token_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/token"
	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/user"
)

func principal() auth.Principal {
	return auth.Principal{UserID: "u1", Name: "Sammy Adjei", Role: user.RoleExecutive}
}

func TestIssueAndVerifyRoundTrip(t *testing.T) {
	svc := token.New([]byte("test-secret"), time.Hour)

	raw, err := svc.Issue(principal())
	require.NoError(t, err)

	got, err := svc.Verify(raw)
	require.NoError(t, err)
	assert.Equal(t, "u1", got.UserID)
	assert.Equal(t, "Sammy Adjei", got.Name)
	assert.Equal(t, user.RoleExecutive, got.Role)
}

func TestVerifyRejectsExpired(t *testing.T) {
	expired := token.New([]byte("test-secret"), -time.Minute)
	valid := token.New([]byte("test-secret"), time.Hour)

	raw, err := expired.Issue(principal())
	require.NoError(t, err)

	_, err = valid.Verify(raw)
	assert.ErrorIs(t, err, token.ErrInvalidToken)
}

func TestVerifyRejectsWrongSecret(t *testing.T) {
	signer := token.New([]byte("secret-a"), time.Hour)
	verifier := token.New([]byte("secret-b"), time.Hour)

	raw, err := signer.Issue(principal())
	require.NoError(t, err)

	_, err = verifier.Verify(raw)
	assert.ErrorIs(t, err, token.ErrInvalidToken)
}

func TestVerifyRejectsGarbage(t *testing.T) {
	svc := token.New([]byte("test-secret"), time.Hour)
	_, err := svc.Verify("not.a.jwt")
	assert.ErrorIs(t, err, token.ErrInvalidToken)
}
