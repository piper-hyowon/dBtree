package dbprovisioning

import (
	"context"
	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
)

type Service interface {
	ProvisionInstance(ctx context.Context, instance *dbservice.DatabaseInstance) error
	DeleteInstance(ctx context.Context, instanceID string) error
	InstanceStatus(ctx context.Context, instanceID string) (*dbservice.DatabaseInstance, string, error)
	ConnectionInfo(ctx context.Context, instanceID string) (string, int, string, error)
}
