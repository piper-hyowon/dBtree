package primary

import (
	"context"

	"github.com/piper-hyowon/dBtree/internal/domain/model"
)

type AuthService interface {
	StartAuth(ctx context.Context, email string) (isNewUser bool, err error)
	VerifyOTP(ctx context.Context, email string, code string) (createdUser *model.User, err error)
	ResendOTP(ctx context.Context, email string) error
}
