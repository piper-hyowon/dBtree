package lemon

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
	"runtime/debug"
	"strconv"
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
		return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
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
		return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
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
			return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
		}
	}

	if err = dbTx.Commit(); err != nil {
		return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
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
		return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	return &tx, nil
}

func (s *PostgresStore) TransactionListByUserID(ctx context.Context, userID string, limit, offset int) ([]*lemon.Transaction, error) {
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
		return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
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
			return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
		}
		transactions = append(transactions, &tx)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	return transactions, nil
}

func (s *PostgresStore) TransactionListByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*lemon.Transaction, error) {
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
		return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
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
			return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
		}
		transactions = append(transactions, &tx)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	return transactions, nil
}

func (s *PostgresStore) UserBalance(ctx context.Context, userID string) (int, error) {
	query := `SELECT lemon_balance FROM users WHERE id = $1`

	var balance int
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.NewResourceNotFoundError("user", userID)
		}

		return 0, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	return balance, nil
}

func (s *PostgresStore) UserLastHarvestTime(ctx context.Context, userID string) (*time.Time, error) {
	query := `SELECT last_harvest_at FROM users WHERE id = $1`

	var lastHarvestAt sql.NullTime
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&lastHarvestAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.NewResourceNotFoundError("user", userID)
		}
		return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	if !lastHarvestAt.Valid {
		return nil, nil
	}

	return &lastHarvestAt.Time, nil
}

func (s *PostgresStore) AvailablePositions(ctx context.Context) ([]int, error) {
	query := `SELECT position_id FROM lemons WHERE is_available = true ORDER BY position_id`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}
	defer rows.Close()

	var positions []int
	for rows.Next() {
		var posID int
		if err := rows.Scan(&posID); err != nil {
			return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
		}
		positions = append(positions, posID)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	return positions, nil
}

func (s *PostgresStore) TotalHarvestedCount(ctx context.Context) (int, error) {
	query := `
        SELECT COUNT(*) 
        FROM user_lemon_transactions 
        WHERE action_type = 'harvest' AND status = 'successful'
    `

	var count int
	if err := s.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	return count, nil
}

func (s *PostgresStore) UserTotalHarvestedCount(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM user_lemon_transactions
		WHERE action_type = 'harvest' AND status = 'successful' AND user_id = $1
	`

	var count int
	if err := s.db.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		return 0, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	return count, nil
}

func (s *PostgresStore) HarvestWithTransaction(ctx context.Context, positionID int, userID string, harvestAmount int, now time.Time) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}
	defer tx.Rollback()

	var isAvailable bool
	query := `SELECT is_available FROM lemons WHERE position_id = $1 FOR UPDATE`
	if err := tx.QueryRowContext(ctx, query, positionID).Scan(&isAvailable); err != nil {
		return "", errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	if !isAvailable {
		return "", errors.NewResourceNotFoundError("available_lemon", strconv.Itoa(positionID))
	}

	// 레몬 수확 처리
	nextTime := now.Add(lemon.DefaultHarvestRules.CooldownPeriod)
	updateQuery := `
        UPDATE lemons 
        SET is_available = false, 
            last_harvested_at = $1,
            next_available_at = $2
        WHERE position_id = $3
    `
	if _, err := tx.ExecContext(ctx, updateQuery, now, nextTime, positionID); err != nil {
		return "", errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	// 사용자 잔액 조회
	var balance int
	balanceQuery := `SELECT lemon_balance FROM users WHERE id = $1 FOR UPDATE`
	if err := tx.QueryRowContext(ctx, balanceQuery, userID).Scan(&balance); err != nil {
		return "", errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	newBalance := balance + harvestAmount

	// 유저 마지막 수확 시간, 잔액 업데이트
	userUpdateQuery := `
        UPDATE users 
        SET last_harvest_at = $1, updated_at = $2, lemon_balance = $3
        WHERE id = $4
    `
	if _, err := tx.ExecContext(ctx, userUpdateQuery, now, now, newBalance, userID); err != nil {
		return "", errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	txID := uuid.New().String()
	txQuery := `
        INSERT INTO "user_lemon_transactions" (
            id, user_id, action_type, status, 
            amount, balance, created_at, note
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8
        )
    `
	note := fmt.Sprintf("레몬 위치 %d에서 수확", positionID)
	if _, err := tx.ExecContext(
		ctx,
		txQuery,
		txID,
		userID,
		lemon.ActionHarvest,
		lemon.StatusSuccessful,
		harvestAmount,
		newBalance,
		now,
		now,
		note,
	); err != nil {
		return "", errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	if err := tx.Commit(); err != nil {
		return "", errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	return txID, nil
}

func (s *PostgresStore) RegrowLemons(ctx context.Context, now time.Time) (int, error) {
	query := `
        UPDATE lemons 
        SET is_available = true 
        WHERE is_available = false 
          AND next_available_at <= $1
        RETURNING position_id
    `

	rows, err := s.db.QueryContext(ctx, query, now)
	if err != nil {
		return 0, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}
	defer rows.Close()

	var regrown int
	for rows.Next() {
		var posID int
		if err := rows.Scan(&posID); err != nil {
			return 0, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
		}
		regrown++
	}

	return regrown, nil
}

func (s *PostgresStore) NextRegrowthTime(ctx context.Context) (*time.Time, error) {
	query := `
        SELECT MIN(next_available_at)
        FROM lemons
        WHERE is_available = false AND next_available_at IS NOT NULL
    `
	var nextTime sql.NullTime
	if err := s.db.QueryRowContext(ctx, query).Scan(&nextTime); err != nil {
		return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	if !nextTime.Valid {
		// 재생성 예정인 레몬이 없음 - 현재 시간 반환
		return nil, nil
	}

	return &nextTime.Time, nil
}
