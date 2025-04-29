package user

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/piper-hyowon/dBtree/internal/common"
	"sync"
	"time"
)

type store struct {
	mu           sync.RWMutex
	usersByID    map[string]*User
	usersByEmail map[string]*User
}

var _ Store = (*store)(nil)

func NewMemoryStore() Store {
	return &store{
		usersByID:    make(map[string]*User),
		usersByEmail: make(map[string]*User),
	}
}

func (r *store) FindByEmail(_ context.Context, email string) (*User, error) {
	if email == "" {
		return nil, errors.New("empty Email")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.usersByEmail[email]
	if !exists {
		return nil, common.ErrUserNotFound
	}

	return user, nil
}

func (r *store) FindById(_ context.Context, id string) (*User, error) {
	if id == "" {
		return nil, errors.New("empty ID")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.usersByID[id]
	if !exists {
		return nil, common.ErrUserNotFound
	}

	return user, nil
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

	user := &User{
		ID:        id,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	r.usersByID[id] = user
	r.usersByEmail[email] = user

	return nil
}
