package dbservice

import (
	"context"
	"time"
)

type DBInstanceStore interface {
	// Instance CRUD

	Create(ctx context.Context, instance *DBInstance) error
	Detail(ctx context.Context, externalID string) (*DBInstance, error)
	DetailByUserAndName(ctx context.Context, userID string, name string) (*DBInstance, error)
	List(ctx context.Context, userID string, filters ListInstancesRequest) ([]*DBInstance, error)
	Update(ctx context.Context, instance *DBInstance) error
	UpdateStatus(ctx context.Context, id int64, status InstanceStatus, reason string) error
	Delete(ctx context.Context, externalID string) error

	// Billing & Scheduler

	FindRunningInstances(ctx context.Context) ([]*DBInstance, error)
	FindPausedInstancesBefore(ctx context.Context, before time.Time) ([]*DBInstance, error)
	UpdateBillingTime(ctx context.Context, id int64, billedAt time.Time) error

	// Backup

	CreateBackupRecord(ctx context.Context, backup *BackupRecord) error
	ListBackupRecords(ctx context.Context, instanceID string) ([]*BackupRecord, error)
	DetailBackupByID(ctx context.Context, backupID string) (*BackupRecord, error)
	UpdateBackupStatus(ctx context.Context, backupID string, status BackupStatus, errorMsg string) error
}

type PresetStore interface {
	Detail(ctx context.Context, id string) (*DBPreset, error)
	ListByType(ctx context.Context, dbType DBType) ([]*DBPreset, error)
}
