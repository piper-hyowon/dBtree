package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
	"time"
)

type LemonStore struct {
	db *sql.DB
}

var _ lemon.Store = (*LemonStore)(nil)

func NewLemonStore(db *sql.DB) lemon.Store {
	return &LemonStore{
		db: db,
	}
}

func (s *LemonStore) ByPositionID(ctx context.Context, positionID int) (*lemon.Lemon, error) {
	query := `SELECT position_id, is_available, last_harvested_at, next_available_at  FROM lemons WHERE position_id = $1`

	var lemonData lemon.Lemon
	var lastHarvestedAtNull sql.NullTime
	var nextAvailableAtNull sql.NullTime
	err := s.db.QueryRowContext(ctx, query, positionID).Scan(
		&lemonData.PositionID,
		&lemonData.IsAvailable,
		&lastHarvestedAtNull,
		&nextAvailableAtNull,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err)
	}

	if lastHarvestedAtNull.Valid {
		lemonData.LastHarvestedAt = lastHarvestedAtNull.Time
	}

	if nextAvailableAtNull.Valid {
		lemonData.NextAvailableAt = nextAvailableAtNull.Time
	}

	return &lemonData, nil
}

func (s *LemonStore) CreateTransaction(ctx context.Context, tx *lemon.Transaction) error {
	dbTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err)
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
		return errors.Wrap(err)
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
			return errors.Wrap(err)
		}
	}

	if err = dbTx.Commit(); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func (s *LemonStore) TransactionByID(ctx context.Context, id string) (*lemon.Transaction, error) {
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
		return nil, errors.Wrap(err)
	}

	return &tx, nil
}

func (s *LemonStore) UserBalance(ctx context.Context, userID string) (int, error) {
	query := `SELECT lemon_balance FROM users WHERE id = $1`

	var balance int
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.NewResourceNotFoundError("user", userID)
		}

		return 0, errors.Wrap(err)
	}

	return balance, nil
}

func (s *LemonStore) UserLastHarvestTime(ctx context.Context, userID string) (*time.Time, error) {
	query := `SELECT last_harvest_at FROM users WHERE id = $1`

	var lastHarvestAt sql.NullTime
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&lastHarvestAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.NewResourceNotFoundError("user", userID)
		}
		return nil, errors.Wrap(err)
	}

	if !lastHarvestAt.Valid {
		return nil, nil
	}

	return &lastHarvestAt.Time, nil
}

func (s *LemonStore) AvailablePositions(ctx context.Context) ([]int, error) {
	query := `SELECT position_id FROM lemons WHERE is_available = true ORDER BY position_id`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer rows.Close()

	var positions = []int{}
	for rows.Next() {
		var posID int
		if err := rows.Scan(&posID); err != nil {
			return nil, errors.Wrap(err)
		}
		positions = append(positions, posID)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err)
	}

	return positions, nil
}

func (s *LemonStore) TotalHarvestedCount(ctx context.Context) (int, error) {
	query := `
        SELECT COUNT(*) 
        FROM user_lemon_transactions 
        WHERE action_type = 'harvest' AND status = 'successful'
    `

	var count int
	if err := s.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, errors.Wrap(err)
	}

	return count, nil
}

func (s *LemonStore) UserTotalHarvestedCount(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM user_lemon_transactions
		WHERE action_type = 'harvest' AND status = 'successful' AND user_id = $1
	`

	var count int
	if err := s.db.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		return 0, errors.Wrap(err)
	}

	return count, nil
}

func (s *LemonStore) HarvestWithTransaction(ctx context.Context, positionID int, userID string, harvestAmount int, now time.Time) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", errors.Wrap(err)
	}
	defer tx.Rollback()

	var isAvailable bool
	query := `SELECT is_available FROM lemons WHERE position_id = $1 FOR UPDATE`
	if err := tx.QueryRowContext(ctx, query, positionID).Scan(&isAvailable); err != nil {
		return "", errors.Wrap(err)
	}

	if !isAvailable {
		return "", errors.NewLemonAlreadyHarvestedError()
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
		return "", errors.Wrap(err)
	}

	// 사용자 잔액 조회
	var balance int
	balanceQuery := `SELECT lemon_balance FROM users WHERE id = $1 FOR UPDATE`
	if err := tx.QueryRowContext(ctx, balanceQuery, userID).Scan(&balance); err != nil {
		return "", errors.Wrap(err)
	}

	newBalance := balance + harvestAmount

	// 유저 마지막 수확 시간, 잔액 업데이트
	userUpdateQuery := `
        UPDATE users 
        SET last_harvest_at = $1, updated_at = $2, lemon_balance = $3
        WHERE id = $4
    `
	if _, err := tx.ExecContext(ctx, userUpdateQuery, now, now, newBalance, userID); err != nil {
		return "", errors.Wrap(err)
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
		note,
	); err != nil {
		return "", errors.Wrap(err)
	}

	if err := tx.Commit(); err != nil {
		return "", errors.Wrap(err)
	}

	return txID, nil
}

func (s *LemonStore) RegrowLemons(ctx context.Context, now time.Time) ([]int, error) {
	query := `
        UPDATE lemons 
        SET is_available = true 
        WHERE is_available = false 
          AND next_available_at <= $1
        RETURNING position_id
    `

	rows, err := s.db.QueryContext(ctx, query, now)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer rows.Close()

	var positionIDs []int
	for rows.Next() {
		var posID int
		if err := rows.Scan(&posID); err != nil {
			return nil, errors.Wrap(err)
		}
		positionIDs = append(positionIDs, posID)
	}

	return positionIDs, nil
}

func (s *LemonStore) NextRegrowthTime(ctx context.Context) (*time.Time, error) {
	query := `
        SELECT MIN(next_available_at)
        FROM lemons
        WHERE is_available = false AND next_available_at IS NOT NULL
    `
	var nextTime sql.NullTime
	if err := s.db.QueryRowContext(ctx, query).Scan(&nextTime); err != nil {
		return nil, errors.Wrap(err)
	}

	if !nextTime.Valid {
		// 재생성 예정인 레몬이 없음 - 현재 시간 반환
		return nil, nil
	}

	return &nextTime.Time, nil
}

func (s *LemonStore) DailyHarvestStats(ctx context.Context, userID string, days int) ([]*lemon.DailyHarvest, error) {
	query := `
		SELECT 
			DATE(created_at) as harvest_date,
			COALESCE(SUM(amount), 0) as total_amount
		FROM user_lemon_transactions 
		WHERE user_id = $1 
			AND action_type = 'harvest' 
			AND status = 'successful'
			AND created_at >= CURRENT_DATE - INTERVAL '%d days'
		GROUP BY DATE(created_at)
		ORDER BY harvest_date DESC
	`

	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(query, days), userID)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer rows.Close()

	var dailyHarvests []*lemon.DailyHarvest
	for rows.Next() {
		var harvestDate time.Time
		var amount int

		err := rows.Scan(&harvestDate, &amount)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		dailyHarvests = append(dailyHarvests, &lemon.DailyHarvest{
			Date:   harvestDate,
			Amount: amount,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err)
	}

	return dailyHarvests, nil
}

func (s *LemonStore) UserTransactionCount(ctx context.Context, userID string, instanceName *string) (int, error) {
	var query string
	var args []interface{}

	if instanceName != nil {
		query = `
			SELECT COUNT(*) 
			FROM user_lemon_transactions t
			LEFT JOIN db_instances i ON t.db_instance_id = i.id
			WHERE t.user_id = $1::uuid AND i.name = $2
		`
		args = []interface{}{userID, *instanceName}
	} else {
		query = `SELECT COUNT(*) FROM user_lemon_transactions WHERE user_id = $1::uuid`
		args = []interface{}{userID}
	}

	var count int
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err)
	}

	return count, nil
}

func (s *LemonStore) UserTransactionsWithInstance(ctx context.Context, userID string, instanceName *string, limit, offset int) ([]*lemon.TransactionWithInstance, error) {
	baseQuery := `
		SELECT 
			t.id, i.name as instance_name,
			t.action_type, t.status, t.amount, t.balance, t.created_at, t.note
		FROM user_lemon_transactions t
		LEFT JOIN db_instances i ON t.db_instance_id = i.id
		WHERE t.user_id = $1
	`

	var args []interface{}
	args = append(args, userID)

	wb := &whereBuilder{argIndex: 1}
	if instanceName != nil {
		wb.add("i.name = $%d", *instanceName)
	}

	query := baseQuery + wb.build() + " ORDER BY t.created_at DESC"
	args = append(args, wb.args...)

	query, args = addPagination(query, args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer rows.Close()

	var transactions []*lemon.TransactionWithInstance
	for rows.Next() {
		var tx lemon.TransactionWithInstance
		var instanceName sql.NullString
		var note sql.NullString

		err := rows.Scan(
			&tx.ID,
			&instanceName,
			&tx.ActionType,
			&tx.Status,
			&tx.Amount,
			&tx.Balance,
			&tx.CreatedAt,
			&note,
		)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		if instanceName.Valid {
			tx.InstanceName = &instanceName.String
		}

		if note.Valid {
			tx.Note = note.String
		} else {
			tx.Note = ""
		}

		transactions = append(transactions, &tx)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err)
	}

	return transactions, nil
}
