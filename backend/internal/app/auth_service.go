package app

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/mfa"
	"github.com/xcreativs/gigmann/internal/ports"
)

// ErrInvalidCredentials is returned for any failed login or refresh — the same
// error regardless of cause, so callers cannot enumerate accounts or tokens.
var ErrInvalidCredentials = errors.New("app: invalid credentials")

// ErrMFARequired means the password was correct but a valid TOTP code is needed.
var ErrMFARequired = errors.New("app: mfa code required")

// ErrInvalidMFACode means an enrollment confirmation code did not validate.
var ErrInvalidMFACode = errors.New("app: invalid mfa code")

const (
	recoveryCodeCount = 10
	recoveryCodeBytes = 10
)

// AuthService is the authentication use case: verify credentials, issue an
// access token plus a rotating refresh token, refresh, and revoke (logout).
type AuthService struct {
	users      ports.UserRepository
	hasher     ports.PasswordHasher
	tokens     ports.TokenService
	refresh    ports.RefreshTokenStore
	refreshTTL time.Duration
	audit      ports.AuditLogger

	mfaMu   sync.Mutex
	mfaUsed map[string]uint64 // userID -> last consumed TOTP step (single-use / anti-replay)

	// recoveryMu serialises recovery-code consumption so the verify→remove→save is
	// atomic on this instance — without it two concurrent logins with the same code
	// both verify before either saves and the code is spent twice (MFA bypass). A
	// multi-instance deployment additionally needs a row-locked UPDATE; see
	// docs/security/audit-findings-2026-06-29.md.
	recoveryMu sync.Mutex
}

// NewAuthService wires the authentication use case to its ports.
func NewAuthService(
	users ports.UserRepository,
	hasher ports.PasswordHasher,
	tokens ports.TokenService,
	refresh ports.RefreshTokenStore,
	refreshTTL time.Duration,
	audit ports.AuditLogger,
) *AuthService {
	return &AuthService{
		users: users, hasher: hasher, tokens: tokens, refresh: refresh,
		refreshTTL: refreshTTL, audit: audit, mfaUsed: map[string]uint64{},
	}
}

// Login verifies the email/password and returns an access token, a refresh
// token, and the principal.
func (s *AuthService) Login(ctx context.Context, email, password, code string) (string, string, auth.Principal, error) {
	acct, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		s.audit.Record(ctx, ports.AuditEvent{Actor: email, Action: "auth.login", Outcome: "failure"})
		return "", "", auth.Principal{}, ErrInvalidCredentials
	}
	ok, err := s.hasher.Verify(password, acct.PasswordHash)
	if err != nil || !ok {
		s.audit.Record(ctx, ports.AuditEvent{Actor: email, Action: "auth.login", Outcome: "failure"})
		return "", "", auth.Principal{}, ErrInvalidCredentials
	}
	if acct.MFASecret != "" {
		ok, err := s.checkSecondFactor(ctx, acct, code, time.Now())
		if err != nil {
			return "", "", auth.Principal{}, err
		}
		if !ok {
			s.audit.Record(ctx, ports.AuditEvent{Actor: acct.User.ID, Action: "auth.login", Outcome: "mfa_required"})
			return "", "", auth.Principal{}, ErrMFARequired
		}
	}
	p := principalOf(acct)
	s.audit.Record(ctx, ports.AuditEvent{Actor: p.UserID, Action: "auth.login", Outcome: "success"})
	return s.issue(ctx, p)
}

// Refresh rotates a valid refresh token into a fresh access + refresh token pair.
// It re-reads the live account so a role/facility change — or a deleted account —
// takes effect on the next refresh (within the access-token TTL) rather than
// persisting for the whole refresh-token lifetime. The principal is rebuilt from
// current account data, never trusted from the rotated token's snapshot.
func (s *AuthService) Refresh(ctx context.Context, rawRefresh string) (string, string, auth.Principal, error) {
	p, err := s.refresh.Consume(ctx, rawRefresh)
	if err != nil {
		return "", "", auth.Principal{}, ErrInvalidCredentials
	}
	acct, err := s.users.FindByID(ctx, p.UserID)
	if err != nil {
		s.audit.Record(ctx, ports.AuditEvent{Actor: p.UserID, Action: "auth.refresh", Outcome: "failure"})
		return "", "", auth.Principal{}, ErrInvalidCredentials
	}
	return s.issue(ctx, principalOf(acct))
}

// principalOf builds an auth principal from the live account.
func principalOf(acct ports.Account) auth.Principal {
	return auth.Principal{
		UserID:     acct.User.ID,
		Name:       acct.User.Name,
		Role:       acct.User.Role,
		FacilityID: acct.User.FacilityID,
	}
}

// Logout revokes a refresh token so it can no longer be rotated.
func (s *AuthService) Logout(ctx context.Context, rawRefresh string) error {
	s.audit.Record(ctx, ports.AuditEvent{Action: "auth.logout", Outcome: "success"})
	return s.refresh.Revoke(ctx, rawRefresh)
}

// CurrentUser re-reads the authenticated account so account state changes (such
// as enabling or disabling MFA) are reflected in /auth/me without waiting for a
// token refresh.
func (s *AuthService) CurrentUser(ctx context.Context, p auth.Principal) (auth.Principal, bool, error) {
	acct, err := s.users.FindByID(ctx, p.UserID)
	if err != nil {
		s.audit.Record(ctx, ports.AuditEvent{Actor: p.UserID, Action: "auth.me", Outcome: "failure"})
		return auth.Principal{}, false, ErrInvalidCredentials
	}
	return principalOf(acct), acct.MFASecret != "", nil
}

// BeginMFAEnrollment mints a fresh TOTP secret and its otpauth URI. The secret
// is not persisted until ConfirmMFAEnrollment proves the user can generate codes.
func (s *AuthService) BeginMFAEnrollment(_ context.Context, p auth.Principal) (string, string, error) {
	secret, err := mfa.NewSecret()
	if err != nil {
		return "", "", fmt.Errorf("app: begin mfa enrollment: %w", err)
	}
	// Reset the single-use step counter for a fresh enrollment, so re-enrolling
	// (even within the same TOTP window) isn't rejected as a replay.
	s.mfaMu.Lock()
	delete(s.mfaUsed, p.UserID)
	s.mfaMu.Unlock()
	return secret, mfa.OTPAuthURI(secret, p.Name, "Gigmann"), nil
}

// ConfirmMFAEnrollment validates a code against the secret and, on success,
// persists the secret on the principal's account (activating MFA). It returns
// one-time recovery codes; callers must show them once and never persist them raw.
func (s *AuthService) ConfirmMFAEnrollment(ctx context.Context, p auth.Principal, secret, code string) ([]string, error) {
	if !s.checkMFA(p.UserID, secret, code, time.Now()) {
		return nil, ErrInvalidMFACode
	}
	acct, err := s.users.FindByID(ctx, p.UserID)
	if err != nil {
		return nil, fmt.Errorf("app: load account: %w", err)
	}
	codes, hashes, err := s.generateRecoveryCodes()
	if err != nil {
		return nil, err
	}
	acct.MFASecret = secret
	acct.RecoveryCodeHashes = hashes
	if err := s.users.Save(ctx, acct); err != nil {
		return nil, fmt.Errorf("app: save account: %w", err)
	}
	s.audit.Record(ctx, ports.AuditEvent{Actor: p.UserID, Action: "auth.mfa.enroll", Outcome: "success"})
	return codes, nil
}

// DisableMFA clears the persisted TOTP secret and recovery-code hashes after the
// user proves possession with a current TOTP or an unused recovery code.
func (s *AuthService) DisableMFA(ctx context.Context, p auth.Principal, code string) error {
	acct, err := s.users.FindByID(ctx, p.UserID)
	if err != nil {
		return fmt.Errorf("app: load account: %w", err)
	}
	if acct.MFASecret == "" {
		return nil
	}
	ok, err := s.checkSecondFactor(ctx, acct, code, time.Now())
	if err != nil {
		return err
	}
	if !ok {
		s.audit.Record(ctx, ports.AuditEvent{Actor: p.UserID, Action: "auth.mfa.disable", Outcome: "failure"})
		return ErrInvalidMFACode
	}
	acct.MFASecret = ""
	acct.RecoveryCodeHashes = nil
	if err := s.users.Save(ctx, acct); err != nil {
		return fmt.Errorf("app: disable mfa: %w", err)
	}
	s.mfaMu.Lock()
	delete(s.mfaUsed, p.UserID)
	s.mfaMu.Unlock()
	s.audit.Record(ctx, ports.AuditEvent{Actor: p.UserID, Action: "auth.mfa.disable", Outcome: "success"})
	return nil
}

func (s *AuthService) checkSecondFactor(ctx context.Context, acct ports.Account, code string, now time.Time) (bool, error) {
	if s.checkMFA(acct.User.ID, acct.MFASecret, code, now) {
		return true, nil
	}
	if code == "" {
		return false, nil
	}
	return s.consumeRecoveryCode(ctx, acct, code)
}

func (s *AuthService) consumeRecoveryCode(ctx context.Context, acct ports.Account, code string) (bool, error) {
	normalized := normalizeRecoveryCode(code)
	if normalized == "" {
		return false, nil
	}
	// Serialise and reload under the lock so verify→remove→save is atomic: the
	// passed acct may be stale (loaded before another concurrent login consumed a
	// code), so always work from the freshly persisted state.
	s.recoveryMu.Lock()
	defer s.recoveryMu.Unlock()
	fresh, err := s.users.FindByID(ctx, acct.User.ID)
	if err != nil {
		return false, fmt.Errorf("app: reload account: %w", err)
	}
	for i, hash := range fresh.RecoveryCodeHashes {
		ok, err := s.hasher.Verify(normalized, hash)
		if err != nil {
			return false, fmt.Errorf("app: verify recovery code: %w", err)
		}
		if !ok {
			continue
		}
		fresh.RecoveryCodeHashes = append(fresh.RecoveryCodeHashes[:i], fresh.RecoveryCodeHashes[i+1:]...)
		if err := s.users.Save(ctx, fresh); err != nil {
			return false, fmt.Errorf("app: consume recovery code: %w", err)
		}
		s.audit.Record(ctx, ports.AuditEvent{Actor: fresh.User.ID, Action: "auth.mfa.recovery", Outcome: "success"})
		return true, nil
	}
	return false, nil
}

func (s *AuthService) generateRecoveryCodes() ([]string, []string, error) {
	codes := make([]string, 0, recoveryCodeCount)
	hashes := make([]string, 0, recoveryCodeCount)
	for range recoveryCodeCount {
		code, normalized, err := newRecoveryCode()
		if err != nil {
			return nil, nil, err
		}
		hash, err := s.hasher.Hash(normalized)
		if err != nil {
			return nil, nil, fmt.Errorf("app: hash recovery code: %w", err)
		}
		codes = append(codes, code)
		hashes = append(hashes, hash)
	}
	return codes, hashes, nil
}

func newRecoveryCode() (formatted string, normalized string, err error) {
	var b [recoveryCodeBytes]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", "", fmt.Errorf("app: generate recovery code: %w", err)
	}
	normalized = strings.TrimRight(base32.StdEncoding.EncodeToString(b[:]), "=")
	return normalized[:4] + "-" + normalized[4:8] + "-" + normalized[8:12] + "-" + normalized[12:16], normalized, nil
}

func normalizeRecoveryCode(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	code = strings.ReplaceAll(code, "-", "")
	code = strings.ReplaceAll(code, " ", "")
	return code
}

// checkMFA validates a TOTP code and enforces single-use: a given time-step
// counter is accepted at most once per user, so a captured code cannot be
// replayed within its ±1-step validity window.
func (s *AuthService) checkMFA(userID, secret, code string, now time.Time) bool {
	step, ok := mfa.ValidateAt(secret, code, now)
	if !ok {
		return false
	}
	s.mfaMu.Lock()
	defer s.mfaMu.Unlock()
	if step <= s.mfaUsed[userID] {
		return false // this code (or a newer one) has already been consumed — replay
	}
	s.mfaUsed[userID] = step
	return true
}

func (s *AuthService) issue(ctx context.Context, p auth.Principal) (string, string, auth.Principal, error) {
	access, err := s.tokens.Issue(p)
	if err != nil {
		return "", "", auth.Principal{}, fmt.Errorf("app: issue access token: %w", err)
	}
	refresh, err := s.refresh.Issue(ctx, p, s.refreshTTL)
	if err != nil {
		return "", "", auth.Principal{}, fmt.Errorf("app: issue refresh token: %w", err)
	}
	return access, refresh, p, nil
}
