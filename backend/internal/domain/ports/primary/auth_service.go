package primary

import (
	"github.com/piper-hyowon/dBtree/internal/domain/model"
)

type AuthService interface {
	// 인증 시작
	StartAuth(email string) (isNewUser bool, err error)
	VerifyOTP(email string, code string) (createdUser *model.User, err error)
	ResendOTP(email string) error
}
