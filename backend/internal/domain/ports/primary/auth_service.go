package primary

import (
	"context"

	"github.com/piper-hyowon/dBtree/internal/domain/model"
)

type AuthService interface {
	StartAuth(ctx context.Context, email string) (isNewUser bool, err error)
	GetSession(ctx context.Context, email string) (*model.AuthSession, error)
	ResendOTP(ctx context.Context, email string) error
	VerifyOTP(ctx context.Context, email string, code string) (*model.User, string, error) // token 반환 추가
	ValidateSession(ctx context.Context, token string) (*model.User, error)
	Logout(ctx context.Context, token string) error
}
