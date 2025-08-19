package postgres

import (
	"context"
	"database/sql"
	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
)

// TODO: 값 설정
const (
	MinNodePort = 30000
	MaxNodePort = 31999
)

type PortStore struct {
	db *sql.DB
}

var _ dbservice.PortStore = (*PortStore)(nil)

func NewPortStore(db *sql.DB) dbservice.PortStore {
	return &PortStore{db: db}
}

func (s *PortStore) AllocatePort(ctx context.Context, instanceID string) (int, error) {
	// 이미 할당된 포트가 있는지 확인
	var existingPort int
	err := s.db.QueryRowContext(ctx,
		"SELECT port FROM port_allocations WHERE instance_id = $1",
		instanceID).Scan(&existingPort)

	if err == nil {
		return existingPort, nil
	}

	// 사용 가능한 첫 번째 포트 찾기
	for port := MinNodePort; port <= MaxNodePort; port++ {
		if dbservice.ReservedPorts[port] {
			continue
		}

		_, err := s.db.ExecContext(ctx,
			"INSERT INTO port_allocations (instance_id, port) VALUES ($1, $2)",
			instanceID, port)

		if err == nil {
			return port, nil
		}

		// UNIQUE 제약 위반이면 다음 포트 시도
		if isUniqueViolation(err, "port_allocations_port_key") {
			continue
		}

		return 0, err
	}

	return 0, errors.New("no available ports")
}

func (s *PortStore) ReleasePort(ctx context.Context, instanceID string) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM port_allocations WHERE instance_id = $1",
		instanceID)
	return err
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
