package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/postgres/sqlcgen"
	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/ports"
)

// UserRepo is a PostgreSQL implementation of ports.UserRepository. Profile and
// credential data live in separate tables (users, credentials); Save writes both
// atomically in a transaction.
type UserRepo struct {
	pool *pgxpool.Pool
	q    *sqlcgen.Queries
}

var _ ports.UserRepository = (*UserRepo)(nil)

// NewUserRepo builds a UserRepo over a pgx pool. The pool is required (not a bare
// DBTX) because Save spans two statements in a transaction.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool, q: sqlcgen.New(pool)}
}

// FindByEmail returns the account for the (normalised) email, or ErrAccountNotFound.
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (ports.Account, error) {
	row, err := r.q.FindAccountByEmail(ctx, normalizeEmail(email))
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.Account{}, ports.ErrAccountNotFound
	}
	if err != nil {
		return ports.Account{}, fmt.Errorf("postgres: find account by email: %w", err)
	}
	return accountFrom(row.ID, row.Name, row.Role, row.FacilityID, row.Preferences, row.Email, row.PasswordHash, row.MfaSecret, row.RecoveryCodeHashes)
}

// FindByID returns the account for the user id, or ErrAccountNotFound.
func (r *UserRepo) FindByID(ctx context.Context, id string) (ports.Account, error) {
	row, err := r.q.FindAccountByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.Account{}, ports.ErrAccountNotFound
	}
	if err != nil {
		return ports.Account{}, fmt.Errorf("postgres: find account by id: %w", err)
	}
	return accountFrom(row.ID, row.Name, row.Role, row.FacilityID, row.Preferences, row.Email, row.PasswordHash, row.MfaSecret, row.RecoveryCodeHashes)
}

// Save upserts the account's profile and credentials in a single transaction.
func (r *UserRepo) Save(ctx context.Context, account ports.Account) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op once committed

	if err := saveAccountTx(ctx, r.q.WithTx(tx), account); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres: commit save account: %w", err)
	}
	return nil
}

// saveAccountTx upserts an account's profile + credentials using the supplied
// (transactional) Queries. Callers provide the transaction so the two writes are
// atomic together — and so first-run seeding can fold them into a larger tx.
func saveAccountTx(ctx context.Context, q *sqlcgen.Queries, account ports.Account) error {
	prefs, err := marshalPrefs(account.User.Preferences)
	if err != nil {
		return err
	}
	if err := q.UpsertUser(ctx, userParams(account, prefs)); err != nil {
		return fmt.Errorf("postgres: upsert user: %w", err)
	}
	if err := q.UpsertCredentials(ctx, credentialParams(account)); err != nil {
		return fmt.Errorf("postgres: upsert credentials: %w", err)
	}
	return nil
}

func userParams(account ports.Account, prefs []byte) sqlcgen.UpsertUserParams {
	return sqlcgen.UpsertUserParams{
		ID:          account.User.ID,
		Name:        account.User.Name,
		Role:        string(account.User.Role),
		FacilityID:  nullableStr(account.User.FacilityID),
		Preferences: prefs,
	}
}

func credentialParams(account ports.Account) sqlcgen.UpsertCredentialsParams {
	// recovery_code_hashes is NOT NULL DEFAULT '{}'; a nil Go slice serialises to
	// SQL NULL (the default only applies when the column is omitted), which the
	// upsert always supplies — so normalise nil to an empty array.
	recoveryCodeHashes := account.RecoveryCodeHashes
	if recoveryCodeHashes == nil {
		recoveryCodeHashes = []string{}
	}
	return sqlcgen.UpsertCredentialsParams{
		UserID:             account.User.ID,
		Email:              normalizeEmail(account.Email),
		PasswordHash:       account.PasswordHash,
		MfaSecret:          account.MFASecret,
		RecoveryCodeHashes: recoveryCodeHashes,
	}
}

func accountFrom(id, name, role string, facilityID *string, prefs []byte, email, hash, mfa string, recoveryCodeHashes []string) (ports.Account, error) {
	p, err := unmarshalPrefs(prefs)
	if err != nil {
		return ports.Account{}, err
	}
	u, err := user.New(user.User{
		ID:          id,
		Name:        name,
		Role:        user.Role(role),
		FacilityID:  derefStr(facilityID),
		Preferences: p,
	})
	if err != nil {
		return ports.Account{}, fmt.Errorf("postgres: map account %q: %w", id, err)
	}
	return ports.Account{User: u, Email: email, PasswordHash: hash, MFASecret: mfa, RecoveryCodeHashes: recoveryCodeHashes}, nil
}
