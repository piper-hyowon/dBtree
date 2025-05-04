package dbprovisioning

import (
	"context"
	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
)

type Service interface {
	ProvisionMongoDB(ctx context.Context, instance *dbservice.DBInstance) error
	UpdateMongoDB(ctx context.Context, instance *dbservice.DBInstance) error

	ProvisionRedis(ctx context.Context, instance *dbservice.DBInstance) error
	UpdateRedis(ctx context.Context, instance *dbservice.DBInstance) error

	DeleteInstance(ctx context.Context, instanceID string) error
	StartInstance(ctx context.Context, instance *dbservice.DBInstance) error
	StopInstance(ctx context.Context, instance *dbservice.DBInstance) error
	RestartInstance(ctx context.Context, instance *dbservice.DBInstance) error

	InstanceStatus(ctx context.Context, instanceID string) (*dbservice.DBInstance, string, error)
	ConnectionInfo(ctx context.Context, instanceID string) (string, int, string, error)

	BackupInstance(ctx context.Context, instance *dbservice.DBInstance) (string, error)
	RestoreInstance(ctx context.Context, instance *dbservice.DBInstance, backupID string) error
}
