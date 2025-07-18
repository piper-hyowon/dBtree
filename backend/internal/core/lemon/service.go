package lemon

import (
	"context"
	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
)

type Service interface {
	/* --------레몬 나무-------- */

	TreeStatus(ctx context.Context) (TreeStatusResponse, error)
	HarvestLemon(ctx context.Context, userID string, positionID int, attemptID int) (HarvestResponse, error)
	CanHarvest(ctx context.Context, userID string) (HarvestAvailability, error) // 마지막 수확 가능시간 체크

	ValidateInstanceCreation(ctx context.Context, userID string, cost dbservice.LemonCost) error                       // 잔액 체크
	ProcessInstanceFee(ctx context.Context, userID string, instanceID string, amount int, actionType ActionType) error // 인스턴스 비용 처리

	Transactions(ctx context.Context, userID string, limit, offset int) ([]*Transaction, error)

	GiveWelcomeLemon(ctx context.Context, userId string) error

	/* --------레몬 잔액 직접 변경-------- */

	AddLemons(ctx context.Context, userID string, amount int, actionType ActionType, note string) error
	DeductLemons(ctx context.Context, userID string, amount int, actionType ActionType, note string) error
}
