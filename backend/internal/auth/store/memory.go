package store

import (
	"context"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/auth"
	"github.com/piper-hyowon/dBtree/internal/common"
	"sync"
	"time"
)

type memorySessionStore struct {
	mu              sync.RWMutex
	sessionsByEmail map[string]*auth.Session
	sessionsByToken map[string]string
}

var _ auth.SessionStore = (*memorySessionStore)(nil)

func NewSessionStore() auth.SessionStore {
	return &memorySessionStore{
		sessionsByEmail: make(map[string]*auth.Session),
		sessionsByToken: make(map[string]string),
	}
}

func (r *memorySessionStore) Save(_ context.Context, session *auth.Session) error {
	if session == nil || session.Email == "" {
		return fmt.Errorf("invalid session: %w", common.ErrInternal)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	session.UpdatedAt = time.Now().UTC()

	oldSession, exists := r.sessionsByEmail[session.Email]
	if exists && oldSession.Token != "" {
		delete(r.sessionsByToken, oldSession.Token)
	}

	if session.Token != "" {
		r.sessionsByToken[session.Token] = session.Email
	}

	r.sessionsByEmail[session.Email] = session
	return nil
}

func (r *memorySessionStore) GetByEmail(_ context.Context, email string) (*auth.Session, error) {
	if email == "" {
		return nil, fmt.Errorf("empty email: %w", common.ErrInternal)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessionsByEmail[email]
	if !exists {
		return nil, common.ErrSessionNotFound
	}

	return session, nil
}

func (r *memorySessionStore) GetByToken(_ context.Context, token string) (*auth.Session, error) {
	if token == "" {
		return nil, fmt.Errorf("empty token: %w", common.ErrInvalidToken)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	email, exists := r.sessionsByToken[token]
	if !exists {
		return nil, common.ErrSessionNotFound
	}

	session, exists := r.sessionsByEmail[email]
	if !exists {
		// 토큰은 있으나 세션이 없음 - 정리 필요
		return nil, common.ErrSessionNotFound
	}

	// 토큰 만료 확인
	now := time.Now().UTC()
	if session.TokenExpiresAt.Before(now) {
		return nil, common.ErrTokenExpired
	}

	return session, nil
}

func (r *memorySessionStore) Delete(_ context.Context, email string) error {
	if email == "" {
		return fmt.Errorf("empty email: %w", common.ErrInternal)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	session, exists := r.sessionsByEmail[email]
	if !exists {
		return common.ErrSessionNotFound
	}

	if session.Token != "" {
		delete(r.sessionsByToken, session.Token)
	}

	delete(r.sessionsByEmail, email)
	return nil
}

func (r *memorySessionStore) Cleanup(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()

	for email, session := range r.sessionsByEmail {
		otpExpired := session.OTP != nil && session.OTP.ExpiresAt.Before(now)
		tokenExpired := session.Token != "" && session.TokenExpiresAt.Before(now)

		if otpExpired {
			// OTP가 만료된 경우 세션 전체 삭제
			if session.Token != "" {
				delete(r.sessionsByToken, session.Token)
			}
			delete(r.sessionsByEmail, email)
		} else if tokenExpired {
			// 토큰만 만료된 경우 토큰만 삭제
			delete(r.sessionsByToken, session.Token)
			session.Token = ""
			session.TokenExpiresAt = time.Time{}
			session.UpdatedAt = now
		}
	}

	return nil
}
