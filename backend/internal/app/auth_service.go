package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/ports"
)

// ErrInvalidCredentials is returned for any failed login (unknown email or wrong
// password) — the same error for both, so callers cannot enumerate accounts.
var ErrInvalidCredentials = errors.New("app: invalid credentials")

// AuthService is the authentication use case: verify credentials, issue a token.
type AuthService struct {
	users  ports.UserRepository
	hasher ports.PasswordHasher
	tokens ports.TokenService
}

// NewAuthService wires the authentication use case to its ports.
func NewAuthService(users ports.UserRepository, hasher ports.PasswordHasher, tokens ports.TokenService) *AuthService {
	return &AuthService{users: users, hasher: hasher, tokens: tokens}
}

// Login verifies the email/password and returns a signed token and principal.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, auth.Principal, error) {
	acct, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return "", auth.Principal{}, ErrInvalidCredentials
	}
	ok, err := s.hasher.Verify(password, acct.PasswordHash)
	if err != nil || !ok {
		return "", auth.Principal{}, ErrInvalidCredentials
	}
	p := auth.Principal{
		UserID:     acct.User.ID,
		Name:       acct.User.Name,
		Role:       acct.User.Role,
		FacilityID: acct.User.FacilityID,
	}
	tok, err := s.tokens.Issue(p)
	if err != nil {
		return "", auth.Principal{}, fmt.Errorf("app: issue token: %w", err)
	}
	return tok, p, nil
}
