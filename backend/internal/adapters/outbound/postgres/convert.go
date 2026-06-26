package postgres

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/xcreativs/gigmann/internal/core/user"
)

// refreshTokenBytes is the entropy of a raw refresh token before encoding.
const refreshTokenBytes = 32

// normalizeEmail lower-cases and trims an email so lookups are case-insensitive.
// Stored values are normalised too, so the UNIQUE(email) constraint matches.
func normalizeEmail(e string) string { return strings.ToLower(strings.TrimSpace(e)) }

// hashToken returns the SHA-256 hash of a raw token. Only hashes are persisted;
// the raw token is shown to the client once and never stored.
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// nullableStr maps "" to a SQL NULL (*string nil) and any other value to a pointer.
func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// derefStr maps a nullable *string back to a plain string ("" for NULL).
func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// normalizeTime canonicalises a time to UTC at microsecond resolution — the
// precision Postgres timestamptz stores — so a Save->Get round-trip is exact
// (pgx scans back in time.Local at microsecond precision otherwise).
func normalizeTime(t time.Time) time.Time { return t.UTC().Truncate(time.Microsecond) }

// tsRequired wraps a time for a NOT NULL timestamptz column (always valid).
func tsRequired(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: normalizeTime(t), Valid: true}
}

// tsOptional maps the zero time to SQL NULL and any other time to a valid value.
func tsOptional(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: normalizeTime(t), Valid: true}
}

// timeFromTS maps a nullable timestamptz back to a time (zero for NULL), in UTC.
func timeFromTS(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time.UTC()
}

// i16/i32 narrow small, domain-bounded ints for storage. Ranges are guaranteed
// by domain invariants and table CHECK constraints (e.g. payer 0-100).
func i16(n int) int16 { return int16(n) } //nolint:gosec // bounded by domain invariants + schema CHECKs
func i32(n int) int32 { return int32(n) } //nolint:gosec // bounded by domain invariants + schema CHECKs

// prefsJSON is the persistence shape for user.Preferences. The adapter owns its
// own serialization DTO so the core domain struct stays free of encoding tags.
type prefsJSON struct {
	WatchedMetrics []string           `json:"watched_metrics"`
	Thresholds     map[string]float64 `json:"thresholds"`
}

func marshalPrefs(p user.Preferences) ([]byte, error) {
	b, err := json.Marshal(prefsJSON{WatchedMetrics: p.WatchedMetrics, Thresholds: p.Thresholds})
	if err != nil {
		return nil, fmt.Errorf("postgres: marshal preferences: %w", err)
	}
	return b, nil
}

func unmarshalPrefs(b []byte) (user.Preferences, error) {
	if len(b) == 0 {
		return user.Preferences{}, nil
	}
	var dto prefsJSON
	if err := json.Unmarshal(b, &dto); err != nil {
		return user.Preferences{}, fmt.Errorf("postgres: unmarshal preferences: %w", err)
	}
	return user.Preferences{WatchedMetrics: dto.WatchedMetrics, Thresholds: dto.Thresholds}, nil
}
