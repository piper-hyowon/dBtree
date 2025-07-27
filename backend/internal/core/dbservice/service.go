package dbservice

import (
	"context"
	"github.com/google/uuid"
)

type Service interface {
	CreateInstance(ctx context.Context, userID uuid.UUID, req *CreateInstanceRequest) (*DBInstance, error)
	GetInstance(ctx context.Context, id uuid.UUID) (*DBInstance, error)
	ListInstances(ctx context.Context, userID uuid.UUID, filters map[string]interface{}) ([]*DBInstance, error)
	UpdateInstance(ctx context.Context, id uuid.UUID, req *UpdateInstanceRequest) (*DBInstance, error)
	DeleteInstance(ctx context.Context, id uuid.UUID) error

	StartInstance(ctx context.Context, id uuid.UUID) error
	StopInstance(ctx context.Context, id uuid.UUID) error
	RestartInstance(ctx context.Context, id uuid.UUID) error

	CreateBackup(ctx context.Context, id uuid.UUID, description string) (string, error)
	ListBackups(ctx context.Context, id uuid.UUID) ([]map[string]interface{}, error)
	RestoreFromBackup(ctx context.Context, id uuid.UUID, backupID string) error
}
