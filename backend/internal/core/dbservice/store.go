package dbservice

import (
	"context"
	"time"
)

type DBInstanceStore interface {
	Create(ctx context.Context, instance *DBInstance) error
	Find(ctx context.Context, externalID string) (*DBInstance, error)
	FindByUserAndName(ctx context.Context, userID string, name string) (*DBInstance, error)
	List(ctx context.Context, userID string) ([]*DBInstance, error)
	ListRunning(ctx context.Context) ([]*DBInstance, error)
	ListPausedBefore(ctx context.Context, before time.Time) ([]*DBInstance, error)
	Update(ctx context.Context, instance *DBInstance) error
	UpdateStatus(ctx context.Context, id int64, status InstanceStatus, reason string) error
	UpdateBillingTime(ctx context.Context, id int64, billedAt time.Time) error
	Delete(ctx context.Context, externalID string) error

	CountActive(ctx context.Context, userID string) (int, error)

	CreateBackup(ctx context.Context, backup *BackupRecord) error
	FindBackup(ctx context.Context, backupID string) (*BackupRecord, error)
	ListBackups(ctx context.Context, instanceID string) ([]*BackupRecord, error)
	UpdateBackupStatus(ctx context.Context, backupID string, status BackupStatus, errorMsg string) error
}

type PresetStore interface {
	Find(ctx context.Context, id string) (*DBPreset, error)
	ListByType(ctx context.Context, dbType DBType) ([]*DBPreset, error)
}

type PortStore interface {
	AllocatePort(ctx context.Context, instanceID string) (int, error)
	ReleasePort(ctx context.Context, instanceID string) error
	GetPort(ctx context.Context, instanceID string) (int, error)
}
