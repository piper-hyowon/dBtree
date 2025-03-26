package secondary

import "github.com/piper-hyowon/dBtree/internal/domain/model"

type UserRepository interface {
	FindByEmail(email string) (*model.User, error)
	FindById(id string) (*model.User, error)
	Create(email string) error // 추가 파라미터 계획 없으므로 dto가 아닌 email만 사용
}
