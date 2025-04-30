package auth

import (
	"database/sql"
	"github.com/piper-hyowon/dBtree/internal/core/auth"
)

func NewSessionStore(useLocalMemoryStore bool, db *sql.DB) auth.SessionStore {
	if useLocalMemoryStore {
		return NewMemoryStore()
	} else {
		return NewPostgresStore(db)
	}
}
