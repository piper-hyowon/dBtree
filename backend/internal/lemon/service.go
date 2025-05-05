package lemon

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
	"runtime/debug"
	"time"
)

type service struct {
	store lemon.Store
}

var _ lemon.Service = (*service)(nil)

func NewService(store lemon.Store) lemon.Service {
	return &service{
		store: store,
	}
}

func (s *service) TreeStatus(ctx context.Context) (lemon.TreeStatus, error) {
	positions, err := s.store.AvailablePositions(ctx)
	if err != nil {
		return lemon.TreeStatus{}, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	totalHarvested, err := s.store.TotalHarvestedCount(ctx)
	if err != nil {
		return lemon.TreeStatus{}, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	var nextRegrowthTime *time.Time
	// 레몬이 모두 수확가능한 경우는 생략
	if len(positions) < lemon.DefaultRegrowthRules.MaxPositions {
		// 다음 재생성 시간 계산(가장 빠른 시간)
		t, err := s.store.NextRegrowthTime(ctx)
		if err != nil {
			return lemon.TreeStatus{}, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
		}
		nextRegrowthTime = t
	} else {
		nextRegrowthTime = nil
	}

	return lemon.TreeStatus{
		AvailablePositions: positions,
		TotalHarvested:     totalHarvested,
		NextRegrowthTime:   nextRegrowthTime,
	}, nil
}

func (s *service) HarvestLemon(ctx context.Context, userID string, positionID int) (lemon.HarvestResult, error) {
	availability, err := s.CanHarvest(ctx, userID)
	if err != nil {
		return lemon.HarvestResult{}, err
	}

	if !availability.CanHarvest {
		return lemon.HarvestResult{}, errors.NewHarvestCooldownError(availability.WaitTime)
	}

	now := time.Now()

	// 사용자 잔액 조회
	balanceBefore, err := s.store.UserBalance(ctx, userID)
	if err != nil {
		return lemon.HarvestResult{}, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	harvestAmount := lemon.DefaultHarvestRules.BaseAmount
	newBalance := balanceBefore + lemon.DefaultHarvestRules.BaseAmount

	// 최대 저장 가능 레몬 수 제한
	if newBalance > lemon.DefaultHarvestRules.MaxStoredLemons {
		harvestAmount = lemon.DefaultHarvestRules.MaxStoredLemons - balanceBefore
		if harvestAmount <= 0 {
			return lemon.HarvestResult{}, errors.NewLemonStorageFullError()
		}
		newBalance = lemon.DefaultHarvestRules.MaxStoredLemons
	}

	txID, err := s.store.HarvestWithTransaction(ctx, positionID, userID, harvestAmount, now)
	if err != nil {
		return lemon.HarvestResult{}, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	return lemon.HarvestResult{
		HarvestAmount:   harvestAmount,
		NewBalance:      newBalance,
		TransactionID:   txID,
		NextHarvestTime: lemon.DefaultHarvestRules.CooldownPeriod,
	}, nil
}

func (s *service) AddLemons(ctx context.Context, userID string, amount int, actionType lemon.ActionType, note string) error {
	// 사용자 잔액 조회
	balance, err := s.store.UserBalance(ctx, userID)
	if err != nil {
		return err
	}

	newBalance := balance + amount

	// 최대 저장량 초과 체크
	if newBalance > lemon.DefaultHarvestRules.MaxStoredLemons {
		return errors.NewLemonStorageFullError()
	}

	// 트랜잭션 생성
	tx := &lemon.Transaction{
		ID:         uuid.New().String(),
		UserID:     userID,
		ActionType: actionType, // 적절한 액션 타입 설정 필요
		Status:     lemon.StatusSuccessful,
		Amount:     amount,
		Balance:    newBalance,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Note:       note,
	}

	// 트랜잭션 저장
	return s.store.CreateTransaction(ctx, tx)
}

func (s *service) DeductLemons(ctx context.Context, userID string, amount int, actionType lemon.ActionType, note string) error {
	// 사용자 잔액 조회
	balance, err := s.store.UserBalance(ctx, userID)
	if err != nil {
		return err
	}

	// 잔액 부족 체크
	if balance < amount {
		//  잔액 부족 에러
		return errors.NewInsufficientLemonsError(amount, amount-balance)
	}

	newBalance := balance - amount

	// 트랜잭션 생성
	tx := &lemon.Transaction{
		ID:         uuid.New().String(),
		UserID:     userID,
		ActionType: actionType, // 적절한 액션 타입 설정 필요
		Status:     lemon.StatusSuccessful,
		Amount:     -amount, // 음수로 저장
		Balance:    newBalance,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Note:       note,
	}

	// 트랜잭션 저장
	return s.store.CreateTransaction(ctx, tx)
}

func (s *service) ValidateInstanceCreation(ctx context.Context, userID string, cost dbservice.LemonCost) error {
	// 사용자 잔액 조회
	balance, err := s.store.UserBalance(ctx, userID)
	if err != nil {
		return err
	}

	// 최소 필요 레몬 체크
	if balance < cost.MinimumLemons {
		return errors.NewInsufficientLemonsError(cost.MinimumLemons, cost.MinimumLemons-balance)
	}

	return nil
}

func (s *service) ProcessInstanceFee(ctx context.Context, userID string, instanceID string, amount int, actionType lemon.ActionType) error {
	return s.DeductLemons(ctx, userID, amount, actionType, fmt.Sprintf("인스턴스 %s 유지 비용", instanceID))

}

func (s *service) Transactions(ctx context.Context, userID string, limit, offset int) ([]*lemon.Transaction, error) {
	return s.store.TransactionListByUserID(ctx, userID, limit, offset)
}

func (s *service) GiveWelcomeLemon(ctx context.Context, userID string) error {
	return s.AddLemons(ctx, userID, lemon.WelcomeBonusAmount, lemon.ActionWelcomeBonus, "회원가입 보너스")
}

func (s *service) CanHarvest(ctx context.Context, userID string) (lemon.HarvestAvailability, error) {
	lastHarvestTime, err := s.store.UserLastHarvestTime(ctx, userID)
	if err != nil {
		return lemon.HarvestAvailability{}, err
	}

	now := time.Now()

	if lastHarvestTime == nil {
		return lemon.HarvestAvailability{
			CanHarvest: true,
			WaitTime:   0,
		}, nil
	}

	// 쿨다운 시간 계산
	cooldownEndTime := lastHarvestTime.Add(lemon.DefaultHarvestRules.CooldownPeriod)
	if now.Before(cooldownEndTime) {
		waitTime := cooldownEndTime.Sub(now)
		return lemon.HarvestAvailability{
			CanHarvest: false,
			WaitTime:   waitTime,
		}, nil
	}

	return lemon.HarvestAvailability{
		CanHarvest: true,
		WaitTime:   0,
	}, nil
}
