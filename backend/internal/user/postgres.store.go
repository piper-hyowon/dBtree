package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/piper-hyowon/dBtree/internal/common"
	"github.com/piper-hyowon/dBtree/internal/common/user"
	"time"
)

type PostgresStore struct {
	db *sql.DB
}

var _ user.Store = (*PostgresStore)(nil)

func NewPostgresStore(db *sql.DB) user.Store {
	return &PostgresStore{
		db: db,
	}
}

func (s *PostgresStore) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	if email == "" {
		return nil, errors.New("empty Email")
	}

	query := `SELECT id, email, created_at, updated_at 
              FROM users 
              WHERE email = $1 AND is_deleted = FALSE`

	var u user.User
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (s *PostgresStore) FindById(ctx context.Context, id string) (*user.User, error) {
	if id == "" {
		return nil, errors.New("empty ID")
	}

	query := `SELECT id, email, created_at, updated_at FROM users WHERE id = $1`

	var u user.User
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (s *PostgresStore) Create(ctx context.Context, email string) error {
	if email == "" {
		return errors.New("empty Email")
	}

	exists, err := s.emailExists(ctx, email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("duplicated email")
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	query := `
		INSERT INTO users (id, email, created_at, updated_at) 
		VALUES ($1, $2, $3, $4)
	`

	_, err = s.db.ExecContext(ctx, query, id, email, now, now)
	return err
}

func (s *PostgresStore) emailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *PostgresStore) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("empty ID")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var email string
	err = tx.QueryRowContext(ctx, `SELECT email FROM users WHERE id = $1 AND is_deleted = FALSE`, id).Scan(&email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrUserNotFound
		}
		return err
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
