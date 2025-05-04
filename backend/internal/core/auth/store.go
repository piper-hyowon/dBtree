package auth

import (
	"context"
)

type SessionStore interface {
	Save(ctx context.Context, session *Session) error
	FindByEmail(ctx context.Context, email string) (*Session, error)
	FindByToken(ctx context.Context, token string) (*Session, error)
	Delete(ctx context.Context, email string) error
	Cleanup(ctx context.Context) error // 만료 세션 정리
}
