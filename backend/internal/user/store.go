package user

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"sync"
	"time"
)

type Store interface {
	FindByEmail(ctx context.Context, email string) (*User, error)

	FindById(ctx context.Context, id string) (*User, error)

	// 추가 파라미터 계획 없으므로 dto가 아닌 email만 사용
	Create(ctx context.Context, email string) error
}

// TODO: PostgreSQL 연동 후 교체

var (
	ErrUserNotFound = errors.New("user not found")
)

type store struct {
	mu           sync.RWMutex
	usersByID    map[string]*User
	usersByEmail map[string]*User
}

var _ Store = (*store)(nil)

func NewStore() Store {
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
		return nil, ErrUserNotFound
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
		return nil, ErrUserNotFound
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
