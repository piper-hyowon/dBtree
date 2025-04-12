package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/piper-hyowon/dBtree/internal/domain/model"
)

var (
	ErrSessionNotFound = errors.New("세션 404")
)

type SessionRepo struct {
	mu       sync.RWMutex
	sessions map[string]*model.AuthSession // key는 이메일
}

func NewSessionRepo() *SessionRepo {
	return &SessionRepo{
		sessions: make(map[string]*model.AuthSession),
	}
}

// upsert
func (r *SessionRepo) Save(ctx context.Context, session *model.AuthSession) error {
	if session == nil || session.Email == "" {
		return errors.New("세션 or 이메일 404")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	session.UpdatedAt = time.Now().UTC()

	r.sessions[session.Email] = session
	return nil
}

func (r *SessionRepo) Get(ctx context.Context, email string) (*model.AuthSession, error) {
	if email == "" {
		return nil, errors.New("empty string")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[email]
	if !exists {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

func (r *SessionRepo) Delete(ctx context.Context, email string) error {
	if email == "" {
		return errors.New("empty string")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[email]; !exists {
		return ErrSessionNotFound
	}

	delete(r.sessions, email)
	return nil
}

func (r *SessionRepo) Cleanup(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	expiredEmails := []string{}

	for email, session := range r.sessions {
		if session.OTP != nil && session.OTP.ExpiresAt.Before(now) {
			expiredEmails = append(expiredEmails, email)
		}
	}

	for _, email := range expiredEmails {
		delete(r.sessions, email)
	}

	return nil
}
