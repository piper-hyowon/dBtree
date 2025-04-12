package secondary

import (
	"context"

	"github.com/piper-hyowon/dBtree/internal/domain/model"
)

type SessionRepo interface {
	Save(ctx context.Context, session *model.AuthSession) error
	Get(ctx context.Context, email string) (*model.AuthSession, error)
	Delete(ctx context.Context, email string) error
	Cleanup(ctx context.Context) error // 만료 세션 정리
}
