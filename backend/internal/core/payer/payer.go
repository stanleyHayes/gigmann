// Package payer models a facility's payer mix (spec §4.2): the split between
// NHIS, cash & mobile money, and private insurance, as percentages summing 100.
package payer

import "fmt"

// Mix is a payer split in whole-percent points.
type Mix struct {
	NHIS     int
	CashMoMo int
	Private  int
}

// New validates and builds a Mix. Each share must be 0..100 and the three must sum to 100.
func New(nhis, cashMoMo, private int) (Mix, error) {
	for _, v := range []int{nhis, cashMoMo, private} {
		if v < 0 || v > 100 {
			return Mix{}, fmt.Errorf("payer: share out of range: %d", v)
		}
	}
	if nhis+cashMoMo+private != 100 {
		return Mix{}, fmt.Errorf("payer: shares must sum to 100, got %d", nhis+cashMoMo+private)
	}
	return Mix{NHIS: nhis, CashMoMo: cashMoMo, Private: private}, nil
}

// Valid reports whether the mix is internally consistent.
func (m Mix) Valid() bool {
	_, err := New(m.NHIS, m.CashMoMo, m.Private)
	return err == nil
}
