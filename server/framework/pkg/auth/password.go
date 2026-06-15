package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// HashPassword hashes the given plain-text password using Argon2id
// with the same parameters apps/boot/auth uses.
//
// Moved from framework/internal/module/auth/password.go to framework/pkg/auth
// during Phase 2: framework's bootstrap (in boot/bootstrap.go) needs to
// create admin accounts at startup, but apps/boot/auth can't be imported
// from framework/internal.
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	iterations := uint32(3)
	memory := uint32(64 * 1024)
	parallelism := uint8(4)
	keyLength := uint32(32)

	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, iterations, parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash)), nil
}

// VerifyPassword verifies a plain-text password against an Argon2id
// encoded hash. Returns (true, nil) on success, (false, nil) on
// mismatch, and (false, err) on malformed hash. Used by apps/boot/auth
// and any external authenticator that needs to verify against this
// framework's hashes.
func VerifyPassword(stored, plain string) (bool, error) {
	if !strings.HasPrefix(stored, "$argon2id$") {
		return false, nil
	}
	parts := strings.Split(stored, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}
	if parts[1] != "argon2id" {
		return false, fmt.Errorf("unsupported hash type")
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	keyLength := uint32(len(hash))
	calculated := argon2.IDKey([]byte(plain), salt, iterations, memory, parallelism, keyLength)
	return subtle.ConstantTimeCompare(hash, calculated) == 1, nil
}