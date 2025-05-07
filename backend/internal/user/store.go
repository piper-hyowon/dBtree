package user

import (
	"database/sql"
	"github.com/piper-hyowon/dBtree/internal/core/user"
	"github.com/piper-hyowon/dBtree/internal/platform/store/postgres"
)

func NewStore(_ bool, db *sql.DB) user.Store {
	return postgres.NewUserStore(db)
}
