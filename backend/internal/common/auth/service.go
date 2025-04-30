package auth

import (
	"context"
	"github.com/piper-hyowon/dBtree/internal/common/user"
)

type Service interface {
	StartAuth(ctx context.Context, email string) (isNewUser bool, err error)
	GetSession(ctx context.Context, email string) (*Session, error)
	ResendOTP(ctx context.Context, email string) error
	VerifyOTP(ctx context.Context, email string, code string) (*user.User, string, error) // token 반환 추가
	ValidateSession(ctx context.Context, token string) (*user.User, error)
	Logout(ctx context.Context, token string) error
}
