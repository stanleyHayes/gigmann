// Package passwordhash implements ports.PasswordHasher with argon2id and the
// standard PHC string encoding ($argon2id$v=19$m=...,t=...,p=...$salt$hash).
package passwordhash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"github.com/xcreativs/gigmann/internal/ports"
)

const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
	phcParts     = 6
)

// ErrBadHash is returned when an encoded hash cannot be parsed.
var ErrBadHash = errors.New("passwordhash: malformed encoded hash")

// Hasher is the argon2id implementation of ports.PasswordHasher.
type Hasher struct{}

// New returns an argon2id password hasher.
func New() Hasher { return Hasher{} }

var _ ports.PasswordHasher = Hasher{}

// Hash derives an argon2id PHC-encoded hash with a fresh random salt.
func (Hasher) Hash(plain string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("passwordhash: read salt: %w", err)
	}
	key := argon2.IDKey([]byte(plain), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// Verify reports whether plain matches the encoded argon2id hash (constant time).
func (Hasher) Verify(plain, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != phcParts || parts[1] != "argon2id" {
		return false, ErrBadHash
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, ErrBadHash
	}
	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, ErrBadHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, ErrBadHash
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, ErrBadHash
	}
	got := argon2.IDKey([]byte(plain), salt, time, memory, threads, uint32(len(want))) //nolint:gosec // hash length is small and non-negative
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
