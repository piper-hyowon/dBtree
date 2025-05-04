package auth

import (
	"database/sql"
	"github.com/piper-hyowon/dBtree/internal/core/auth"
)

func NewSessionStore(_ bool, db *sql.DB) auth.SessionStore {
	return NewPostgresStore(db)
}
