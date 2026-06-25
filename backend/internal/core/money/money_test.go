package money_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xcreativs/gigmann/internal/core/money"
)

func TestConstructionAndAccessors(t *testing.T) {
	c := money.FromCedis(1234, 56)
	assert.Equal(t, int64(123456), c.Pesewas())
	assert.InDelta(t, 1234.56, c.Float(), 0.001)
	assert.Equal(t, int64(50), money.FromPesewas(50).Pesewas())
}

func TestArithmetic(t *testing.T) {
	a := money.FromCedis(100, 0)
	b := money.FromCedis(40, 50)
	assert.Equal(t, int64(14050), a.Add(b).Pesewas())
	assert.True(t, b.Sub(a).IsNegative())
	assert.False(t, a.IsNegative())
}

func TestString(t *testing.T) {
	cases := map[money.Cedis]string{
		money.FromCedis(0, 5):       "GH₵ 0.05",
		money.FromCedis(7, 0):       "GH₵ 7.00",
		money.FromCedis(1234, 56):   "GH₵ 1,234.56",
		money.FromCedis(1000000, 0): "GH₵ 1,000,000.00",
		money.FromCedis(-85000, 0):  "-GH₵ 85,000.00",
	}
	for amount, want := range cases {
		assert.Equal(t, want, amount.String())
	}
}
