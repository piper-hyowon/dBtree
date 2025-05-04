package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"log"
	"runtime/debug"
	"time"

	_ "github.com/lib/pq"
	"github.com/piper-hyowon/dBtree/internal/platform/config"
)

type Client interface {
	Close() error
	DB() *sql.DB
}

type client struct {
	db       *sql.DB
	config   config.PostgresConfig
	migrator *Migrator
	logger   *log.Logger
}

var _ Client = (*client)(nil)

func NewClient(config config.PostgresConfig, logger *log.Logger) (Client, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.DBName, config.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, errors.NewInternalErrorWithStack(fmt.Errorf("PostgreSQL 연결 실패: %w", err), string(debug.Stack()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, errors.NewInternalErrorWithStack(fmt.Errorf("PostgreSQL 핑 실패: %w", err), string(debug.Stack()))
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetimeMinutes) * time.Minute)

	client := &client{
		db:       db,
		config:   config,
		migrator: NewMigrator(db, logger),
		logger:   logger,
	}

	if err = client.runMigrations(); err != nil {
		db.Close()
		return nil, err
	}

	return client, nil
}

func (c *client) DB() *sql.DB {
	return c.db
}

func (c *client) Close() error {
	return c.db.Close()
}

func (c *client) runMigrations() error {
	c.logger.Println("데이터베이스 마이그레이션 시작...")
	return c.migrator.RunMigrations()
}
