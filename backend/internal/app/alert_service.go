package app

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/xcreativs/gigmann/internal/core/alert"
	"github.com/xcreativs/gigmann/internal/core/auth"
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
func (s *AlertService) Feed(ctx context.Context, p auth.Principal, cursor string, limit int) ([]alert.Alert, string, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("app: list alerts: %w", err)
	}
	open := make([]alert.Alert, 0, len(all))
	for _, a := range all {
		// Facility managers see only their own facility's alerts (no IDOR);
		// executives see the whole network.
		if a.Status == alert.StatusOpen && p.CanAccessFacility(a.FacilityID) {
			open = append(open, a)
		}
	}
	open = dedupAlerts(open)
	sort.SliceStable(open, func(i, j int) bool { return alertSortLess(open[i], open[j]) })

	// Keyset cursor on the full sort position (rank, createdAt, id), so paging is
	// resilient to an alert changing status between pages: we skip past the cursor
	// position rather than matching an exact id that may have dropped off the feed.
	start := cursorStart(open, cursor)
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
		next = encodeAlertCursor(page[len(page)-1])
	}
	return page, next, nil
}

// UpdateStatus dismisses or resolves an alert (the only valid transitions). A
// facility manager may only act on their own facility's alerts (ErrForbidden).
func (s *AlertService) UpdateStatus(ctx context.Context, p auth.Principal, id string, status alert.Status) (alert.Alert, error) {
	a, err := s.repo.Get(ctx, id)
	if err != nil {
		return alert.Alert{}, err
	}
	if !p.CanAccessFacility(a.FacilityID) {
		return alert.Alert{}, ErrForbidden
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

// dedupAlerts collapses open alerts sharing a (facility_id, type) to the most
// recent one, so a recurring condition surfaces once in the feed.
func dedupAlerts(alerts []alert.Alert) []alert.Alert {
	latest := make(map[string]int) // facility|type -> index in out
	out := make([]alert.Alert, 0, len(alerts))
	for _, a := range alerts {
		key := a.FacilityID + "|" + a.Type
		if i, ok := latest[key]; ok {
			if a.CreatedAt.After(out[i].CreatedAt) {
				out[i] = a
			}
			continue
		}
		latest[key] = len(out)
		out = append(out, a)
	}
	return out
}

// alertCursor is the keyset position: the full sort key of the last item on a page.
type alertCursor struct {
	rank      int
	createdAt int64 // unix nanoseconds
	id        string
}

// alertSortLess orders the feed worst-first: severity rank desc, then newest, then id.
func alertSortLess(a, b alert.Alert) bool {
	if ra, rb := a.Severity.Rank(), b.Severity.Rank(); ra != rb {
		return ra > rb
	}
	if !a.CreatedAt.Equal(b.CreatedAt) {
		return a.CreatedAt.After(b.CreatedAt)
	}
	return a.ID < b.ID
}

// afterAlertCursor reports whether a sorts strictly after the cursor position
// (used to resume paging even if the cursor's own alert has dropped off the feed).
func afterAlertCursor(a alert.Alert, c alertCursor) bool {
	if r := a.Severity.Rank(); r != c.rank {
		return r < c.rank
	}
	if ts := a.CreatedAt.UnixNano(); ts != c.createdAt {
		return ts < c.createdAt
	}
	return a.ID > c.id
}

// cursorStart returns the index of the first alert that sorts after the cursor,
// or 0 when the cursor is empty/invalid (restart from the top).
func cursorStart(open []alert.Alert, cursor string) int {
	if cursor == "" {
		return 0
	}
	c, ok := decodeAlertCursor(cursor)
	if !ok {
		return 0
	}
	for i := range open {
		if afterAlertCursor(open[i], c) {
			return i
		}
	}
	return len(open)
}

func encodeAlertCursor(a alert.Alert) string {
	s := strconv.Itoa(a.Severity.Rank()) + "|" +
		strconv.FormatInt(a.CreatedAt.UnixNano(), 10) + "|" + a.ID
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

func decodeAlertCursor(cursor string) (alertCursor, bool) {
	b, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return alertCursor{}, false
	}
	parts := strings.SplitN(string(b), "|", 3)
	if len(parts) != 3 {
		return alertCursor{}, false
	}
	rank, err1 := strconv.Atoi(parts[0])
	ts, err2 := strconv.ParseInt(parts[1], 10, 64)
	if err1 != nil || err2 != nil {
		return alertCursor{}, false
	}
	return alertCursor{rank: rank, createdAt: ts, id: parts[2]}, true
}
