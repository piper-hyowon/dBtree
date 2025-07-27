package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"strings"
)

type DBInstanceStore struct {
	db *sql.DB
}

var _ dbservice.DBInstanceStore = (*DBInstanceStore)(nil)

func NewDBInstanceStore(db *sql.DB) dbservice.DBInstanceStore {
	return &DBInstanceStore{
		db: db,
	}
}

func (s *DBInstanceStore) Create(ctx context.Context, instance *dbservice.DBInstance) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err)
	}
	defer tx.Rollback()

	query := `
        INSERT INTO db_instances (
            external_id, user_id, name, type, size, mode,
            created_from_preset, 
            cpu, memory, disk,
            creation_cost, hourly_cost, minimum_lemons,
            status, config,
            backup_enabled, backup_schedule, backup_retention_days
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
            $11, $12, $13, $14, $15, $16, $17, $18
        ) RETURNING id, created_at, updated_at
    `

	configJSON, err := json.Marshal(instance.Config)
	if err != nil {
		return errors.Wrap(err)
	}

	var backupSchedule sql.NullString
	var backupRetentionDays sql.NullInt32

	if instance.BackupConfig.Schedule != "" {
		backupSchedule = sql.NullString{String: instance.BackupConfig.Schedule, Valid: true}
	}
	if instance.BackupConfig.RetentionDays > 0 {
		backupRetentionDays = sql.NullInt32{Int32: int32(instance.BackupConfig.RetentionDays), Valid: true}
	}

	err = tx.QueryRowContext(ctx, query,
		instance.ExternalID,
		instance.UserID,
		instance.Name,
		instance.Type,
		instance.Size,
		instance.Mode,
		instance.CreatedFromPreset,
		instance.Resources.CPU,
		instance.Resources.Memory,
		instance.Resources.Disk,
		instance.Cost.CreationCost,
		instance.Cost.HourlyLemons,
		instance.Cost.MinimumLemons,
		instance.Status,
		configJSON,
		instance.BackupConfig.Enabled,
		backupSchedule,
		backupRetentionDays,
	).Scan(&instance.ID, &instance.CreatedAt, &instance.UpdatedAt)

	if err != nil {
		// 이름 중복 체크
		if strings.Contains(err.Error(), "unique_user_instance_name") {
			return errors.NewInstanceNameConflictError(instance.Name)
		}
		return errors.Wrap(err)
	}

	return tx.Commit()
}

func (s *DBInstanceStore) Detail(ctx context.Context, externalID string) (*dbservice.DBInstance, error) {
	query := `
        SELECT 
            id, external_id, user_id, name, type, size, mode,
            created_from_preset,
            cpu, memory, disk,
            creation_cost, hourly_cost, minimum_lemons,
            status, status_reason,
            k8s_namespace, k8s_resource_name,
            endpoint, port,
            config,
            backup_enabled, backup_schedule, backup_retention_days,
            created_at, updated_at, last_billed_at, paused_at, deleted_at
        FROM db_instances 
        WHERE external_id = $1 AND deleted_at IS NULL
    `

	var instance dbservice.DBInstance
	var statusReason sql.NullString
	var createdFromPreset sql.NullString
	var k8sNamespace, k8sResourceName sql.NullString
	var endpoint sql.NullString
	var port sql.NullInt32
	var configJSON []byte
	var backupSchedule sql.NullString
	var backupRetentionDays sql.NullInt32
	var lastBilledAt, pausedAt, deletedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, externalID).Scan(
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
		&instance.Cost.MinimumLemons,
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err)
	}

	// Nullable 값 처리
	if createdFromPreset.Valid {
		instance.CreatedFromPreset = &createdFromPreset.String
	}
	if statusReason.Valid {
		instance.StatusReason = statusReason.String
	}
	if k8sNamespace.Valid {
		instance.K8sNamespace = k8sNamespace.String
	}
	if k8sResourceName.Valid {
		instance.K8sResourceName = k8sResourceName.String
	}
	if endpoint.Valid {
		instance.Endpoint = endpoint.String
	}
	if port.Valid {
		instance.Port = int(port.Int32)
	}
	if backupSchedule.Valid {
		instance.BackupConfig.Schedule = backupSchedule.String
	}
	if backupRetentionDays.Valid {
		instance.BackupConfig.RetentionDays = int(backupRetentionDays.Int32)
	}
	if lastBilledAt.Valid {
		instance.LastBilledAt = &lastBilledAt.Time
	}
	if pausedAt.Valid {
		instance.PausedAt = &pausedAt.Time
	}
	if deletedAt.Valid {
		instance.DeletedAt = &deletedAt.Time
	}

	// JSONB 파싱
	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &instance.Config); err != nil {
			return nil, errors.Wrap(err)
		}
	} else {
		instance.Config = make(map[string]interface{})
	}

	return &instance, nil
}

func (s *DBInstanceStore) List(ctx context.Context, userID string, filters map[string]interface{}) ([]*dbservice.DBInstance, error) {
	query := `
        SELECT 
            id, external_id, user_id, name, type, size, mode,
            cpu, memory, disk,
            hourly_cost,
            status, status_reason,
            endpoint, port,
            created_at, updated_at
        FROM db_instances 
        WHERE user_id = $1 AND deleted_at IS NULL
    `

	args := []interface{}{userID}
	argCount := 1

	// 필터 적용
	if status, ok := filters["status"].(string); ok && status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	if dbType, ok := filters["type"].(string); ok && dbType != "" {
		argCount++
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, dbType)
	}

	if name, ok := filters["name"].(string); ok && name != "" {
		argCount++
		query += fmt.Sprintf(" AND name ILIKE $%d", argCount)
		args = append(args, "%"+name+"%")
	}

	query += " ORDER BY created_at DESC"

	// 페이징
	if limit, ok := filters["limit"].(int); ok && limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)

		if offset, ok := filters["offset"].(int); ok && offset > 0 {
			argCount++
			query += fmt.Sprintf(" OFFSET $%d", argCount)
			args = append(args, offset)
		}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer rows.Close()

	var instances []*dbservice.DBInstance
	for rows.Next() {
		var i dbservice.DBInstance
		var statusReason sql.NullString
		var endpoint sql.NullString
		var port sql.NullInt32

		err := rows.Scan(
			&i.ID,
			&i.ExternalID,
			&i.UserID,
			&i.Name,
			&i.Type,
			&i.Size,
			&i.Mode,
			&i.Resources.CPU,
			&i.Resources.Memory,
			&i.Resources.Disk,
			&i.Cost.HourlyLemons,
			&i.Status,
			&statusReason,
			&endpoint,
			&port,
			&i.CreatedAt,
			&i.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		if statusReason.Valid {
			i.StatusReason = statusReason.String
		}
		if endpoint.Valid {
			i.Endpoint = endpoint.String
		}
		if port.Valid {
			i.Port = int(port.Int32)
		}

		instances = append(instances, &i)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err)
	}

	return instances, nil
}

func (s *DBInstanceStore) Update(ctx context.Context, instance *dbservice.DBInstance) error {
	query := `
        UPDATE db_instances SET
            k8s_namespace = $2,
            k8s_resource_name = $3,
            endpoint = $4,
            port = $5,
            status = $6,
            status_reason = $7,
            last_billed_at = $8,
            paused_at = $9,
            updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `

	result, err := s.db.ExecContext(ctx, query,
		instance.ID,
		instance.K8sNamespace,
		instance.K8sResourceName,
		instance.Endpoint,
		instance.Port,
		instance.Status,
		instance.StatusReason,
		instance.LastBilledAt,
		instance.PausedAt,
	)

	if err != nil {
		return errors.Wrap(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err)
	}

	if rows == 0 {
		return errors.NewResourceNotFoundError("db_instance", fmt.Sprintf("%d", instance.ID))
	}

	return nil
}

func (s *DBInstanceStore) UpdateStatus(ctx context.Context, id int64, status dbservice.InstanceStatus, reason string) error {
	query := `
        UPDATE db_instances SET
            status = $2,
            status_reason = $3,
            updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `

	result, err := s.db.ExecContext(ctx, query, id, status, reason)
	if err != nil {
		return errors.Wrap(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err)
	}

	if rows == 0 {
		return errors.NewResourceNotFoundError("db_instance", fmt.Sprintf("%d", id))
	}

	return nil
}

func (s *DBInstanceStore) Delete(ctx context.Context, externalID string) error {
	query := `
        UPDATE db_instances SET
            deleted_at = NOW(),
            updated_at = NOW()
        WHERE external_id = $1 AND deleted_at IS NULL
    `

	result, err := s.db.ExecContext(ctx, query, externalID)
	if err != nil {
		return errors.Wrap(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err)
	}

	if rows == 0 {
		return errors.NewResourceNotFoundError("db_instance", externalID)
	}

	return nil
}
