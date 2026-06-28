package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/xcreativs/gigmann/internal/core/approval"
	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/ports"
)

// ErrForbidden is returned when a principal lacks the role for an action.
var ErrForbidden = errors.New("app: forbidden")

// ApprovalService is the decision-routing use case. Authorization lives here:
// only executives may decide approvals (spec §5.8, §7).
type ApprovalService struct {
	repo  ports.ApprovalRepository
	audit ports.AuditLogger
}

// NewApprovalService wires the approval use case to its repository.
func NewApprovalService(repo ports.ApprovalRepository, audit ports.AuditLogger) *ApprovalService {
	return &ApprovalService{repo: repo, audit: audit}
}

// List returns the approvals the principal may see: executives see the whole
// network; a facility manager sees only their own facility's approvals (no IDOR).
func (s *ApprovalService) List(ctx context.Context, p auth.Principal) ([]approval.Approval, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("app: list approvals: %w", err)
	}
	if p.IsExecutive() {
		return items, nil
	}
	scoped := make([]approval.Approval, 0, len(items))
	for _, a := range items {
		if p.CanAccessFacility(a.FacilityID) {
			scoped = append(scoped, a)
		}
	}
	return scoped, nil
}

// Decide records an approve/decline decision. Only executives are authorized;
// the decision is an explicit, user-initiated side-effect (never AI-triggered).
func (s *ApprovalService) Decide(
	ctx context.Context, p auth.Principal, id string, approved bool, note string, at time.Time,
) (approval.Approval, error) {
	if !p.IsExecutive() {
		s.audit.Record(ctx, ports.AuditEvent{Actor: p.UserID, Action: "approval.decide", Target: id, Outcome: "forbidden"})
		return approval.Approval{}, ErrForbidden
	}
	current, err := s.repo.Get(ctx, id)
	if err != nil {
		return approval.Approval{}, err
	}
	decided, err := current.Decide(approved, note, at)
	if err != nil {
		return approval.Approval{}, err
	}
	if err := s.repo.Save(ctx, decided); err != nil {
		return approval.Approval{}, fmt.Errorf("app: save approval: %w", err)
	}
	s.audit.Record(ctx, ports.AuditEvent{Actor: p.UserID, Action: "approval.decide", Target: id, Outcome: string(decided.Status)})
	return decided, nil
}
