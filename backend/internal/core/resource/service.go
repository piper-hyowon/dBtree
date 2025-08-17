package resource

import "context"

type Manager interface {
	// CanAllocate 리소스 할당 가능 여부 확인
	CanAllocate(ctx context.Context, resources SystemResourceSpec) (bool, string, error)

	// GetStatus 전체 리소스 상태 조회
	GetStatus(ctx context.Context) (*SystemResourceStatus, error)

	// GetAvailable 사용 가능한 리소스 조회
	GetAvailable(ctx context.Context) (*SystemResourceSpec, error)
}
