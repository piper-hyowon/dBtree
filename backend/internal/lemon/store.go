package lemon

import (
	"database/sql"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
	"github.com/piper-hyowon/dBtree/internal/platform/store/postgres"
)

func NewLemonStore(_ bool, db *sql.DB) lemon.Store {
	return postgres.NewLemonStore(db)
}
