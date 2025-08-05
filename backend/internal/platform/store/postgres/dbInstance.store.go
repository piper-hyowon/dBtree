package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
)

const (
	instanceColumns = `
        id, external_id, user_id, name, type, size, mode,
        created_from_preset,
        cpu, memory, disk,
        creation_cost, hourly_cost,
        status, status_reason,
        k8s_namespace, k8s_resource_name,
        endpoint, port,
        config,
        backup_enabled, backup_schedule, backup_retention_days,
        created_at, updated_at, last_billed_at, paused_at, deleted_at
    `

	selectInstancesQuery = "SELECT " + instanceColumns + " FROM db_instances"
)

type DBInstanceStore struct {
	db *sql.DB
}

var _ dbservice.DBInstanceStore = (*DBInstanceStore)(nil)

func NewDBInstanceStore(db *sql.DB) dbservice.DBInstanceStore {
	return &DBInstanceStore{db: db}
}

func (s *DBInstanceStore) Create(ctx context.Context, instance *dbservice.DBInstance) error {
	return withTx(ctx, s.db, func(ctx context.Context, tx *sql.Tx) error {
		query := `
            INSERT INTO db_instances (
                external_id, user_id, name, type, size, mode,
                created_from_preset,
                cpu, memory, disk,
                creation_cost, hourly_cost,
                status, config,
                backup_enabled, backup_schedule, backup_retention_days
            ) VALUES (
                $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
                $11, $12, $13, $14, $15, $16, $17
            ) RETURNING id, created_at, updated_at
        `

		configJSON, err := json.Marshal(instance.Config)
		if err != nil {
			return fmt.Errorf("marshal config: %w", err)
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
			instance.Status,
			configJSON,
			instance.BackupConfig.Enabled,
			toNullString(instance.BackupConfig.Schedule),
			toNullInt32(instance.BackupConfig.RetentionDays),
		).Scan(&instance.ID, &instance.CreatedAt, &instance.UpdatedAt)

		if err != nil {
			if isUniqueViolation(err, "unique_user_instance_name") {
				return errors.NewInstanceNameConflictError(instance.Name)
			}
			return fmt.Errorf("insert instance: %w", err)
		}

		return nil
	})
}

func (s *DBInstanceStore) Find(ctx context.Context, externalID string) (*dbservice.DBInstance, error) {
	query := selectInstancesQuery + " WHERE external_id = $1 AND deleted_at IS NULL"

	row := s.db.QueryRowContext(ctx, query, externalID)
	instance, err := scanInstance(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find instance: %w", err)
	}

	return instance, nil
}

func (s *DBInstanceStore) FindByUserAndName(ctx context.Context, userID, name string) (*dbservice.DBInstance, error) {
	query := selectInstancesQuery + " WHERE user_id = $1 AND name = $2 AND deleted_at IS NULL"

	row := s.db.QueryRowContext(ctx, query, userID, name)
	instance, err := scanInstance(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find instance by name: %w", err)
	}

	return instance, nil
}

func (s *DBInstanceStore) List(ctx context.Context, userID string, filters dbservice.ListInstancesRequest) ([]*dbservice.DBInstance, error) {
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

	// WHERE 절 구성
	wb := &whereBuilder{
		conditions: []string{},
		args:       []interface{}{userID},
		argIndex:   1,
	}

	if filters.Status != nil {
		wb.add("status = $%d", *filters.Status)
	}
	if filters.Type != nil {
		wb.add("type = $%d", *filters.Type)
	}
	if filters.NameLike != "" {
		wb.add("name ILIKE $%d", "%"+filters.NameLike+"%")
	}

	query += wb.build()
	query += " ORDER BY created_at DESC"

	// 페이지네이션
	limit := normalizeLimit(filters.Limit)
	offset := calculateOffset(filters.Page, limit)

	query, args := addPagination(query, wb.args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}
	defer rows.Close()

	instances := make([]*dbservice.DBInstance, 0, limit)

	for rows.Next() {
		instance, err := scanInstanceList(rows)
		if err != nil {
			return nil, fmt.Errorf("scan instance: %w", err)
		}
		instances = append(instances, instance)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return instances, nil
}

func (s *DBInstanceStore) ListRunning(ctx context.Context) ([]*dbservice.DBInstance, error) {
	query := selectInstancesQuery + ` 
        WHERE status = $1 AND deleted_at IS NULL
        ORDER BY created_at DESC
    `

	return s.queryInstances(ctx, query, dbservice.StatusRunning)
}

func (s *DBInstanceStore) ListPausedBefore(ctx context.Context, before time.Time) ([]*dbservice.DBInstance, error) {
	query := selectInstancesQuery + ` 
        WHERE status = $1 AND paused_at < $2 AND deleted_at IS NULL
        ORDER BY paused_at ASC
    `

	return s.queryInstances(ctx, query, dbservice.StatusPaused, before)
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
		return fmt.Errorf("update instance: %w", err)
	}

	return checkRowsAffected(result, "instance", fmt.Sprintf("%d", instance.ID))
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
		return fmt.Errorf("update status: %w", err)
	}

	return checkRowsAffected(result, "instance", fmt.Sprintf("%d", id))
}

func (s *DBInstanceStore) UpdateBillingTime(ctx context.Context, id int64, billedAt time.Time) error {
	query := `
        UPDATE db_instances SET
            last_billed_at = $2,
            updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `

	result, err := s.db.ExecContext(ctx, query, id, billedAt)
	if err != nil {
		return fmt.Errorf("update billing time: %w", err)
	}

	return checkRowsAffected(result, "instance", fmt.Sprintf("%d", id))
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
		return fmt.Errorf("delete instance: %w", err)
	}

	return checkRowsAffected(result, "instance", externalID)
}

func (s *DBInstanceStore) CreateBackup(ctx context.Context, backup *dbservice.BackupRecord) error {
	query := `
        INSERT INTO db_instance_backups (
            instance_id, external_id, name, type, status,
            k8s_job_name
        ) VALUES (
            $1, $2, $3, $4, $5, $6
        ) RETURNING id, created_at
    `

	err := s.db.QueryRowContext(ctx, query,
		backup.InstanceID,
		backup.ExternalID,
		backup.Name,
		backup.Type,
		backup.Status,
		backup.K8sJobName,
	).Scan(&backup.ID, &backup.CreatedAt)

	if err != nil {
		return fmt.Errorf("create backup: %w", err)
	}

	return nil
}

func (s *DBInstanceStore) FindBackup(ctx context.Context, backupID string) (*dbservice.BackupRecord, error) {
	query := `
        SELECT
            id, instance_id, external_id, name, type, status,
            k8s_job_name, size_bytes, storage_path, error_message,
            created_at, completed_at, expires_at
        FROM db_instance_backups
        WHERE external_id = $1
    `

	row := s.db.QueryRowContext(ctx, query, backupID)
	backup, err := scanBackup(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find backup: %w", err)
	}

	return backup, nil
}

func (s *DBInstanceStore) ListBackups(ctx context.Context, instanceID string) ([]*dbservice.BackupRecord, error) {
	query := `
        SELECT
            id, instance_id, external_id, name, type, status,
            k8s_job_name, size_bytes, storage_path, error_message,
            created_at, completed_at, expires_at
        FROM db_instance_backups
        WHERE instance_id IN (
            SELECT id FROM db_instances WHERE external_id = $1
        )
        ORDER BY created_at DESC
    `

	rows, err := s.db.QueryContext(ctx, query, instanceID)
	if err != nil {
		return nil, fmt.Errorf("list backups: %w", err)
	}
	defer rows.Close()

	backups := make([]*dbservice.BackupRecord, 0, 10)

	for rows.Next() {
		backup, err := scanBackup(rows)
		if err != nil {
			return nil, fmt.Errorf("scan backup: %w", err)
		}
		backups = append(backups, backup)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return backups, nil
}

func (s *DBInstanceStore) UpdateBackupStatus(ctx context.Context, backupID string, status dbservice.BackupStatus, errorMsg string) error {
	query := `
        UPDATE db_instance_backups
        SET status = $2, updated_at = NOW()
    `

	args := []interface{}{backupID, status}

	if errorMsg != "" {
		query += ", error_message = $3"
		args = append(args, errorMsg)
	}

	if status == dbservice.BackupStatusCompleted {
		query += ", completed_at = NOW()"
	}

	query += " WHERE external_id = $1"

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update backup status: %w", err)
	}

	return checkRowsAffected(result, "backup", backupID)
}

func (s *DBInstanceStore) queryInstances(ctx context.Context, query string, args ...interface{}) ([]*dbservice.DBInstance, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*dbservice.DBInstance
	for rows.Next() {
		instance, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return instances, nil
}
