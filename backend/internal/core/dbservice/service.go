package dbservice

import (
	"context"
)

type Service interface {
	// CRUD
	CreateInstance(ctx context.Context, userID string, userLemon int, req *CreateInstanceRequest) (*CreateInstanceResponse, error)
	ListInstances(ctx context.Context, userID string) ([]*DBInstance, error)
	UpdateInstance(ctx context.Context, userID, instanceID string, req *UpdateInstanceRequest) (*DBInstance, error)
	DeleteInstance(ctx context.Context, userID, instanceID string) error

	// Control

	StartInstance(ctx context.Context, userID, instanceID string) error
	StopInstance(ctx context.Context, userID, instanceID string) error
	RestartInstance(ctx context.Context, userID, instanceID string) error

	// Status Sync

	GetInstanceWithSync(ctx context.Context, userID, instanceID string) (*DBInstance, error)

	// Backup

	CreateBackup(ctx context.Context, userID, instanceID string, name string) (*BackupRecord, error)
	ListBackups(ctx context.Context, userID, instanceID string) ([]*BackupRecord, error)
	RestoreFromBackup(ctx context.Context, userID, instanceID string, backupID string) error

	// Metrics

	InstanceMetrics(ctx context.Context, instanceID string) (*InstanceMetrics, error)

	// Presets & Cost

	ListPresets(ctx context.Context) ([]*DBPreset, error)
	EstimateCost(ctx context.Context, req *EstimateCostRequest) (*EstimateCostResponse, error)
}
