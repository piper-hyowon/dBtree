package postgres

import (
	"context"
	"database/sql"
	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
)

type PortStore struct {
	db *sql.DB
}

var _ dbservice.PortStore = (*PortStore)(nil)

func NewPortStore(db *sql.DB) dbservice.PortStore {
	return &PortStore{db: db}
}

func (s *PortStore) AllocatePort(ctx context.Context, instanceID string) (int, error) {
	var port int

	err := s.db.QueryRowContext(ctx, `
        INSERT INTO port_allocations (instance_id, port)
        SELECT $1, port FROM (
            SELECT generate_series(30000, 31999) AS port
        ) AS available_ports
        WHERE port NOT IN (
            SELECT port FROM port_allocations
        )
        ORDER BY port
        LIMIT 1
        RETURNING port
    `, instanceID).Scan(&port)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.NewError(errors.ErrInternalServer,
				"사용 가능한 포트가 없습니다", nil, nil)
		}
		return 0, errors.Wrap(err)
	}

	return port, nil
}

func (s *PortStore) ReleasePort(ctx context.Context, instanceID string) error {
	_, err := s.db.ExecContext(ctx, `
        DELETE FROM port_allocations 
        WHERE instance_id = $1
    `, instanceID)

	return errors.Wrap(err)
}

func (s *PortStore) GetPort(ctx context.Context, instanceID string) (int, error) {
	var port int

	err := s.db.QueryRowContext(ctx, `
        SELECT port 
        FROM port_allocations 
        WHERE instance_id = $1
    `, instanceID).Scan(&port)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil // 포트 없음은 에러가 아님
		}
		return 0, errors.Wrap(err)
	}

	return port, nil
}
