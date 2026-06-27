package app

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/ports"
)

// maxPrefEntries caps watched-metrics and thresholds so a request cannot bloat
// the stored preferences blob.
const maxPrefEntries = 24

// PreferencesService is the personalisation use case: read and update the
// current user's watched metrics and thresholds (spec §5.12).
type PreferencesService struct {
	users ports.UserRepository
}

// NewPreferencesService wires the service to the user repository.
func NewPreferencesService(users ports.UserRepository) *PreferencesService {
	return &PreferencesService{users: users}
}

// Get returns the user's stored preferences.
func (s *PreferencesService) Get(ctx context.Context, userID string) (user.Preferences, error) {
	acct, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return user.Preferences{}, fmt.Errorf("app: load preferences: %w", err)
	}
	return acct.User.Preferences, nil
}

// Update sanitises and persists new preferences, returning the stored result.
func (s *PreferencesService) Update(ctx context.Context, userID string, prefs user.Preferences) (user.Preferences, error) {
	clean := sanitizePreferences(prefs)
	acct, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return user.Preferences{}, fmt.Errorf("app: load account: %w", err)
	}
	acct.User.Preferences = clean
	if err := s.users.Save(ctx, acct); err != nil {
		return user.Preferences{}, fmt.Errorf("app: save preferences: %w", err)
	}
	return clean, nil
}

// sanitizePreferences trims, de-duplicates, drops empties / non-finite values,
// and caps the number of entries (input validation at the app boundary).
func sanitizePreferences(p user.Preferences) user.Preferences {
	metrics := make([]string, 0, len(p.WatchedMetrics))
	seen := make(map[string]bool, len(p.WatchedMetrics))
	for _, m := range p.WatchedMetrics {
		m = strings.TrimSpace(m)
		if m == "" || seen[m] {
			continue
		}
		seen[m] = true
		metrics = append(metrics, m)
		if len(metrics) >= maxPrefEntries {
			break
		}
	}
	thresholds := make(map[string]float64, len(p.Thresholds))
	for k, v := range p.Thresholds {
		k = strings.TrimSpace(k)
		if k == "" || math.IsNaN(v) || math.IsInf(v, 0) || len(thresholds) >= maxPrefEntries {
			continue
		}
		thresholds[k] = v
	}
	return user.Preferences{WatchedMetrics: metrics, Thresholds: thresholds}
}
