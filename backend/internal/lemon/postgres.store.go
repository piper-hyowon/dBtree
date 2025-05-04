package lemon

import (
	"context"
	"database/sql"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
	"runtime/debug"
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

func (s *PostgresStore) GetUserBalance(ctx context.Context, userID string) (int, error) {
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

func (s *PostgresStore) GetUserLastHarvestTime(ctx context.Context, userID string) (*time.Time, error) {
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

func (s *PostgresStore) UpdateUserLastHarvestTime(ctx context.Context, userID string, harvestTime time.Time) error {
	query := `
		UPDATE users 
		SET last_harvest_at = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := s.db.ExecContext(ctx, query, harvestTime, time.Now(), userID)
	if err != nil {
		return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	return nil
}
