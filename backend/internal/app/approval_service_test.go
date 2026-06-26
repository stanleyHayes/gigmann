package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/approval"
	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/ports"
	"github.com/xcreativs/gigmann/internal/ports/mocks"
)

func executive() auth.Principal {
	return auth.Principal{UserID: "u1", Name: "Sammy", Role: user.RoleExecutive}
}

func manager() auth.Principal {
	return auth.Principal{UserID: "u2", Name: "Ama", Role: user.RoleFacilityManager, FacilityID: "kasoa"}
}

func pending() approval.Approval {
	return approval.Approval{ID: "ap1", Type: approval.TypeCapital, Title: "Ultrasound", Status: approval.StatusPending}
}

func TestApprovalList(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockApprovalRepository(ctrl)
	repo.EXPECT().List(gomock.Any()).Return([]approval.Approval{pending()}, nil)

	got, err := app.NewApprovalService(repo, auditMock(ctrl)).List(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
}

func TestApprovalDecideApproves(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockApprovalRepository(ctrl)
	repo.EXPECT().Get(gomock.Any(), "ap1").Return(pending(), nil)
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

	at := time.Date(2026, 6, 26, 8, 0, 0, 0, time.UTC)
	out, err := app.NewApprovalService(repo, auditMock(ctrl)).Decide(context.Background(), executive(), "ap1", true, "Go ahead", at)
	require.NoError(t, err)
	assert.Equal(t, approval.StatusApproved, out.Status)
	assert.Equal(t, "Go ahead", out.DecisionNote)
}

func TestApprovalDecideForbiddenForManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockApprovalRepository(ctrl) // no calls expected

	_, err := app.NewApprovalService(repo, auditMock(ctrl)).Decide(context.Background(), manager(), "ap1", true, "", time.Now())
	assert.ErrorIs(t, err, app.ErrForbidden)
}

func TestApprovalDecideNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockApprovalRepository(ctrl)
	repo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(approval.Approval{}, ports.ErrApprovalNotFound)

	_, err := app.NewApprovalService(repo, auditMock(ctrl)).Decide(context.Background(), executive(), "missing", true, "", time.Now())
	assert.ErrorIs(t, err, ports.ErrApprovalNotFound)
}

func TestApprovalDecideAlreadyDecided(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockApprovalRepository(ctrl)
	decided := pending()
	decided.Status = approval.StatusApproved
	repo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(decided, nil)

	_, err := app.NewApprovalService(repo, auditMock(ctrl)).Decide(context.Background(), executive(), "ap1", false, "", time.Now())
	assert.ErrorIs(t, err, approval.ErrAlreadyDecided)
}

func TestDecideAudits(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockApprovalRepository(ctrl)
	repo.EXPECT().Get(gomock.Any(), "ap1").Return(pending(), nil)
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
	auditL := mocks.NewMockAuditLogger(ctrl)
	auditL.EXPECT().Record(gomock.Any(), ports.AuditEvent{Actor: "u1", Action: "approval.decide", Target: "ap1", Outcome: "approved"})

	svc := app.NewApprovalService(repo, auditL)
	_, err := svc.Decide(context.Background(), executive(), "ap1", true, "ok", time.Now())
	require.NoError(t, err)
}
