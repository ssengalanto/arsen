package jwt

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims extends jwt.RegisteredClaims for token validation.
type Claims struct {
	jwt.RegisteredClaims
}

// Service handles JWT creation and validation.
type Service struct {
	secret   []byte
	issuer   string
	audience string
}

// NewService creates a new JWT service with the given secret, issuer, and audience.
func NewService(secret, issuer, audience string) *Service {
	return &Service{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
	}
}

// CreateToken creates an HS256-signed JWT for the given user ID and duration.
func (s *Service) CreateToken(userID string, duration time.Duration) (string, error) {
	now := time.Now()

	jti, err := generateUUID()
	if err != nil {
		return "", fmt.Errorf("jwt: failed to generate jti: %w", err)
	}

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("jwt: failed to sign token: %w", err)
	}

	return signed, nil
}

// ValidateToken parses and validates a JWT string, returning the claims on success.
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("jwt: unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	},
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("jwt: invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("jwt: token is not valid")
	}

	return claims, nil
}

// generateUUID generates a random v4 UUID string from crypto/rand.
func generateUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Set version 4 (bits 12-15 of time_hi_and_version).
	b[6] = (b[6] & 0x0f) | 0x40
	// Set variant bits (bits 6-7 of clock_seq_hi_and_reserved).
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
