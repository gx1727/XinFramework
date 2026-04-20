package auth

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func verifyPassword(stored, plain string) (bool, error) {
	if !strings.HasPrefix(stored, "$argon2id$") {
		return subtle.ConstantTimeCompare([]byte(stored), []byte(plain)) == 1, nil
	}
	return verifyArgon2ID(stored, plain)
}

func verifyArgon2ID(encodedHash, plain string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid argon2id hash format")
	}
	if parts[1] != "argon2id" {
		return false, errors.New("unsupported hash algorithm")
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
