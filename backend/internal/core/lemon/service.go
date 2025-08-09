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

	/* --------레몬 잔액 변경-------- */

	ValidateInstanceCreation(ctx context.Context, userID string, cost dbservice.LemonCost) error                                                  // 잔액 체크
	ProcessInstanceFee(ctx context.Context, userID string, externalInstanceID string, amount int, actionType ActionType, instanceID *int64) error // 인스턴스 비용 처리
	GiveWelcomeLemon(ctx context.Context, userId string) error

	/* --------레몬 잔액 직접 변경-------- */

	AddLemons(ctx context.Context, userID string, amount int, actionType ActionType, note string, instanceID *int64) error
	DeductLemons(ctx context.Context, userID string, amount int, actionType ActionType, note string, instanceID *int64) error

	/* --------유저 데이터 조회-------- */

	DailyHarvestStats(ctx context.Context, userID string, days int) ([]*DailyHarvest, error)
}
