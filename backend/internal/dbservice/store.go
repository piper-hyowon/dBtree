package dbservice

import (
	"database/sql"
	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/platform/store/postgres"
)

func NewDBIStore(_ bool, db *sql.DB) dbservice.DBInstanceStore {
	return postgres.NewDBInstanceStore(db)
}

func NewPresetStore(_ bool, db *sql.DB) dbservice.PresetStore {
	return postgres.NewPresetStore(db)
}
