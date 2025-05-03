package lemon

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
	"time"
)

type PostgresStore struct {
	db *sql.DB
}

var _ lemon.Store = (*PostgresStore)(nil)

func NewPostgresStore(db *sql.DB) lemon.Store {
	return &PostgresStore{
		db: db,
	}
}

func (s *PostgresStore) CreateTransaction(ctx context.Context, tx *lemon.Transaction) error {
	dbTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("트랜잭션 시작 실패: %w", err)
	}
	defer dbTx.Rollback()

	query := `
		INSERT INTO "user_lemon_transactions" (
			id, user_id, db_instance_id, action_type, status, 
			amount, balance, created_at, note
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	_, err = dbTx.ExecContext(
		ctx,
		query,
		tx.ID,
		tx.UserID,
		tx.InstanceID,
		tx.ActionType,
		tx.Status,
		tx.Amount,
		tx.Balance,
		tx.CreatedAt,
		tx.Note,
	)

	if err != nil {
		return fmt.Errorf("트랜잭션 삽입 실패: %w", err)
	}

	if tx.Status == lemon.StatusSuccessful {
		updateQuery := `
			UPDATE users 
			SET lemon_balance = $1, updated_at = $2
			WHERE id = $3
		`

		_, err = dbTx.ExecContext(
			ctx,
			updateQuery,
			tx.Balance,
			time.Now(),
			tx.UserID,
		)

		if err != nil {
			return fmt.Errorf("사용자 잔액 업데이트 실패: %w", err)
		}
	}

	if err = dbTx.Commit(); err != nil {
		return fmt.Errorf("트랜잭션 커밋 실패: %w", err)
	}

	return nil
}

func (s *PostgresStore) FindTransactionByID(ctx context.Context, id string) (*lemon.Transaction, error) {
	query := `
		SELECT 
			id, user_id, db_instance_id, action_type, status, 
			amount, balance, created_at, note
		FROM user_lemon_transactions
		WHERE id = $1
	`

	var tx lemon.Transaction
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&tx.ID,
		&tx.UserID,
		&tx.InstanceID,
		&tx.ActionType,
		&tx.Status,
		&tx.Amount,
		&tx.Balance,
		&tx.CreatedAt,
		&tx.Note,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("트랜잭션 조회 실패: %w", err)
	}

	return &tx, nil
}

func (s *PostgresStore) FindTransactionsByUserID(ctx context.Context, userID string, limit, offset int) ([]*lemon.Transaction, error) {
	query := `
		SELECT 
			id, user_id, db_instance_id, action_type, status, 
			amount, balance, created_at, note
		FROM user_lemon_transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("트랜잭션 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	var transactions []*lemon.Transaction
	for rows.Next() {
		var tx lemon.Transaction
		err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.InstanceID,
			&tx.ActionType,
			&tx.Status,
			&tx.Amount,
			&tx.Balance,
			&tx.CreatedAt,
			&tx.Note,
		)
		if err != nil {
			return nil, fmt.Errorf("트랜잭션 스캔 실패: %w", err)
		}
		transactions = append(transactions, &tx)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("트랜잭션 행 처리 중 오류: %w", err)
	}

	return transactions, nil
}

func (s *PostgresStore) FindTransactionsByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*lemon.Transaction, error) {
	query := `
        SELECT 
            id, user_id, db_instance_id, action_type, status, 
            amount, balance, created_at, note
        FROM user_lemon_transactions
        WHERE db_instance_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `

	rows, err := s.db.QueryContext(ctx, query, instanceID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("인스턴스별 트랜잭션 조회 실패: %w", err)
	}
	defer rows.Close()

	var transactions []*lemon.Transaction
	for rows.Next() {
		var tx lemon.Transaction
		err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.InstanceID,
			&tx.ActionType,
			&tx.Status,
			&tx.Amount,
			&tx.Balance,
			&tx.CreatedAt,
			&tx.Note,
		)
		if err != nil {
			return nil, fmt.Errorf("트랜잭션 스캔 실패: %w", err)
		}
		transactions = append(transactions, &tx)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("트랜잭션 행 처리 중 오류: %w", err)
	}

	return transactions, nil
}

func (s *PostgresStore) GetUserBalance(ctx context.Context, userID string) (int, error) {
	query := `SELECT lemon_balance FROM users WHERE id = $1`

	var balance int
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("사용자를 찾을 수 없음")
		}
		return 0, fmt.Errorf("사용자 잔액 조회 실패: %w", err)
	}

	return balance, nil
}

func (s *PostgresStore) GetUserLastHarvestTime(ctx context.Context, userID string) (*time.Time, error) {
	query := `SELECT last_harvest_at FROM users WHERE id = $1`

	var lastHarvestAt sql.NullTime
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&lastHarvestAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("사용자를 찾을 수 없음")
		}
		return nil, fmt.Errorf("마지막 수확 시간 조회 실패: %w", err)
	}

	if !lastHarvestAt.Valid {
		return nil, nil
	}

	return &lastHarvestAt.Time, nil
}

func (s *PostgresStore) UpdateUserLastHarvestTime(ctx context.Context, userID string, harvestTime time.Time) error {
	query := `
		UPDATE users 
		SET last_harvest_at = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := s.db.ExecContext(ctx, query, harvestTime, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("마지막 수확 시간 업데이트 실패: %w", err)
	}

	return nil
}
