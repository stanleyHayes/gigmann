package app

import (
	"context"
	"errors"
	"fmt"
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

// AuthService is the authentication use case: verify credentials, issue an
// access token plus a rotating refresh token, refresh, and revoke (logout).
type AuthService struct {
	users      ports.UserRepository
	hasher     ports.PasswordHasher
	tokens     ports.TokenService
	refresh    ports.RefreshTokenStore
	refreshTTL time.Duration
	audit      ports.AuditLogger
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
	return &AuthService{users: users, hasher: hasher, tokens: tokens, refresh: refresh, refreshTTL: refreshTTL, audit: audit}
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
	if acct.MFASecret != "" && !mfa.Validate(acct.MFASecret, code, time.Now()) {
		s.audit.Record(ctx, ports.AuditEvent{Actor: acct.User.ID, Action: "auth.login", Outcome: "mfa_required"})
		return "", "", auth.Principal{}, ErrMFARequired
	}
	p := auth.Principal{
		UserID:     acct.User.ID,
		Name:       acct.User.Name,
		Role:       acct.User.Role,
		FacilityID: acct.User.FacilityID,
	}
	s.audit.Record(ctx, ports.AuditEvent{Actor: p.UserID, Action: "auth.login", Outcome: "success"})
	return s.issue(ctx, p)
}

// Refresh rotates a valid refresh token into a fresh access + refresh token pair.
func (s *AuthService) Refresh(ctx context.Context, rawRefresh string) (string, string, auth.Principal, error) {
	p, err := s.refresh.Consume(ctx, rawRefresh)
	if err != nil {
		return "", "", auth.Principal{}, ErrInvalidCredentials
	}
	return s.issue(ctx, p)
}

// Logout revokes a refresh token so it can no longer be rotated.
func (s *AuthService) Logout(ctx context.Context, rawRefresh string) error {
	s.audit.Record(ctx, ports.AuditEvent{Action: "auth.logout", Outcome: "success"})
	return s.refresh.Revoke(ctx, rawRefresh)
}

// BeginMFAEnrollment mints a fresh TOTP secret and its otpauth URI. The secret
// is not persisted until ConfirmMFAEnrollment proves the user can generate codes.
func (s *AuthService) BeginMFAEnrollment(_ context.Context, p auth.Principal) (string, string, error) {
	secret, err := mfa.NewSecret()
	if err != nil {
		return "", "", fmt.Errorf("app: begin mfa enrollment: %w", err)
	}
	return secret, mfa.OTPAuthURI(secret, p.Name, "Gigmann"), nil
}

// ConfirmMFAEnrollment validates a code against the secret and, on success,
// persists the secret on the principal's account (activating MFA).
func (s *AuthService) ConfirmMFAEnrollment(ctx context.Context, p auth.Principal, secret, code string) error {
	if !mfa.Validate(secret, code, time.Now()) {
		return ErrInvalidMFACode
	}
	acct, err := s.users.FindByID(ctx, p.UserID)
	if err != nil {
		return fmt.Errorf("app: load account: %w", err)
	}
	acct.MFASecret = secret
	if err := s.users.Save(ctx, acct); err != nil {
		return fmt.Errorf("app: save account: %w", err)
	}
	s.audit.Record(ctx, ports.AuditEvent{Actor: p.UserID, Action: "auth.mfa.enroll", Outcome: "success"})
	return nil
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
