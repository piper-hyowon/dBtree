package user

import "context"

type Store interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindById(ctx context.Context, id string) (*User, error)

	// 추가 파라미터 계획 없으므로 dto가 아닌 email만 사용
	CreateIfNotExists(ctx context.Context, email string) error

	Delete(ctx context.Context, id string) error // 탈퇴
}
