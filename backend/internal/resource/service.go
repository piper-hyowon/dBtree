package resource

import (
	"context"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"log"

	coredbservice "github.com/piper-hyowon/dBtree/internal/core/dbservice"
	coreresource "github.com/piper-hyowon/dBtree/internal/core/resource"
)

type manager struct {
	dbiStore coredbservice.DBInstanceStore
	logger   *log.Logger
}

var _ coreresource.Manager = (*manager)(nil)

func NewManager(
	dbiStore coredbservice.DBInstanceStore,
	logger *log.Logger,
) coreresource.Manager {
	return &manager{
		dbiStore: dbiStore,
		logger:   logger,
	}
}

func (m *manager) CanAllocate(ctx context.Context, resources coreresource.SystemResourceSpec) (bool, string, error) {
	available, err := m.GetAvailable(ctx)
	if err != nil {
		return false, "리소스 확인 실패", errors.Wrapf(err, "리소스 확인 실패:")
	}

	if resources.CPU > available.CPU {
		return false, fmt.Sprintf("CPU 리소스 부족 (가용: %.2f vCPU, 요청: %.2f vCPU)",
			available.CPU, resources.CPU), nil
	}

	if resources.Memory > available.Memory {
		return false, fmt.Sprintf("메모리 부족 (가용: %dMB, 요청: %dMB)",
			available.Memory, resources.Memory), nil
	}

	return true, "", nil
}

func (m *manager) GetStatus(ctx context.Context) (*coreresource.SystemResourceStatus, error) {
	// 실행 중인 인스턴스 조회
	instances, err := m.dbiStore.ListRunning(ctx)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	// 사용 중인 리소스 계산
	var usedCPU float64
	var usedMemory int
	var instanceResources []coreresource.InstanceResources
	var activeCount int

	for _, inst := range instances {
		if inst.Status == coredbservice.StatusRunning ||
			inst.Status == coredbservice.StatusProvisioning {
			usedCPU += inst.Resources.CPU
			usedMemory += inst.Resources.Memory
			activeCount++

			instanceResources = append(instanceResources, coreresource.InstanceResources{
				InstanceID:   inst.ExternalID,
				InstanceName: inst.Name,
				Resources: coreresource.SystemResourceSpec{
					CPU:    inst.Resources.CPU,
					Memory: inst.Resources.Memory,
				},
				Status: string(inst.Status),
			})
		}
	}

	// 시스템 리소스 정보
	system := coreresource.SystemResources{
		Total: coreresource.SystemResourceSpec{
			CPU:    coreresource.TotalCPU,
			Memory: coreresource.TotalMemory,
		},
		Reserved: coreresource.SystemResourceSpec{
			CPU:    coreresource.SystemReservedCPU,
			Memory: coreresource.SystemReservedMemory,
		},
		Available: coreresource.SystemResourceSpec{
			CPU:    coreresource.AvailableCPU,
			Memory: coreresource.AvailableMemory,
		},
		Used: coreresource.SystemResourceSpec{
			CPU:    usedCPU,
			Memory: usedMemory,
		},
	}

	// 남은 리소스로 생성 가능 여부 체크
	remainingCPU := coreresource.AvailableCPU - usedCPU
	remainingMemory := coreresource.AvailableMemory - usedMemory

	return &coreresource.SystemResourceStatus{
		Info:            system,
		Instances:       instanceResources,
		ActiveCount:     activeCount,
		CanCreateTiny:   remainingCPU >= coreresource.TinyCPU && remainingMemory >= coreresource.TinyMemory,
		CanCreateSmall:  remainingCPU >= coreresource.SmallCPU && remainingMemory >= coreresource.SmallMemory,
		CanCreateMedium: remainingCPU >= coreresource.MediumCPU && remainingMemory >= coreresource.MediumMemory,
	}, nil
}

func (m *manager) GetAvailable(ctx context.Context) (*coreresource.SystemResourceSpec, error) {
	status, err := m.GetStatus(ctx)
	if err != nil {
		return nil, err
	}

	return &coreresource.SystemResourceSpec{
		CPU:    coreresource.AvailableCPU - status.Info.Used.CPU,
		Memory: coreresource.AvailableMemory - status.Info.Used.Memory,
	}, nil
}
