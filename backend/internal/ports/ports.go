// Package ports declares the interfaces the application layer depends on.
// Outbound adapters (Postgres, Redis, in-memory, Anthropic, crypto) implement these.
package ports

import (
	"context"
	"errors"
	"time"

	"github.com/xcreativs/gigmann/internal/core/approval"
	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/task"
	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/intel"
)

//go:generate go tool mockgen -destination=mocks/mocks.go -package=mocks github.com/xcreativs/gigmann/internal/ports FacilityRepository,Narrator,BriefGenerator,UserRepository,PasswordHasher,TokenService,RefreshTokenStore,ApprovalRepository,TaskRepository

// ErrAccountNotFound is returned by UserRepository when no account matches.
var ErrAccountNotFound = errors.New("ports: account not found")

// FacilityRepository is a driven port for reading/writing facilities.
type FacilityRepository interface {
	List(ctx context.Context) ([]facility.Facility, error)
}

// Narrator turns a computed brief context into a narrated Daily Brief.
// Implementations must narrate only the supplied figures and never fabricate numbers.
type Narrator interface {
	NarrateBrief(ctx context.Context, c intel.Context) (brief.Brief, error)
}

// BriefGenerator produces the current Daily Brief for the network (inbound use case).
type BriefGenerator interface {
	Generate(ctx context.Context) (brief.Brief, error)
}

// Account is a user profile plus the credentials used to authenticate them.
type Account struct {
	User         user.User
	Email        string
	PasswordHash string
}

// UserRepository is a driven port for looking up accounts by email.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (Account, error)
}

// PasswordHasher hashes and verifies passwords (argon2id in the adapter).
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(plain, encoded string) (bool, error)
}

// TokenService issues and verifies signed access tokens for a principal.
type TokenService interface {
	Issue(p auth.Principal) (string, error)
	Verify(token string) (auth.Principal, error)
}

// ErrInvalidRefreshToken is returned for a missing, expired, or already-used refresh token.
var ErrInvalidRefreshToken = errors.New("ports: invalid or expired refresh token")

// RefreshTokenStore issues, rotates (single-use), and revokes refresh tokens.
// Implementations persist only a hash of the raw token.
type RefreshTokenStore interface {
	Issue(ctx context.Context, p auth.Principal, ttl time.Duration) (string, error)
	Consume(ctx context.Context, raw string) (auth.Principal, error)
	Revoke(ctx context.Context, raw string) error
}

// ErrApprovalNotFound is returned by ApprovalRepository when no approval matches.
var ErrApprovalNotFound = errors.New("ports: approval not found")

// ApprovalRepository is a driven port for reading and updating approvals.
type ApprovalRepository interface {
	List(ctx context.Context) ([]approval.Approval, error)
	Get(ctx context.Context, id string) (approval.Approval, error)
	Save(ctx context.Context, a approval.Approval) error
}

// ErrTaskNotFound is returned by TaskRepository when no task matches.
var ErrTaskNotFound = errors.New("ports: task not found")

// TaskRepository is a driven port for reading and updating "My Day" tasks.
type TaskRepository interface {
	List(ctx context.Context) ([]task.Task, error)
	Get(ctx context.Context, id string) (task.Task, error)
	Save(ctx context.Context, t task.Task) error
}
