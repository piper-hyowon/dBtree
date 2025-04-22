package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/piper-hyowon/dBtree/internal/domain/errors"
	"github.com/piper-hyowon/dBtree/internal/domain/model"
)

type SessionRepo struct {
	mu              sync.RWMutex
	sessionsByEmail map[string]*model.AuthSession
	sessionsByToken map[string]string
}

func NewSessionRepo() *SessionRepo {
	return &SessionRepo{
		sessionsByEmail: make(map[string]*model.AuthSession),
		sessionsByToken: make(map[string]string),
	}
}

func (r *SessionRepo) Save(ctx context.Context, session *model.AuthSession) error {
	if session == nil || session.Email == "" {
		return fmt.Errorf("invalid session: %w", errors.ErrInternal)
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

func (r *SessionRepo) GetByEmail(ctx context.Context, email string) (*model.AuthSession, error) {
	if email == "" {
		return nil, fmt.Errorf("empty email: %w", errors.ErrInternal)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessionsByEmail[email]
	if !exists {
		return nil, errors.ErrSessionNotFound
	}

	return session, nil
}

func (r *SessionRepo) GetByToken(ctx context.Context, token string) (*model.AuthSession, error) {
	if token == "" {
		return nil, fmt.Errorf("empty token: %w", errors.ErrInvalidToken)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	email, exists := r.sessionsByToken[token]
	if !exists {
		return nil, errors.ErrSessionNotFound
	}

	session, exists := r.sessionsByEmail[email]
	if !exists {
		// 토큰은 있으나 세션이 없음 - 정리 필요
		return nil, errors.ErrSessionNotFound
	}

	// 토큰 만료 확인
	now := time.Now().UTC()
	if session.TokenExpiresAt.Before(now) {
		return nil, errors.ErrTokenExpired
	}

	return session, nil
}

func (r *SessionRepo) Delete(ctx context.Context, email string) error {
	if email == "" {
		return fmt.Errorf("empty email: %w", errors.ErrInternal)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	session, exists := r.sessionsByEmail[email]
	if !exists {
		return errors.ErrSessionNotFound
	}

	if session.Token != "" {
		delete(r.sessionsByToken, session.Token)
	}

	delete(r.sessionsByEmail, email)
	return nil
}

func (r *SessionRepo) Cleanup(ctx context.Context) error {
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
