package auth

import (
	"database/sql"
	"github.com/piper-hyowon/dBtree/internal/core/auth"
	"github.com/piper-hyowon/dBtree/internal/platform/store/postgres"
)

func NewSessionStore(_ bool, db *sql.DB) auth.SessionStore {
	return postgres.NewSessionStore(db)
}
