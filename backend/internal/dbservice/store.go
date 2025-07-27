package dbservice

import (
	"database/sql"
	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/platform/store/postgres"
)

func NewDBInstanceStore(_ bool, db *sql.DB) dbservice.DBInstanceStore {
	return postgres.NewDBInstanceStore(db)
}
