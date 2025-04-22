// TODO: PostgreSQL 연동 후 교체

package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/piper-hyowon/dBtree/internal/domain/model"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepo struct {
	mu           sync.RWMutex
	usersByID    map[string]*model.User
	usersByEmail map[string]*model.User
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		usersByID:    make(map[string]*model.User),
		usersByEmail: make(map[string]*model.User),
	}
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
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

func (r *UserRepo) FindById(ctx context.Context, id string) (*model.User, error) {
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

func (r *UserRepo) Create(ctx context.Context, email string) error {
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

	user := &model.User{
		ID:        id,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	r.usersByID[id] = user
	r.usersByEmail[email] = user

	return nil
}
