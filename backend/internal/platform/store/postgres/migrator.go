package postgres

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"io/fs"
	"log"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Migrator struct {
	db     *sql.DB
	logger *log.Logger
}

func NewMigrator(db *sql.DB, logger *log.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

// 현재는 테이블 존재 여부만 판단, TODO 스키마 변경 감지
func (m *Migrator) RunMigrations() error {
	_, err := m.db.Exec(`
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version TEXT PRIMARY KEY,
            applied_at TIMESTAMP NOT NULL DEFAULT NOW()
        )
    `)
	if err != nil {
		return errors.NewInternalErrorWithStack(fmt.Errorf("마이그레이션 테이블 생성 실패: %w", err), string(debug.Stack()))
	}

	rows, err := m.db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return errors.NewInternalErrorWithStack(fmt.Errorf("마이그레이션 목록 조회 실패: %w", err), string(debug.Stack()))
	}
	defer rows.Close()

	appliedMigrations := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return errors.NewInternalErrorWithStack(fmt.Errorf("마이그레이션 버전 스캔 실패: %w", err), string(debug.Stack()))
		}
		appliedMigrations[version] = true
	}

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return errors.NewInternalErrorWithStack(fmt.Errorf("마이그레이션 디렉토리 읽기 실패: %w", err), string(debug.Stack()))
	}

	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	for _, fileName := range migrationFiles {
		version := strings.TrimSuffix(fileName, ".sql")

		if appliedMigrations[version] {
			m.logger.Printf("마이그레이션 %s 이미 적용됨", version)
			continue
		}

		tx, err := m.db.Begin()
		if err != nil {
			return errors.NewInternalErrorWithStack(fmt.Errorf("트랜잭션 시작 실패: %w", err), string(debug.Stack()))
		}

		content, err := fs.ReadFile(migrationsFS, filepath.Join("migrations", fileName))
		if err != nil {
			tx.Rollback()
			return errors.NewInternalErrorWithStack(fmt.Errorf("마이그레이션 파일 읽기 실패: %w", err), string(debug.Stack()))
		}

		m.logger.Printf("마이그레이션 적용 중: %s", fileName)
		_, err = tx.Exec(string(content))
		if err != nil {
			tx.Rollback()

			return errors.NewInternalErrorWithStack(fmt.Errorf("마이그레이션 실행 실패: %w", err), string(debug.Stack()))
		}

		_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			tx.Rollback()
			return errors.NewInternalErrorWithStack(fmt.Errorf("마이그레이션 버전 기록 실패: %w", err), string(debug.Stack()))
		}

		if err := tx.Commit(); err != nil {
			return errors.NewInternalErrorWithStack(fmt.Errorf("마이그레이션 트랜잭션 커밋 실패: %w", err), string(debug.Stack()))
		}

		m.logger.Printf("마이그레이션 %s 적용 완료", fileName)
	}

	m.logger.Println("모든 마이그레이션 성공적으로 적용됨")
	return nil
}
