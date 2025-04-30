package auth

import (
	"context"
	"database/sql"
)

type SessionStore interface {
	Save(ctx context.Context, session *Session) error
	GetByEmail(ctx context.Context, email string) (*Session, error)
	GetByToken(ctx context.Context, token string) (*Session, error)
	Delete(ctx context.Context, email string) error
	Cleanup(ctx context.Context) error // 만료 세션 정리
}

type StoreFactory func(useLocalMemoryStore bool, db *sql.DB) SessionStore
