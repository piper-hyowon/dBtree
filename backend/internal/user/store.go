package user

import (
	"database/sql"
	"github.com/piper-hyowon/dBtree/internal/core/user"
)

func NewStore(useLocalMemoryStore bool, db *sql.DB) user.Store {
	if useLocalMemoryStore {
		return NewMemoryStore()
	} else {
		return NewPostgresStore(db)
	}
}
