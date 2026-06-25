package payer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/core/payer"
)

func TestNew(t *testing.T) {
	m, err := payer.New(65, 25, 10)
	require.NoError(t, err)
	assert.Equal(t, 65, m.NHIS)
	assert.Equal(t, 25, m.CashMoMo)
	assert.Equal(t, 10, m.Private)
	assert.True(t, m.Valid())
}

func TestNewRejectsBadInput(t *testing.T) {
	_, err := payer.New(50, 25, 10)
	require.Error(t, err)
	_, err = payer.New(-1, 51, 50)
	require.Error(t, err)
	assert.False(t, (payer.Mix{NHIS: 10, CashMoMo: 10, Private: 10}).Valid())
}
