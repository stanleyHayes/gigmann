// Package token implements ports.TokenService with signed HS256 JWT access
// tokens carrying the authenticated principal.
package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/ports"
)

const issuer = "gigmann"

// ErrInvalidToken is returned when a token is missing, expired, or tampered with.
var ErrInvalidToken = errors.New("token: invalid token")

type claims struct {
	jwt.RegisteredClaims

	Name       string `json:"name"`
	Role       string `json:"role"`
	FacilityID string `json:"fid,omitempty"`
}

// Service issues and verifies HS256 access tokens.
type Service struct {
	secret []byte
	ttl    time.Duration
}

// New builds a token Service with the given signing secret and access-token TTL.
func New(secret []byte, ttl time.Duration) Service {
	return Service{secret: secret, ttl: ttl}
}

var _ ports.TokenService = Service{}

// Issue signs an access token for the principal.
func (s Service) Issue(p auth.Principal) (string, error) {
	now := time.Now()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Name:       p.Name,
		Role:       string(p.Role),
		FacilityID: p.FacilityID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   p.UserID,
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	})
	signed, err := tok.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("token: sign: %w", err)
	}
	return signed, nil
}

// Verify parses and validates a token, returning the principal it carries.
func (s Service) Verify(raw string) (auth.Principal, error) {
	parsed, err := jwt.ParseWithClaims(raw, &claims{}, func(*jwt.Token) (any, error) {
		return s.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(issuer), jwt.WithExpirationRequired())
	if err != nil {
		return auth.Principal{}, ErrInvalidToken
	}
	c, ok := parsed.Claims.(*claims)
	if !ok || !parsed.Valid {
		return auth.Principal{}, ErrInvalidToken
	}
	return auth.Principal{
		UserID:     c.Subject,
		Name:       c.Name,
		Role:       user.Role(c.Role),
		FacilityID: c.FacilityID,
	}, nil
}
