package app

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sort"

	"github.com/xcreativs/gigmann/internal/core/alert"
	"github.com/xcreativs/gigmann/internal/ports"
)

// ErrInvalidAlertStatus is returned when an update targets a non-terminal status.
var ErrInvalidAlertStatus = errors.New("app: alert status must be dismissed or resolved")

const (
	defaultAlertLimit = 20
	maxAlertLimit     = 50
)

// AlertService is the attention-feed use case: a ranked, cursor-paginated feed of
// open alerts, and dismiss/resolve transitions. Resolved/dismissed alerts drop
// off the feed.
type AlertService struct {
	repo ports.AlertRepository
}

// NewAlertService wires the service to an alert repository.
func NewAlertService(repo ports.AlertRepository) *AlertService {
	return &AlertService{repo: repo}
}

// Feed returns open alerts ranked worst-first (severity, then newest, then id),
// cursor-paginated. The returned cursor is empty when there are no more pages.
func (s *AlertService) Feed(ctx context.Context, cursor string, limit int) ([]alert.Alert, string, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("app: list alerts: %w", err)
	}
	open := make([]alert.Alert, 0, len(all))
	for _, a := range all {
		if a.Status == alert.StatusOpen {
			open = append(open, a)
		}
	}
	sort.SliceStable(open, func(i, j int) bool {
		if ri, rj := open[i].Severity.Rank(), open[j].Severity.Rank(); ri != rj {
			return ri > rj
		}
		if !open[i].CreatedAt.Equal(open[j].CreatedAt) {
			return open[i].CreatedAt.After(open[j].CreatedAt)
		}
		return open[i].ID < open[j].ID
	})

	start := 0
	if cursor != "" {
		if id, ok := decodeAlertCursor(cursor); ok {
			for i := range open {
				if open[i].ID == id {
					start = i + 1
					break
				}
			}
		}
	}
	if limit <= 0 || limit > maxAlertLimit {
		limit = defaultAlertLimit
	}
	if start > len(open) {
		start = len(open)
	}
	end := min(start+limit, len(open))
	page := open[start:end]

	next := ""
	if end < len(open) && len(page) > 0 {
		next = encodeAlertCursor(page[len(page)-1].ID)
	}
	return page, next, nil
}

// UpdateStatus dismisses or resolves an alert (the only valid transitions).
func (s *AlertService) UpdateStatus(ctx context.Context, id string, status alert.Status) (alert.Alert, error) {
	a, err := s.repo.Get(ctx, id)
	if err != nil {
		return alert.Alert{}, err
	}
	var updated alert.Alert
	switch status {
	case alert.StatusDismissed:
		updated, err = a.Dismiss()
	case alert.StatusResolved:
		updated, err = a.Resolve()
	default:
		return alert.Alert{}, ErrInvalidAlertStatus
	}
	if err != nil {
		return alert.Alert{}, err
	}
	if err := s.repo.Save(ctx, updated); err != nil {
		return alert.Alert{}, fmt.Errorf("app: save alert: %w", err)
	}
	return updated, nil
}

func encodeAlertCursor(id string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(id))
}

func decodeAlertCursor(cursor string) (string, bool) {
	b, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return "", false
	}
	return string(b), true
}
