package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
)

type txFunc func(context.Context, *sql.Tx) error

func withTx(ctx context.Context, db *sql.DB, fn txFunc) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(ctx, tx); err != nil {
		return err
	}

	return tx.Commit()
}

func toNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

func toNullInt32(i int) sql.NullInt32 {
	return sql.NullInt32{
		Int32: int32(i),
		Valid: i > 0,
	}
}

func timePtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func scanInstance(scanner interface{ Scan(...interface{}) error }) (*dbservice.DBInstance, error) {
	var (
		instance            dbservice.DBInstance
		statusReason        sql.NullString
		createdFromPreset   sql.NullString
		k8sNamespace        sql.NullString
		k8sResourceName     sql.NullString
		endpoint            sql.NullString
		port                sql.NullInt32
		configJSON          []byte
		backupSchedule      sql.NullString
		backupRetentionDays sql.NullInt32
		lastBilledAt        sql.NullTime
		pausedAt            sql.NullTime
		deletedAt           sql.NullTime
	)

	err := scanner.Scan(
		&instance.ID,
		&instance.ExternalID,
		&instance.UserID,
		&instance.Name,
		&instance.Type,
		&instance.Size,
		&instance.Mode,
		&createdFromPreset,
		&instance.Resources.CPU,
		&instance.Resources.Memory,
		&instance.Resources.Disk,
		&instance.Cost.CreationCost,
		&instance.Cost.HourlyLemons,
		&instance.Status,
		&statusReason,
		&k8sNamespace,
		&k8sResourceName,
		&endpoint,
		&port,
		&configJSON,
		&instance.BackupConfig.Enabled,
		&backupSchedule,
		&backupRetentionDays,
		&instance.CreatedAt,
		&instance.UpdatedAt,
		&lastBilledAt,
		&pausedAt,
		&deletedAt,
	)

	if err != nil {
		return nil, err
	}

	// Nullable 필드 처리
	if createdFromPreset.Valid {
		instance.CreatedFromPreset = &createdFromPreset.String
	}
	instance.StatusReason = statusReason.String
	instance.K8sNamespace = k8sNamespace.String
	instance.K8sResourceName = k8sResourceName.String
	instance.Endpoint = endpoint.String
	instance.Port = int(port.Int32)
	instance.BackupConfig.Schedule = backupSchedule.String
	instance.BackupConfig.RetentionDays = int(backupRetentionDays.Int32)
	if lastBilledAt.Valid {
		instance.LastBilledAt = &lastBilledAt.Time
	}
	if pausedAt.Valid {
		instance.PausedAt = &pausedAt.Time
	}
	if deletedAt.Valid {
		instance.DeletedAt = &deletedAt.Time
	}

	// JSON 파싱
	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &instance.Config); err != nil {
			return nil, fmt.Errorf("unmarshal config: %w", err)
		}
	} else {
		instance.Config = make(map[string]interface{})
	}

	return &instance, nil
}

// scanInstanceList는 록 조회용 간단한 스캔을 수행합니다
func scanInstanceList(scanner interface{ Scan(...interface{}) error }) (*dbservice.DBInstance, error) {
	var (
		instance     dbservice.DBInstance
		statusReason sql.NullString
		endpoint     sql.NullString
		port         sql.NullInt32
	)

	err := scanner.Scan(
		&instance.ID,
		&instance.ExternalID,
		&instance.UserID,
		&instance.Name,
		&instance.Type,
		&instance.Size,
		&instance.Mode,
		&instance.Resources.CPU,
		&instance.Resources.Memory,
		&instance.Resources.Disk,
		&instance.Cost.HourlyLemons,
		&instance.Status,
		&statusReason,
		&endpoint,
		&port,
		&instance.CreatedAt,
		&instance.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	instance.StatusReason = statusReason.String
	instance.Endpoint = endpoint.String
	instance.Port = int(port.Int32)

	return &instance, nil
}

func scanBackup(scanner interface{ Scan(...interface{}) error }) (*dbservice.BackupRecord, error) {
	var (
		backup       dbservice.BackupRecord
		externalID   string
		sizeBytes    sql.NullInt64
		storagePath  sql.NullString
		errorMessage sql.NullString
		completedAt  sql.NullTime
		expiresAt    sql.NullTime
	)

	err := scanner.Scan(
		&backup.ID,
		&backup.InstanceID,
		&externalID,
		&backup.Name,
		&backup.Type,
		&backup.Status,
		&backup.K8sJobName,
		&sizeBytes,
		&storagePath,
		&errorMessage,
		&backup.CreatedAt,
		&completedAt,
		&expiresAt,
	)

	if err != nil {
		return nil, err
	}

	// UUID 파싱
	backup.ExternalID, err = uuid.Parse(externalID)
	if err != nil {
		return nil, fmt.Errorf("parse backup external id: %w", err)
	}

	backup.SizeBytes = sizeBytes.Int64
	backup.StoragePath = storagePath.String
	backup.ErrorMessage = errorMessage.String
	if completedAt.Valid {
		backup.CompletedAt = &completedAt.Time
	}
	if expiresAt.Valid {
		backup.ExpiresAt = &expiresAt.Time
	}

	return &backup, nil
}

func checkRowsAffected(result sql.Result, resourceType, resourceID string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.NewResourceNotFoundError(resourceType, resourceID)
	}

	return nil
}

func isUniqueViolation(err error, constraint string) bool {
	return err != nil && strings.Contains(err.Error(), constraint)
}

type whereBuilder struct {
	conditions []string
	args       []interface{}
	argIndex   int
}

func (wb *whereBuilder) add(format string, value interface{}) {
	wb.argIndex++
	wb.conditions = append(wb.conditions, fmt.Sprintf(format, wb.argIndex))
	wb.args = append(wb.args, value)
}

func (wb *whereBuilder) build() string {
	if len(wb.conditions) == 0 {
		return ""
	}
	return " AND " + strings.Join(wb.conditions, " AND ")
}

const (
	DefaultPageSize = 5
	MaxPageSize     = 100
	MinPageSize     = 1
	DefaultPage     = 0
)

func normalizeLimit(limit int) int {
	switch {
	case limit <= 0:
		return DefaultPageSize
	case limit > MaxPageSize:
		return MaxPageSize
	default:
		return limit
	}
}

func calculateOffset(page, limit int) int {
	if page < DefaultPage {
		page = DefaultPage
	}
	return page * limit
}

func addPagination(query string, args []interface{}, limit, offset int) (string, []interface{}) {
	argCount := len(args)

	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit)

	if offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	return query, args
}
