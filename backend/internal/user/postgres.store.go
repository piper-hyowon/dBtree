package user

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/piper-hyowon/dBtree/internal/common"
	"time"
)

type PostgresStore struct {
	db *sql.DB
}

var _ Store = (*PostgresStore)(nil)

func NewPostgresStore(db *sql.DB) Store {
	return &PostgresStore{
		db: db,
	}
}

func (s *PostgresStore) FindByEmail(ctx context.Context, email string) (*User, error) {
	if email == "" {
		return nil, errors.New("empty Email")
	}

	query := `SELECT id, email, created_at, updated_at FROM users WHERE email = $1`

	var u User
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

func (s *PostgresStore) FindById(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, errors.New("empty ID")
	}

	query := `SELECT id, email, created_at, updated_at FROM users WHERE id = $1`

	var u User
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
