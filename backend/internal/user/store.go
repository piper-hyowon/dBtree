package user

import (
	"context"
	"database/sql"
)

type Store interface {
	FindByEmail(ctx context.Context, email string) (*User, error)

	FindById(ctx context.Context, id string) (*User, error)

	// 추가 파라미터 계획 없으므로 dto가 아닌 email만 사용
	Create(ctx context.Context, email string) error
}

func NewStore(useLocalMemoryStore bool, db *sql.DB) Store {
	if useLocalMemoryStore {
		return NewMemoryStore()
	} else {
		return NewPostgresStore(db)
	}
}
