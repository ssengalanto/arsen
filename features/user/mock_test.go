package user

import (
	"context"
	"encoding/hex"
	"sync"
	"time"

	"arsen/pkg/cqrs"
	"arsen/pkg/email"

	"github.com/google/uuid"
)

// mockRepository implements Repository for testing.
type mockRepository struct {
	mu                     sync.Mutex
	users                  map[string]*User
	usersByEmail           map[string]*User
	verificationTokens     map[string]*VerificationToken // keyed by hex(token_hash)
	invalidatedUserIDs     []string                      // tracks InvalidateUserVerificationTokens calls
	createErr              error
	getByEmailErr          error
	getByIDErr             error
	updateEmailVerifiedErr error
	updatePasswordHashErr  error
	createTokenErr         error
	getTokenByHashErr      error
	invalidateTokensErr    error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		users:              make(map[string]*User),
		usersByEmail:       make(map[string]*User),
		verificationTokens: make(map[string]*VerificationToken),
	}
}

func (m *mockRepository) Create(_ context.Context, user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.createErr != nil {
		return m.createErr
	}
	user.ID = uuid.New().String()
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	m.users[user.ID] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *mockRepository) GetByEmail(_ context.Context, em string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getByEmailErr != nil {
		return nil, m.getByEmailErr
	}
	u, ok := m.usersByEmail[em]
	if !ok {
		return nil, cqrs.NewNotFoundError("User", em)
	}
	return u, nil
}

func (m *mockRepository) GetByID(_ context.Context, id string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	u, ok := m.users[id]
	if !ok {
		return nil, cqrs.NewNotFoundError("User", id)
	}
	return u, nil
}

func (m *mockRepository) UpdateEmailVerified(_ context.Context, id string, verified bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.updateEmailVerifiedErr != nil {
		return m.updateEmailVerifiedErr
	}
	u, ok := m.users[id]
	if !ok {
		return cqrs.NewNotFoundError("User", id)
	}
	u.EmailVerified = verified
	u.UpdatedAt = time.Now()
	return nil
}

func (m *mockRepository) UpdatePasswordHash(_ context.Context, id string, hash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.updatePasswordHashErr != nil {
		return m.updatePasswordHashErr
	}
	u, ok := m.users[id]
	if !ok {
		return cqrs.NewNotFoundError("User", id)
	}
	u.PasswordHash = hash
	u.UpdatedAt = time.Now()
	return nil
}

func (m *mockRepository) CreateVerificationToken(_ context.Context, vt *VerificationToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.createTokenErr != nil {
		return m.createTokenErr
	}
	vt.ID = uuid.New().String()
	vt.CreatedAt = time.Now()
	key := hex.EncodeToString(vt.TokenHash)
	m.verificationTokens[key] = vt
	return nil
}

func (m *mockRepository) GetVerificationTokenByHash(_ context.Context, hash []byte) (*VerificationToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getTokenByHashErr != nil {
		return nil, m.getTokenByHashErr
	}
	key := hex.EncodeToString(hash)
	vt, ok := m.verificationTokens[key]
	if !ok {
		return nil, cqrs.NewNotFoundError("VerificationToken", "hash")
	}
	if vt.UsedAt != nil {
		return nil, cqrs.NewNotFoundError("VerificationToken", "hash")
	}
	return vt, nil
}

func (m *mockRepository) InvalidateUserVerificationTokens(_ context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.invalidateTokensErr != nil {
		return m.invalidateTokensErr
	}
	m.invalidatedUserIDs = append(m.invalidatedUserIDs, userID)
	now := time.Now()
	for _, vt := range m.verificationTokens {
		if vt.UserID == userID && vt.UsedAt == nil {
			vt.UsedAt = &now
		}
	}
	return nil
}

// mockEmailSender implements email.Sender for testing.
type mockEmailSender struct {
	mu      sync.Mutex
	sent    []email.SendParams
	sendErr error
}

func (m *mockEmailSender) Send(_ context.Context, params email.SendParams) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sent = append(m.sent, params)
	return nil
}

// testEventHandler records EmailVerifiedEvent calls for assertions.
type testEventHandler struct {
	mu     sync.Mutex
	events []EmailVerifiedEvent
}

func (h *testEventHandler) Handle(_ context.Context, e EmailVerifiedEvent) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, e)
	return nil
}
