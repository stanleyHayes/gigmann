package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/postgres/sqlcgen"
	"github.com/xcreativs/gigmann/internal/core/approval"
	"github.com/xcreativs/gigmann/internal/core/money"
	"github.com/xcreativs/gigmann/internal/ports"
)

// ApprovalRepo is a PostgreSQL implementation of ports.ApprovalRepository.
type ApprovalRepo struct {
	q *sqlcgen.Queries
}

var _ ports.ApprovalRepository = (*ApprovalRepo)(nil)

// NewApprovalRepo builds an ApprovalRepo over a pgx pool (or any sqlcgen.DBTX).
func NewApprovalRepo(db sqlcgen.DBTX) *ApprovalRepo {
	return &ApprovalRepo{q: sqlcgen.New(db)}
}

// List returns all approvals ordered by creation, mapped to the domain model.
func (r *ApprovalRepo) List(ctx context.Context) ([]approval.Approval, error) {
	rows, err := r.q.ListApprovals(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres: list approvals: %w", err)
	}
	out := make([]approval.Approval, 0, len(rows))
	for _, row := range rows {
		a, ferr := approvalFromModel(row)
		if ferr != nil {
			return nil, fmt.Errorf("postgres: map approval %q: %w", row.ID, ferr)
		}
		out = append(out, a)
	}
	return out, nil
}

// Get returns the approval with the given id, or ErrApprovalNotFound.
func (r *ApprovalRepo) Get(ctx context.Context, id string) (approval.Approval, error) {
	row, err := r.q.GetApproval(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return approval.Approval{}, ports.ErrApprovalNotFound
	}
	if err != nil {
		return approval.Approval{}, fmt.Errorf("postgres: get approval: %w", err)
	}
	return approvalFromModel(row)
}

// Save upserts an approval.
func (r *ApprovalRepo) Save(ctx context.Context, a approval.Approval) error {
	if err := r.q.UpsertApproval(ctx, approvalParams(a)); err != nil {
		return fmt.Errorf("postgres: upsert approval %q: %w", a.ID, err)
	}
	return nil
}

func approvalParams(a approval.Approval) sqlcgen.UpsertApprovalParams {
	return sqlcgen.UpsertApprovalParams{
		ID:           a.ID,
		Type:         string(a.Type),
		FacilityID:   nullableStr(a.FacilityID),
		Amount:       a.Amount.Pesewas(),
		Title:        a.Title,
		Context:      a.Context,
		RequestedBy:  a.RequestedBy,
		Status:       string(a.Status),
		DecidedAt:    tsOptional(a.DecidedAt),
		DecisionNote: a.DecisionNote,
		CreatedAt:    tsRequired(a.CreatedAt),
	}
}

func approvalFromModel(m sqlcgen.Approval) (approval.Approval, error) {
	return approval.New(approval.Approval{
		ID:           m.ID,
		Type:         approval.Type(m.Type),
		FacilityID:   derefStr(m.FacilityID),
		Amount:       money.FromPesewas(m.Amount),
		Title:        m.Title,
		Context:      m.Context,
		RequestedBy:  m.RequestedBy,
		Status:       approval.Status(m.Status),
		DecidedAt:    timeFromTS(m.DecidedAt),
		DecisionNote: m.DecisionNote,
		CreatedAt:    timeFromTS(m.CreatedAt),
	})
}
