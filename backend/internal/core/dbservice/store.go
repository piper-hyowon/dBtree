package dbservice

import (
	"context"
)

type DBInstanceStore interface {
	Create(ctx context.Context, instance *DBInstance) error
	Detail(ctx context.Context, externalID string) (*DBInstance, error)
	List(ctx context.Context, userID string, filters map[string]interface{}) ([]*DBInstance, error)
	Update(ctx context.Context, instance *DBInstance) error
	UpdateStatus(ctx context.Context, id int64, status InstanceStatus, reason string) error
	Delete(ctx context.Context, id string) error

	CreateBackupRecord(ctx context.Context, backup *BackupRecord) error
	ListBackupRecords(ctx context.Context, instanceID string) ([]*BackupRecord, error)
}

type PresetStore interface {
	List(ctx context.Context) ([]*DBPreset, error)
}
