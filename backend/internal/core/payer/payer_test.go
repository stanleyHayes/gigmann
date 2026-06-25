package payer_test

import (
	"testing"

	"github.com/xcreativs/gigmann/internal/core/payer"
)

func TestNew(t *testing.T) {
	m, err := payer.New(65, 25, 10)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if m.NHIS != 65 || m.CashMoMo != 25 || m.Private != 10 {
		t.Errorf("unexpected mix: %+v", m)
	}
	if !m.Valid() {
		t.Error("mix should be valid")
	}
}

func TestNewRejectsBadInput(t *testing.T) {
	if _, err := payer.New(50, 25, 10); err == nil {
		t.Error("expected error: shares do not sum to 100")
	}
	if _, err := payer.New(-1, 51, 50); err == nil {
		t.Error("expected error: negative share")
	}
	if (payer.Mix{NHIS: 10, CashMoMo: 10, Private: 10}).Valid() {
		t.Error("inconsistent mix should be invalid")
	}
}
