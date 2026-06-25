package money_test

import (
	"testing"

	"github.com/xcreativs/gigmann/internal/core/money"
)

func TestConstructionAndAccessors(t *testing.T) {
	c := money.FromCedis(1234, 56)
	if c.Pesewas() != 123456 {
		t.Errorf("want 123456 pesewas, got %d", c.Pesewas())
	}
	if c.Float() != 1234.56 {
		t.Errorf("want 1234.56, got %v", c.Float())
	}
	if money.FromPesewas(50).Pesewas() != 50 {
		t.Error("FromPesewas mismatch")
	}
}

func TestArithmetic(t *testing.T) {
	a := money.FromCedis(100, 0)
	b := money.FromCedis(40, 50)
	if got := a.Add(b); got.Pesewas() != 14050 {
		t.Errorf("add: got %d", got.Pesewas())
	}
	if got := b.Sub(a); !got.IsNegative() {
		t.Error("expected negative result")
	}
	if a.IsNegative() {
		t.Error("100 cedis should not be negative")
	}
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
		if got := amount.String(); got != want {
			t.Errorf("String(%d) = %q, want %q", amount.Pesewas(), got, want)
		}
	}
}
