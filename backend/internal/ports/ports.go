// Package ports declares the interfaces the application layer depends on.
// Outbound adapters (Postgres, Redis, in-memory, Anthropic, crypto) implement these.
package ports

import (
	"context"
	"errors"
	"time"

	"github.com/xcreativs/gigmann/internal/core/alert"
	"github.com/xcreativs/gigmann/internal/core/approval"
	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/metric"
	"github.com/xcreativs/gigmann/internal/core/task"
	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/intel"
)

//go:generate go tool mockgen -destination=mocks/mocks.go -package=mocks github.com/xcreativs/gigmann/internal/ports FacilityRepository,Narrator,BriefGenerator,UserRepository,PasswordHasher,TokenService,RefreshTokenStore,PasswordResetTokenStore,ApprovalRepository,TaskRepository,AlertRepository,MetricsRepository,Embedder,FacilityEmbeddingRepository,Answerer,QuestionAnswerer,AuditLogger

// ErrAccountNotFound is returned by UserRepository when no account matches.
var ErrAccountNotFound = errors.New("ports: account not found")

// FacilityRepository is a driven port for reading/writing facilities.
type FacilityRepository interface {
	List(ctx context.Context) ([]facility.Facility, error)
}

// MetricsRepository is a driven port for reading the facility metric series.
// KPI figures are computed in Go (kpi.Compute) from this raw series — the store
// is never a source of numbers.
type MetricsRepository interface {
	ListNetwork(ctx context.Context) ([]metric.FacilityMetric, error)
}

// EmbedKind distinguishes corpus documents from search queries (maps to the
// provider's input_type; symmetric embedders may ignore it).
type EmbedKind string

const (
	EmbedDocument EmbedKind = "document"
	EmbedQuery    EmbedKind = "query"
)

// Embedder turns text into fixed-dimension vectors for similarity search.
// Implementations: a cloud provider (Voyage) and a deterministic local fallback.
type Embedder interface {
	Embed(ctx context.Context, texts []string, kind EmbedKind) ([][]float32, error)
	Dimensions() int
}

// FacilityMatch is a facility ranked by vector similarity to a query.
type FacilityMatch struct {
	FacilityID string
	Content    string
	Distance   float64 // cosine distance (0 = identical)
}

// FacilityEmbeddingRepository stores and ANN-searches facility text embeddings.
type FacilityEmbeddingRepository interface {
	Upsert(ctx context.Context, facilityID, content string, embedding []float32) error
	Count(ctx context.Context) (int, error)
	Search(ctx context.Context, embedding []float32, limit int) ([]FacilityMatch, error)
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
	User               user.User
	Email              string
	PasswordHash       string
	MFASecret          string   // base32 TOTP secret; empty means MFA is not enrolled
	RecoveryCodeHashes []string // one-time MFA recovery code hashes; raw codes are shown only once
}

// UserRepository is a driven port for looking up accounts by email.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (Account, error)
	FindByID(ctx context.Context, id string) (Account, error)
	Save(ctx context.Context, account Account) error
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
	RevokeUser(ctx context.Context, userID string) error
}

// ErrInvalidPasswordResetToken is returned for a missing, expired, or already-used reset token.
var ErrInvalidPasswordResetToken = errors.New("ports: invalid or expired password reset token")

// PasswordResetTokenStore issues and consumes short-lived, single-use password
// reset tokens. Implementations persist only a hash of the raw token.
type PasswordResetTokenStore interface {
	Issue(ctx context.Context, userID string, ttl time.Duration) (string, error)
	Consume(ctx context.Context, raw string) (string, error)
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

// ErrAlertNotFound is returned by AlertRepository when no alert matches.
var ErrAlertNotFound = errors.New("ports: alert not found")

// AlertRepository is a driven port for reading and updating alerts.
type AlertRepository interface {
	List(ctx context.Context) ([]alert.Alert, error)
	Get(ctx context.Context, id string) (alert.Alert, error)
	Save(ctx context.Context, a alert.Alert) error
}

// TaskRepository is a driven port for reading and updating "My Day" tasks.
type TaskRepository interface {
	List(ctx context.Context) ([]task.Task, error)
	Get(ctx context.Context, id string) (task.Task, error)
	Save(ctx context.Context, t task.Task) error
}

// Answerer answers a natural-language question grounded in a computed context.
// Implementations must use only the supplied figures and never invent numbers.
type Answerer interface {
	Answer(ctx context.Context, question string, c intel.Context) (intel.Answer, error)
}

// QuestionAnswerer is the inbound "Ask" use case over the current network.
type QuestionAnswerer interface {
	Answer(ctx context.Context, question string) (intel.Answer, error)
}

// AuditEvent is a security-relevant event recorded to the audit trail.
type AuditEvent struct {
	Actor   string // user id or attempted identity
	Action  string // e.g. "auth.login", "auth.logout", "approval.decide"
	Target  string // affected resource id (optional)
	Outcome string // "success" | "failure" | "approved" | "declined" | "forbidden"
}

// AuditLogger records security-relevant events (auth, decisions).
type AuditLogger interface {
	Record(ctx context.Context, e AuditEvent)
}

// Notifier broadcasts a named event to connected clients (realtime channel).
type Notifier interface {
	Notify(event string)
}

// PushSubscription is a browser Web Push subscription (W3C Push API / RFC 8030).
type PushSubscription struct {
	Endpoint string
	P256dh   string // client public key (base64url)
	Auth     string // client auth secret (base64url)
}

// PushSubscriptionStore persists Web Push subscriptions per user (driven port).
// Implementations dedupe by (userID, endpoint).
type PushSubscriptionStore interface {
	Save(ctx context.Context, userID string, sub PushSubscription) error
	Delete(ctx context.Context, userID, endpoint string) error
	ListByUser(ctx context.Context, userID string) ([]PushSubscription, error)
	All(ctx context.Context) (map[string][]PushSubscription, error)
}

// PushSender delivers an encrypted Web Push payload to a single subscription
// (driven port). When VAPID keys are not configured the sender is disabled and
// every Send is a no-op, so the feature degrades to off without keys.
type PushSender interface {
	Enabled() bool
	PublicKey() string
	Send(ctx context.Context, sub PushSubscription, payload []byte) error
}
