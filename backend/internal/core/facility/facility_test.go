package facility_test

import (
	"errors"
	"testing"

	"github.com/xcreativs/gigmann/internal/core/facility"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		fname   string
		region  facility.Region
		beds    int
		status  facility.Status
		wantErr error
	}{
		{"valid", "f1", "Kasoa Polyclinic", "Central", 40, facility.StatusGood, nil},
		{"trims and accepts", "  f2  ", "  Adansi  ", "Ashanti", 30, facility.StatusWatch, nil},
		{"empty id", "", "X", "Central", 1, facility.StatusGood, facility.ErrEmptyID},
		{"empty name", "f1", "   ", "Central", 1, facility.StatusGood, facility.ErrEmptyName},
		{"empty region", "f1", "X", "  ", 1, facility.StatusGood, facility.ErrEmptyRegion},
		{"negative beds", "f1", "X", "Central", -1, facility.StatusGood, facility.ErrNegativeBeds},
		{"bad status", "f1", "X", "Central", 1, facility.Status("nope"), facility.ErrInvalidStatus},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := facility.New(tt.id, tt.fname, tt.region, "Town", tt.beds, tt.status)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want err %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if f.ID == "" || f.Name == "" {
				t.Errorf("fields not set: %+v", f)
			}
		})
	}
}

func TestStatusValid(t *testing.T) {
	for _, s := range []facility.Status{facility.StatusGood, facility.StatusWatch, facility.StatusCritical} {
		if !s.Valid() {
			t.Errorf("%q should be valid", s)
		}
	}
	if facility.Status("x").Valid() {
		t.Error("invalid status reported as valid")
	}
}
