package passwordhash_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/passwordhash"
)

func TestHashAndVerify(t *testing.T) {
	h := passwordhash.New()

	encoded, err := h.Hash("correct horse battery staple")
	require.NoError(t, err)
	assert.Contains(t, encoded, "$argon2id$")

	ok, err := h.Verify("correct horse battery staple", encoded)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = h.Verify("wrong password", encoded)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestHashIsSalted(t *testing.T) {
	h := passwordhash.New()
	a, err := h.Hash("same")
	require.NoError(t, err)
	b, err := h.Hash("same")
	require.NoError(t, err)
	assert.NotEqual(t, a, b, "each hash must use a fresh salt")
}

func TestVerifyMalformed(t *testing.T) {
	h := passwordhash.New()
	for _, bad := range []string{"", "not-a-hash", "$argon2id$v=19$bad$bad$bad"} {
		_, err := h.Verify("x", bad)
		assert.Error(t, err)
	}
}
