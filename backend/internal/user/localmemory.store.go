package user

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/piper-hyowon/dBtree/internal/common"
	"github.com/piper-hyowon/dBtree/internal/common/user"
	"sync"
	"time"
)

type store struct {
	mu           sync.RWMutex
	usersByID    map[string]*user.User
	usersByEmail map[string]*user.User
}

var _ user.Store = (*store)(nil)

func NewMemoryStore() user.Store {
	return &store{
		usersByID:    make(map[string]*user.User),
		usersByEmail: make(map[string]*user.User),
	}
}

func (r *store) FindByEmail(_ context.Context, email string) (*user.User, error) {
	if email == "" {
		return nil, errors.New("empty Email")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	u, exists := r.usersByEmail[email]
	if !exists {
		return nil, common.ErrUserNotFound
	}

	return u, nil
}

func (r *store) FindById(_ context.Context, id string) (*user.User, error) {
	if id == "" {
		return nil, errors.New("empty ID")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	u, exists := r.usersByID[id]
	if !exists {
		return nil, common.ErrUserNotFound
	}

	return u, nil
}

func (r *store) Create(_ context.Context, email string) error {
	if email == "" {
		return errors.New("empty Email")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.usersByEmail[email]; exists {
		return errors.New("duplicated email")
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	u := &user.User{
		ID:        id,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	r.usersByID[id] = u
	r.usersByEmail[email] = u

	return nil
}

func (r *store) Delete(_ context.Context, id string) error {
	if id == "" {
		return errors.New("empty ID")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	u, exists := r.usersByID[id]
	if !exists {
		return common.ErrUserNotFound
	}

	delete(r.usersByID, id)
	delete(r.usersByEmail, u.Email)

	return nil
}
