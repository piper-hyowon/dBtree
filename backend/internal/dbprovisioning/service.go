package dbservice

import (
	"context"
	"github.com/piper-hyowon/dBtree/internal/core/dbprovisioning"
	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
)

type service struct {
}

var _ dbprovisioning.Service = (*service)(nil)

func NewService() dbprovisioning.Service {
	return nil
}

func (s *service) ProvisionInstance(ctx context.Context, instance *dbservice.DatabaseInstance) error {
	return nil
}
func (s *service) DeleteInstance(ctx context.Context, instanceID string) error {
	return nil

}

func (s *service) InstanceStatus(ctx context.Context, instanceID string) (*dbservice.DatabaseInstance, string, error) {
	return nil, "", nil
}

func (s *service) ConnectionInfo(ctx context.Context, instanceID string) (string, int, string, error) {
	return "", 0, "", nil
}
