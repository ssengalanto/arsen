package token

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

func Generate() (raw string, hash []byte, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", nil, fmt.Errorf("token: failed to generate random bytes: %w", err)
	}

	raw = hex.EncodeToString(b)
	h := sha256.Sum256(b)
	return raw, h[:], nil
}

func Hash(raw string) ([]byte, error) {
	b, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("token: invalid hex string: %w", err)
	}

	h := sha256.Sum256(b)
	return h[:], nil
}

func Compare(raw string, hash []byte) bool {
	h, err := Hash(raw)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(h, hash) == 1
}
