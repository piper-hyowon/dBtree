package lemon

import (
	"database/sql"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
)

func NewLemonStore(useLocalMemoryStore bool, db *sql.DB) lemon.Store {
	return NewPostgresStore(db)
}
