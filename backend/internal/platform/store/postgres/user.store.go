package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/piper-hyowon/dBtree/internal/core/errors"

	"github.com/piper-hyowon/dBtree/internal/core/user"
	"time"
)

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) TotalUserCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users`,
	).Scan(&count)
	return count, err
}

func (s *UserStore) TopLemonHolders(ctx context.Context, limit int) ([]*user.User, error) {
	query := `
        SELECT 
            id, 
            email, 
            lemon_balance, 
            last_harvest_at, 
            created_at, 
            updated_at
        FROM users 
        WHERE is_deleted = false 
        ORDER BY lemon_balance DESC 
        LIMIT $1
    `

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		usr := &user.User{}
		var lastHarvest sql.NullTime

		err := rows.Scan(
			&usr.ID,
			&usr.Email,
			&usr.LemonBalance,
			&lastHarvest,
			&usr.CreatedAt,
			&usr.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		// NULL 처리
		if lastHarvest.Valid {
			usr.LastHarvestAt = &lastHarvest.Time
		}

		users = append(users, usr)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err)
	}

	return users, nil
}

var _ user.Store = (*UserStore)(nil)

func NewUserStore(db *sql.DB) user.Store {
	return &UserStore{
		db: db,
	}
}

func (s *UserStore) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `SELECT id, email, lemon_balance, last_harvest_at, created_at, updated_at,
       total_earned_lemons, total_spent_lemons
              FROM users 
              WHERE email = $1 AND is_deleted = FALSE`

	var u user.User
	var lastHarvest sql.NullTime
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.LemonBalance,
		&lastHarvest,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.TotalEarnedLemons,
		&u.TotalSpentLemons,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // 유저가 없는 경우 에러 반환 X
		}
		return nil, errors.Wrap(err)
	}

	if lastHarvest.Valid {
		u.LastHarvestAt = &lastHarvest.Time
	} else {
		u.LastHarvestAt = nil
	}

	return &u, nil
}

func (s *UserStore) FindById(ctx context.Context, id string) (*user.User, error) {
	query := `SELECT id, email, lemon_balance, last_harvest_at, created_at, updated_at FROM users WHERE id = $1`

	var u user.User
	var lastHarvest sql.NullTime
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.LemonBalance,
		&lastHarvest,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err)
	}

	if lastHarvest.Valid {
		u.LastHarvestAt = &lastHarvest.Time
	} else {
		u.LastHarvestAt = nil
	}

	return &u, nil
}

func (s *UserStore) CreateIfNotExists(ctx context.Context, email string) (bool, error) {
	exists, err := s.emailExists(ctx, email)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	query := `
		INSERT INTO users (id, email, created_at, updated_at) 
		VALUES ($1, $2, $3, $4)
	`

	_, err = s.db.ExecContext(ctx, query, id, email, now, now)
	if err != nil {
		return false, errors.Wrap(err)
	}

	return true, nil
}

func (s *UserStore) emailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err)

	}

	return exists, nil
}

func (s *UserStore) Delete(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("트랜잭션 시작 실패: %w", err)
	}
	defer tx.Rollback()

	var email string
	err = tx.QueryRowContext(ctx, `SELECT email FROM users WHERE id = $1 AND is_deleted = FALSE`, id).Scan(&email)
	if err != nil {
		// row 가 반드시 있어야함
		return errors.Wrap(err)
	}

	now := time.Now().UTC()
	timestamp := fmt.Sprintf("%d", now.Unix())

	// 사용자 소프트 삭제 & 이메일 변경
	// 이메일-> deleted_[타임스탬프]_[원본이메일]
	newEmail := fmt.Sprintf("deleted_%s_%s", timestamp, email)

	_, err = tx.ExecContext(ctx, `
        UPDATE users 
        SET is_deleted = TRUE, email = $1, updated_at = $2
        WHERE id = $3 AND is_deleted = FALSE
    `, newEmail, now, id)

	if err != nil {
		return err
	}

	return tx.Commit()
}
