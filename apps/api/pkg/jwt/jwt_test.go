package jwt_test

import (
	"testing"
	"time"

	"arsen/pkg/jwt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSecret   = "super-secret-key-for-testing-only"
	testIssuer   = "test-issuer"
	testAudience = "test-audience"
	testUserID   = "user-123"
)

func newTestService() *jwt.Service {
	return jwt.NewService(testSecret, testIssuer, testAudience)
}

func TestCreateToken_ReturnsNonEmptyString(t *testing.T) {
	svc := newTestService()
	token, err := svc.CreateToken(testUserID, 15*time.Minute)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateToken_SucceedsForValidToken(t *testing.T) {
	svc := newTestService()
	token, err := svc.CreateToken(testUserID, 15*time.Minute)
	require.NoError(t, err)

	claims, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, testUserID, claims.Subject)
}

func TestValidateToken_RejectsExpiredToken(t *testing.T) {
	svc := newTestService()
	token, err := svc.CreateToken(testUserID, -1*time.Minute)
	require.NoError(t, err)

	_, err = svc.ValidateToken(token)
	assert.Error(t, err, "expired token should be rejected")
}

func TestValidateToken_RejectsTamperedToken(t *testing.T) {
	svc := newTestService()
	token, err := svc.CreateToken(testUserID, 15*time.Minute)
	require.NoError(t, err)

	// Flip a character in the middle of the token to simulate tampering.
	mid := len(token) / 2
	tampered := []byte(token)
	if tampered[mid] == 'A' {
		tampered[mid] = 'B'
	} else {
		tampered[mid] = 'A'
	}

	_, err = svc.ValidateToken(string(tampered))
	assert.Error(t, err, "tampered token should be rejected")
}

func TestValidateToken_RejectsWrongIssuer(t *testing.T) {
	svcA := jwt.NewService(testSecret, "issuer-a", testAudience)
	svcB := jwt.NewService(testSecret, "issuer-b", testAudience)

	token, err := svcA.CreateToken(testUserID, 15*time.Minute)
	require.NoError(t, err)

	_, err = svcB.ValidateToken(token)
	assert.Error(t, err, "token with wrong issuer should be rejected")
}

func TestValidateToken_RejectsWrongAudience(t *testing.T) {
	svcA := jwt.NewService(testSecret, testIssuer, "audience-a")
	svcB := jwt.NewService(testSecret, testIssuer, "audience-b")

	token, err := svcA.CreateToken(testUserID, 15*time.Minute)
	require.NoError(t, err)

	_, err = svcB.ValidateToken(token)
	assert.Error(t, err, "token with wrong audience should be rejected")
}

func TestValidateToken_ClaimsContainExpectedFields(t *testing.T) {
	svc := newTestService()
	token, err := svc.CreateToken(testUserID, 15*time.Minute)
	require.NoError(t, err)

	claims, err := svc.ValidateToken(token)
	require.NoError(t, err)

	assert.Equal(t, testUserID, claims.Subject, "Subject should match the user ID")
	assert.Equal(t, testIssuer, claims.Issuer, "Issuer should match")
	assert.Contains(t, claims.Audience, testAudience, "Audience should contain the expected audience")
	assert.NotNil(t, claims.ExpiresAt, "ExpiresAt should be set")
	assert.NotNil(t, claims.IssuedAt, "IssuedAt should be set")
	assert.NotEmpty(t, claims.ID, "ID (jti) should be populated")
}

func TestCreateToken_DifferentTokensHaveDifferentJTI(t *testing.T) {
	svc := newTestService()

	token1, err := svc.CreateToken(testUserID, 15*time.Minute)
	require.NoError(t, err)

	token2, err := svc.CreateToken(testUserID, 15*time.Minute)
	require.NoError(t, err)

	claims1, err := svc.ValidateToken(token1)
	require.NoError(t, err)

	claims2, err := svc.ValidateToken(token2)
	require.NoError(t, err)

	assert.NotEqual(t, claims1.ID, claims2.ID, "two tokens for the same user should have different JTI values")
}
