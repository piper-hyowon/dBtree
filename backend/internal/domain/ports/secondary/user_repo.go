package secondary

import (
	"context"

	"github.com/piper-hyowon/dBtree/internal/domain/model"
)

type UserRepo interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)

	FindById(ctx context.Context, id string) (*model.User, error)

	// 추가 파라미터 계획 없으므로 dto가 아닌 email만 사용
	Create(ctx context.Context, email string) error
}
