package user

import (
	"database/sql"
	"github.com/piper-hyowon/dBtree/internal/core/user"
)

func NewStore(_ bool, db *sql.DB) user.Store {
	return NewPostgresStore(db)
}
