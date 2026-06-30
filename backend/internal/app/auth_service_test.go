package app_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/mfa"
	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/ports"
	"github.com/xcreativs/gigmann/internal/ports/mocks"
)

func execAccount(t *testing.T) ports.Account {
	t.Helper()
	u, err := user.New(user.User{ID: "u1", Name: "Sammy Adjei", Role: user.RoleExecutive})
	require.NoError(t, err)
	return ports.Account{User: u, Email: "ceo@gigmann.health", PasswordHash: "stored-hash"}
}

func newTestAuthService(
	users ports.UserRepository,
	hasher ports.PasswordHasher,
	tokens ports.TokenService,
	refresh ports.RefreshTokenStore,
	refreshTTL time.Duration,
	audit ports.AuditLogger,
) *app.AuthService {
	return app.NewAuthService(users, hasher, tokens, refresh, memory.NewPasswordResetStore(), refreshTTL, audit)
}

func TestLoginSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := mocks.NewMockUserRepository(ctrl)
	hasher := mocks.NewMockPasswordHasher(ctrl)
	tokens := mocks.NewMockTokenService(ctrl)
	refresh := mocks.NewMockRefreshTokenStore(ctrl)

	users.EXPECT().FindByEmail(gomock.Any(), "ceo@gigmann.health").Return(execAccount(t), nil)
	hasher.EXPECT().Verify("pw", "stored-hash").Return(true, nil)
	tokens.EXPECT().Issue(gomock.Any()).Return("access-token", nil)
	refresh.EXPECT().Issue(gomock.Any(), gomock.Any(), gomock.Any()).Return("refresh-token", nil)

	svc := newTestAuthService(users, hasher, tokens, refresh, time.Hour, auditMock(ctrl))
	access, ref, p, err := svc.Login(context.Background(), "ceo@gigmann.health", "pw", "")

	require.NoError(t, err)
	assert.Equal(t, "access-token", access)
	assert.Equal(t, "refresh-token", ref)
	assert.Equal(t, "u1", p.UserID)
	assert.Equal(t, user.RoleExecutive, p.Role)
}

func TestLoginUnknownEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := mocks.NewMockUserRepository(ctrl)
	users.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(ports.Account{}, ports.ErrAccountNotFound)

	svc := newTestAuthService(users, mocks.NewMockPasswordHasher(ctrl), mocks.NewMockTokenService(ctrl), mocks.NewMockRefreshTokenStore(ctrl), time.Hour, auditMock(ctrl))
	_, _, _, err := svc.Login(context.Background(), "nobody@gigmann.health", "pw", "")
	assert.ErrorIs(t, err, app.ErrInvalidCredentials)
}

func TestLoginWrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := mocks.NewMockUserRepository(ctrl)
	hasher := mocks.NewMockPasswordHasher(ctrl)
	users.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(execAccount(t), nil)
	hasher.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(false, nil)

	svc := newTestAuthService(users, hasher, mocks.NewMockTokenService(ctrl), mocks.NewMockRefreshTokenStore(ctrl), time.Hour, auditMock(ctrl))
	_, _, _, err := svc.Login(context.Background(), "ceo@gigmann.health", "bad", "")
	assert.ErrorIs(t, err, app.ErrInvalidCredentials)
}

func TestRefreshSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokens := mocks.NewMockTokenService(ctrl)
	refresh := mocks.NewMockRefreshTokenStore(ctrl)
	users := mocks.NewMockUserRepository(ctrl)

	// The rotated token carries a stale snapshot; the live account is re-read.
	stale := auth.Principal{UserID: "u1", Name: "stale", Role: user.RoleExecutive}
	fresh := auth.Principal{UserID: "u1", Name: "Sammy Adjei", Role: user.RoleExecutive}
	refresh.EXPECT().Consume(gomock.Any(), "old-refresh").Return(stale, nil)
	users.EXPECT().FindByID(gomock.Any(), "u1").Return(execAccount(t), nil)
	tokens.EXPECT().Issue(fresh).Return("new-access", nil)
	refresh.EXPECT().Issue(gomock.Any(), fresh, gomock.Any()).Return("new-refresh", nil)

	svc := newTestAuthService(users, mocks.NewMockPasswordHasher(ctrl), tokens, refresh, time.Hour, auditMock(ctrl))
	access, ref, got, err := svc.Refresh(context.Background(), "old-refresh")
	require.NoError(t, err)
	assert.Equal(t, "new-access", access)
	assert.Equal(t, "new-refresh", ref)
	assert.Equal(t, "Sammy Adjei", got.Name) // reflects the live account, not the rotated snapshot
}

func TestRefreshReflectsPrivilegeChange(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokens := mocks.NewMockTokenService(ctrl)
	refresh := mocks.NewMockRefreshTokenStore(ctrl)
	users := mocks.NewMockUserRepository(ctrl)

	// Token minted while the user was an executive...
	refresh.EXPECT().Consume(gomock.Any(), gomock.Any()).
		Return(auth.Principal{UserID: "u2", Role: user.RoleExecutive}, nil)
	// ...but the account is now a facility manager scoped to kasoa.
	mgr, err := user.New(user.User{ID: "u2", Name: "Ama", Role: user.RoleFacilityManager, FacilityID: "kasoa"})
	require.NoError(t, err)
	users.EXPECT().FindByID(gomock.Any(), "u2").Return(ports.Account{User: mgr}, nil)

	var issued auth.Principal
	tokens.EXPECT().Issue(gomock.Any()).
		DoAndReturn(func(pr auth.Principal) (string, error) { issued = pr; return "a", nil })
	refresh.EXPECT().Issue(gomock.Any(), gomock.Any(), gomock.Any()).Return("r", nil)

	svc := newTestAuthService(users, mocks.NewMockPasswordHasher(ctrl), tokens, refresh, time.Hour, auditMock(ctrl))
	_, _, got, err := svc.Refresh(context.Background(), "tok")
	require.NoError(t, err)
	assert.Equal(t, user.RoleFacilityManager, got.Role)
	assert.Equal(t, "kasoa", got.FacilityID)
	assert.Equal(t, user.RoleFacilityManager, issued.Role) // the new access token carries the downgraded role
}

func TestRefreshAccountGone(t *testing.T) {
	ctrl := gomock.NewController(t)
	refresh := mocks.NewMockRefreshTokenStore(ctrl)
	users := mocks.NewMockUserRepository(ctrl)
	refresh.EXPECT().Consume(gomock.Any(), gomock.Any()).Return(auth.Principal{UserID: "u1"}, nil)
	users.EXPECT().FindByID(gomock.Any(), "u1").Return(ports.Account{}, ports.ErrAccountNotFound)

	svc := newTestAuthService(users, mocks.NewMockPasswordHasher(ctrl), mocks.NewMockTokenService(ctrl), refresh, time.Hour, auditMock(ctrl))
	_, _, _, err := svc.Refresh(context.Background(), "tok")
	assert.ErrorIs(t, err, app.ErrInvalidCredentials)
}

func TestRefreshInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	refresh := mocks.NewMockRefreshTokenStore(ctrl)
	refresh.EXPECT().Consume(gomock.Any(), gomock.Any()).Return(auth.Principal{}, ports.ErrInvalidRefreshToken)

	svc := newTestAuthService(mocks.NewMockUserRepository(ctrl), mocks.NewMockPasswordHasher(ctrl), mocks.NewMockTokenService(ctrl), refresh, time.Hour, auditMock(ctrl))
	_, _, _, err := svc.Refresh(context.Background(), "bad")
	assert.ErrorIs(t, err, app.ErrInvalidCredentials)
}

func TestLogoutRevokes(t *testing.T) {
	ctrl := gomock.NewController(t)
	refresh := mocks.NewMockRefreshTokenStore(ctrl)
	refresh.EXPECT().Revoke(gomock.Any(), "some-refresh").Return(nil)

	svc := newTestAuthService(mocks.NewMockUserRepository(ctrl), mocks.NewMockPasswordHasher(ctrl), mocks.NewMockTokenService(ctrl), refresh, time.Hour, auditMock(ctrl))
	require.NoError(t, svc.Logout(context.Background(), "some-refresh"))
}

func TestPasswordResetFlowChangesPasswordAndConsumesToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokens := mocks.NewMockTokenService(ctrl)
	refresh := mocks.NewMockRefreshTokenStore(ctrl)

	acct := execAccount(t)
	acct.PasswordHash = "hash:old-password"
	users := memory.NewUserRepo(acct)

	svc := newTestAuthService(users, staticHasher{}, tokens, refresh, time.Hour, auditMock(ctrl))
	resetToken, err := svc.RequestPasswordReset(context.Background(), "ceo@gigmann.health")
	require.NoError(t, err)
	require.NotEmpty(t, resetToken)

	require.ErrorIs(t, svc.ConfirmPasswordReset(context.Background(), resetToken, "short"), app.ErrWeakPassword)
	require.ErrorIs(t, svc.ConfirmPasswordReset(context.Background(), resetToken, "password12345"), app.ErrWeakPassword)
	require.ErrorIs(t, svc.ConfirmPasswordReset(context.Background(), resetToken, "aaaaaaaaaaaa"), app.ErrWeakPassword)
	refresh.EXPECT().RevokeUser(gomock.Any(), "u1").Return(nil)
	require.NoError(t, svc.ConfirmPasswordReset(context.Background(), resetToken, "new-password"))

	saved, err := users.FindByID(context.Background(), "u1")
	require.NoError(t, err)
	assert.Equal(t, "hash:new-password", saved.PasswordHash)
	require.ErrorIs(t, svc.ConfirmPasswordReset(context.Background(), resetToken, "other-password"), app.ErrInvalidPasswordReset)

	tokens.EXPECT().Issue(gomock.Any()).Return("access-token", nil)
	refresh.EXPECT().Issue(gomock.Any(), gomock.Any(), gomock.Any()).Return("refresh-token", nil)
	_, _, _, err = svc.Login(context.Background(), "ceo@gigmann.health", "new-password", "")
	require.NoError(t, err)

	_, _, _, err = svc.Login(context.Background(), "ceo@gigmann.health", "old-password", "")
	require.ErrorIs(t, err, app.ErrInvalidCredentials)
}

func TestPasswordResetUnknownEmailIsAcceptedWithoutToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := mocks.NewMockUserRepository(ctrl)
	users.EXPECT().FindByEmail(gomock.Any(), "ghost@gigmann.health").Return(ports.Account{}, ports.ErrAccountNotFound)

	svc := newTestAuthService(users, staticHasher{}, mocks.NewMockTokenService(ctrl), mocks.NewMockRefreshTokenStore(ctrl), time.Hour, auditMock(ctrl))
	resetToken, err := svc.RequestPasswordReset(context.Background(), "ghost@gigmann.health")
	require.NoError(t, err)
	assert.Empty(t, resetToken)
}

func TestLoginTokenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := mocks.NewMockUserRepository(ctrl)
	hasher := mocks.NewMockPasswordHasher(ctrl)
	tokens := mocks.NewMockTokenService(ctrl)

	users.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(execAccount(t), nil)
	hasher.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(true, nil)
	tokens.EXPECT().Issue(gomock.Any()).Return("", errors.New("kms down"))

	svc := newTestAuthService(users, hasher, tokens, mocks.NewMockRefreshTokenStore(ctrl), time.Hour, auditMock(ctrl))
	_, _, _, err := svc.Login(context.Background(), "ceo@gigmann.health", "pw", "")
	require.Error(t, err)
	assert.NotErrorIs(t, err, app.ErrInvalidCredentials)
}

func auditMock(ctrl *gomock.Controller) *mocks.MockAuditLogger {
	m := mocks.NewMockAuditLogger(ctrl)
	m.EXPECT().Record(gomock.Any(), gomock.Any()).AnyTimes()
	return m
}

type staticHasher struct{}

func (staticHasher) Hash(plain string) (string, error) { return "hash:" + plain, nil }
func (staticHasher) Verify(plain, encoded string) (bool, error) {
	return encoded == "hash:"+plain, nil
}

func TestLoginAuditsSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := mocks.NewMockUserRepository(ctrl)
	hasher := mocks.NewMockPasswordHasher(ctrl)
	tokens := mocks.NewMockTokenService(ctrl)
	refresh := mocks.NewMockRefreshTokenStore(ctrl)
	users.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(execAccount(t), nil)
	hasher.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(true, nil)
	tokens.EXPECT().Issue(gomock.Any()).Return("a", nil)
	refresh.EXPECT().Issue(gomock.Any(), gomock.Any(), gomock.Any()).Return("r", nil)
	auditL := mocks.NewMockAuditLogger(ctrl)
	auditL.EXPECT().Record(gomock.Any(), ports.AuditEvent{Actor: "u1", Action: "auth.login", Outcome: "success"})

	svc := newTestAuthService(users, hasher, tokens, refresh, time.Hour, auditL)
	_, _, _, err := svc.Login(context.Background(), "ceo@gigmann.health", "pw", "")
	require.NoError(t, err)
}

func TestLoginAuditsFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := mocks.NewMockUserRepository(ctrl)
	users.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(ports.Account{}, ports.ErrAccountNotFound)
	auditL := mocks.NewMockAuditLogger(ctrl)
	auditL.EXPECT().Record(gomock.Any(), ports.AuditEvent{Actor: "ghost@x.io", Action: "auth.login", Outcome: "failure"})

	svc := newTestAuthService(users, mocks.NewMockPasswordHasher(ctrl), mocks.NewMockTokenService(ctrl), mocks.NewMockRefreshTokenStore(ctrl), time.Hour, auditL)
	_, _, _, err := svc.Login(context.Background(), "ghost@x.io", "pw", "")
	require.ErrorIs(t, err, app.ErrInvalidCredentials)
}

func TestLoginRequiresMFAWhenEnrolled(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := mocks.NewMockUserRepository(ctrl)
	hasher := mocks.NewMockPasswordHasher(ctrl)
	acct := execAccount(t)
	acct.MFASecret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	users.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(acct, nil)
	// A non-TOTP code falls through to the recovery-code path, which reloads the
	// account under the lock before checking (no codes here → still mfa_required).
	users.EXPECT().FindByID(gomock.Any(), "u1").Return(acct, nil)
	hasher.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(true, nil)

	svc := newTestAuthService(users, hasher, mocks.NewMockTokenService(ctrl), mocks.NewMockRefreshTokenStore(ctrl), time.Hour, auditMock(ctrl))
	_, _, _, err := svc.Login(context.Background(), "ceo@gigmann.health", "pw", "000000")
	assert.ErrorIs(t, err, app.ErrMFARequired)
}

func TestConfirmMFAEnrollment(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := mocks.NewMockUserRepository(ctrl)
	refresh := mocks.NewMockRefreshTokenStore(ctrl)
	users.EXPECT().FindByID(gomock.Any(), "u1").Return(execAccount(t), nil)
	users.EXPECT().Save(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, acct ports.Account) error {
			assert.NotEmpty(t, acct.MFASecret)
			require.Len(t, acct.RecoveryCodeHashes, 10)
			for _, hash := range acct.RecoveryCodeHashes {
				assert.True(t, strings.HasPrefix(hash, "hash:"))
			}
			return nil
		})
	refresh.EXPECT().RevokeUser(gomock.Any(), "u1").Return(nil)

	svc := newTestAuthService(users, staticHasher{}, mocks.NewMockTokenService(ctrl), refresh, time.Hour, auditMock(ctrl))
	secret, uri, err := svc.BeginMFAEnrollment(context.Background(), auth.Principal{UserID: "u1", Name: "Sammy"})
	require.NoError(t, err)
	require.NotEmpty(t, secret)
	assert.Contains(t, uri, "otpauth://")

	code, err := mfa.Code(secret, time.Now())
	require.NoError(t, err)
	codes, err := svc.ConfirmMFAEnrollment(context.Background(), auth.Principal{UserID: "u1", Name: "Sammy"}, secret, code)
	require.NoError(t, err)
	require.Len(t, codes, 10)
	assert.Contains(t, codes[0], "-")
}

func TestConfirmMFAEnrollmentBadCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := newTestAuthService(mocks.NewMockUserRepository(ctrl), mocks.NewMockPasswordHasher(ctrl), mocks.NewMockTokenService(ctrl), mocks.NewMockRefreshTokenStore(ctrl), time.Hour, auditMock(ctrl))
	_, err := svc.ConfirmMFAEnrollment(context.Background(), auth.Principal{UserID: "u1"}, "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", "000000")
	assert.ErrorIs(t, err, app.ErrInvalidMFACode)
}

func TestCurrentUserReportsMFAEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	acct := execAccount(t)
	acct.MFASecret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	users := memory.NewUserRepo(acct)

	svc := newTestAuthService(users, staticHasher{}, mocks.NewMockTokenService(ctrl), mocks.NewMockRefreshTokenStore(ctrl), time.Hour, auditMock(ctrl))
	p, enabled, err := svc.CurrentUser(context.Background(), auth.Principal{UserID: "u1"})

	require.NoError(t, err)
	assert.Equal(t, "Sammy Adjei", p.Name)
	assert.True(t, enabled)
}

func TestLoginWithRecoveryCodeConsumesIt(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokens := mocks.NewMockTokenService(ctrl)
	refresh := mocks.NewMockRefreshTokenStore(ctrl)

	acct := execAccount(t)
	acct.PasswordHash = "hash:pw"
	acct.MFASecret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	acct.RecoveryCodeHashes = []string{"hash:ABCD1234EFGH5678"}
	users := memory.NewUserRepo(acct)

	tokens.EXPECT().Issue(gomock.Any()).Return("access-token", nil)
	refresh.EXPECT().Issue(gomock.Any(), gomock.Any(), gomock.Any()).Return("refresh-token", nil)

	svc := newTestAuthService(users, staticHasher{}, tokens, refresh, time.Hour, auditMock(ctrl))
	_, _, _, err := svc.Login(context.Background(), "ceo@gigmann.health", "pw", "abcd-1234-efgh-5678")
	require.NoError(t, err)

	saved, err := users.FindByID(context.Background(), "u1")
	require.NoError(t, err)
	require.Empty(t, saved.RecoveryCodeHashes)

	_, _, _, err = svc.Login(context.Background(), "ceo@gigmann.health", "pw", "abcd-1234-efgh-5678")
	require.ErrorIs(t, err, app.ErrMFARequired)
}

// TestLoginRecoveryCodeNoDoubleSpend guards the TOCTOU fix: two concurrent
// logins presenting the same recovery code must spend it at most once.
func TestLoginRecoveryCodeNoDoubleSpend(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokens := mocks.NewMockTokenService(ctrl)
	refresh := mocks.NewMockRefreshTokenStore(ctrl)

	acct := execAccount(t)
	acct.PasswordHash = "hash:pw"
	acct.MFASecret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	acct.RecoveryCodeHashes = []string{"hash:ABCD1234EFGH5678"}
	users := memory.NewUserRepo(acct)

	// Exactly one login may succeed (and issue a token pair).
	tokens.EXPECT().Issue(gomock.Any()).Return("access-token", nil).Times(1)
	refresh.EXPECT().Issue(gomock.Any(), gomock.Any(), gomock.Any()).Return("refresh-token", nil).Times(1)

	svc := newTestAuthService(users, staticHasher{}, tokens, refresh, time.Hour, auditMock(ctrl))

	var wg sync.WaitGroup
	var successes atomic.Int64
	for range 2 {
		wg.Go(func() {
			if _, _, _, err := svc.Login(context.Background(), "ceo@gigmann.health", "pw", "abcd-1234-efgh-5678"); err == nil {
				successes.Add(1)
			}
		})
	}
	wg.Wait()
	assert.Equal(t, int64(1), successes.Load(), "a recovery code must be spendable at most once, even under concurrent logins")
}

func TestDisableMFAWithRecoveryCodeClearsSecretAndCodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	acct := execAccount(t)
	acct.MFASecret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	acct.RecoveryCodeHashes = []string{"hash:ABCD1234EFGH5678"}
	users := memory.NewUserRepo(acct)
	refresh := mocks.NewMockRefreshTokenStore(ctrl)
	refresh.EXPECT().RevokeUser(gomock.Any(), "u1").Return(nil)

	svc := newTestAuthService(users, staticHasher{}, mocks.NewMockTokenService(ctrl), refresh, time.Hour, auditMock(ctrl))
	err := svc.DisableMFA(context.Background(), auth.Principal{UserID: "u1"}, "abcd-1234-efgh-5678")

	require.NoError(t, err)
	saved, err := users.FindByID(context.Background(), "u1")
	require.NoError(t, err)
	assert.Empty(t, saved.MFASecret)
	assert.Empty(t, saved.RecoveryCodeHashes)
}

func TestDisableMFABadCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	acct := execAccount(t)
	acct.MFASecret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	acct.RecoveryCodeHashes = []string{"hash:ABCD1234EFGH5678"}
	users := memory.NewUserRepo(acct)

	svc := newTestAuthService(users, staticHasher{}, mocks.NewMockTokenService(ctrl), mocks.NewMockRefreshTokenStore(ctrl), time.Hour, auditMock(ctrl))
	err := svc.DisableMFA(context.Background(), auth.Principal{UserID: "u1"}, "000000")

	require.ErrorIs(t, err, app.ErrInvalidMFACode)
	saved, err := users.FindByID(context.Background(), "u1")
	require.NoError(t, err)
	assert.NotEmpty(t, saved.MFASecret)
	assert.NotEmpty(t, saved.RecoveryCodeHashes)
}
