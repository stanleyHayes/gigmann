// Package money models Ghanaian cedis as integer pesewas (1 cedi = 100 pesewas)
// to avoid floating-point error in financial figures. Pure domain code.
package money

import (
	"fmt"
	"strconv"
	"strings"
)

// Cedis is a monetary amount in minor units (pesewas).
type Cedis int64

const pesewasPerCedi = 100

// FromPesewas builds a Cedis from minor units.
func FromPesewas(p int64) Cedis { return Cedis(p) }

// FromCedis builds a Cedis from whole cedis plus pesewas.
func FromCedis(cedis, pesewas int64) Cedis {
	return Cedis(cedis*pesewasPerCedi + pesewas)
}

// Pesewas returns the amount in minor units.
func (c Cedis) Pesewas() int64 { return int64(c) }

// Float returns the amount in cedis (for display/serialisation only).
func (c Cedis) Float() float64 { return float64(c) / pesewasPerCedi }

// Add returns c + o.
func (c Cedis) Add(o Cedis) Cedis { return c + o }

// Sub returns c - o.
func (c Cedis) Sub(o Cedis) Cedis { return c - o }

// IsNegative reports whether the amount is below zero.
func (c Cedis) IsNegative() bool { return c < 0 }

// String renders the amount as "GH₵ 1,234.56".
func (c Cedis) String() string {
	v := int64(c)
	neg := v < 0
	if neg {
		v = -v
	}
	whole := v / pesewasPerCedi
	frac := v % pesewasPerCedi
	out := fmt.Sprintf("GH₵ %s.%02d", group(whole), frac)
	if neg {
		out = "-" + out
	}
	return out
}

// group inserts thousands separators into a non-negative integer.
func group(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
		if len(s) > pre {
			b.WriteByte(',')
		}
	}
	for i := pre; i < len(s); i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < len(s) {
			b.WriteByte(',')
		}
	}
	return b.String()
}
