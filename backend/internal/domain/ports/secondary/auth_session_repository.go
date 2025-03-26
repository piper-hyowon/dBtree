package secondary

import "github.com/piper-hyowon/dBtree/internal/domain/model"

type AuthSessionRepository interface {
	FindByEmail(email string) (*model.AuthSession, error)
	Create(session *model.AuthSession) error
	Update(session *model.AuthSession) error
	Delete(email string) error
}
