package lemon

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
	"github.com/piper-hyowon/dBtree/internal/core/quiz"
	"log"
	"time"
)

type service struct {
	store     lemon.Store
	quizStore quiz.Store
	logger    *log.Logger // TODO: core.Logger 인터페이스 정의해서 사용
}

var _ lemon.Service = (*service)(nil)

func NewService(store lemon.Store, quizStore quiz.Store, logger *log.Logger) lemon.Service {
	return &service{
		store:     store,
		quizStore: quizStore,
		logger:    logger,
	}
}

func (s *service) TreeStatus(ctx context.Context) (lemon.TreeStatusResponse, error) {
	positions, err := s.store.AvailablePositions(ctx)
	if err != nil {
		return lemon.TreeStatusResponse{}, errors.Wrap(err)
	}

	totalHarvested, err := s.store.TotalHarvestedCount(ctx)
	if err != nil {
		return lemon.TreeStatusResponse{}, errors.Wrap(err)
	}

	var nextRegrowthTime *time.Time
	// 레몬이 모두 수확가능한 경우는 생략
	if len(positions) < lemon.DefaultRegrowthRules.MaxPositions {
		// 다음 재생성 시간 계산(가장 빠른 시간)
		t, err := s.store.NextRegrowthTime(ctx)
		if err != nil {
			return lemon.TreeStatusResponse{}, errors.Wrap(err)
		}
		nextRegrowthTime = t
	} else {
		nextRegrowthTime = nil
	}

	return lemon.TreeStatusResponse{
		AvailablePositions: positions,
		TotalHarvested:     totalHarvested,
		NextRegrowthTime:   nextRegrowthTime,
	}, nil
}

func (s *service) HarvestLemon(ctx context.Context, userID string, positionID int, attemptID int) (lemon.HarvestResponse, error) {
	// 퀴즈 시도 기록 조회
	attempt, err := s.quizStore.AttemptByID(ctx, attemptID)
	if err != nil {
		return lemon.HarvestResponse{}, errors.Wrap(err)
	}

	if attempt == nil {
		return lemon.HarvestResponse{}, errors.NewNoQuizPassedError()
	}

	if attempt.UserID != userID || !attempt.IsCorrect || attempt.Status != quiz.StatusDone || attempt.LemonPositionID != positionID {
		return lemon.HarvestResponse{}, errors.NewNoQuizPassedError()
	}

	// 이미 성공/실패 처리되었는지(중복 보상 방지)
	if attempt.HarvestStatus != quiz.HarvestStatusInProgress {
		return lemon.HarvestResponse{}, errors.NewHarvestAlreadyProcessedError()
	}

	// 원 클릭 시간 제한 확인
	now := time.Now()
	timeSinceSubmit := time.Since(attempt.SubmitTime).Seconds()
	if timeSinceSubmit > float64(quiz.HarvestTimeSeconds) {
		// 시간 초과 - 상태 업데이트
		err = s.quizStore.UpdateAttemptHarvestStatus(ctx, attemptID, quiz.HarvestStatusTimeout, now)
		if err != nil {
			s.logger.Printf("수확 상태 업데이트 실패: %v", err)
		}

		return lemon.HarvestResponse{}, errors.NewClickCircleTimeExpiredError(
			now,
			attempt.SubmitTime.Add(time.Duration(quiz.HarvestTimeSeconds)*time.Second),
		)
	}

	// 사용자 잔액 조회
	balanceBefore, err := s.store.UserBalance(ctx, userID)
	if err != nil {
		return lemon.HarvestResponse{}, errors.Wrap(err)
	}

	harvestAmount := lemon.DefaultHarvestRules.BaseAmount
	newBalance := balanceBefore + lemon.DefaultHarvestRules.BaseAmount

	// 최대 저장 가능 레몬 수 제한
	if newBalance > lemon.DefaultHarvestRules.MaxStoredLemons {
		harvestAmount = lemon.DefaultHarvestRules.MaxStoredLemons - balanceBefore
		if harvestAmount <= 0 {
			return lemon.HarvestResponse{}, errors.NewLemonStorageFullError(lemon.DefaultHarvestRules.MaxStoredLemons)
		}
		newBalance = lemon.DefaultHarvestRules.MaxStoredLemons
	}

	txID, err := s.store.HarvestWithTransaction(ctx, positionID, userID, harvestAmount, now)
	if err != nil {
		// 다른 사람이 이미 수확
		if errors.Is(err, errors.NewLemonAlreadyHarvestedError()) {
			updateErr := s.quizStore.UpdateAttemptHarvestStatus(ctx, attemptID, quiz.HarvestStatusFailure, now)
			fmt.Print(updateErr)
			if updateErr != nil {
				s.logger.Printf("수확 상태 업데이트 실패: %v", updateErr)
				return lemon.HarvestResponse{}, errors.Wrapf(updateErr, "수확 상태 업데이트 실패")

			}

			return lemon.HarvestResponse{}, err
		}
		return lemon.HarvestResponse{}, errors.Wrap(err)
	}

	// 수확 상태 업데이트
	err = s.quizStore.UpdateAttemptHarvestStatus(ctx, attemptID, quiz.HarvestStatusSuccess, now)
	if err != nil {
		s.logger.Printf("수확 상태 업데이트 실패: %v", err)
		// 로그만 남기고 계속 진행 (잔액은 이미 반영 완료) // TODO: 에러리포팅?
	}

	return lemon.HarvestResponse{
		HarvestAmount:   harvestAmount,
		NewBalance:      newBalance,
		TransactionID:   txID,
		NextHarvestTime: lemon.DefaultHarvestRules.CooldownPeriod,
	}, nil
}

func (s *service) AddLemons(ctx context.Context, userID string, amount int, actionType lemon.ActionType, note string, instanceID *int64) error {
	// 사용자 잔액 조회
	balance, err := s.store.UserBalance(ctx, userID)
	if err != nil {
		return err
	}

	newBalance := balance + amount

	// 최대 저장량 초과 체크
	if newBalance > lemon.DefaultHarvestRules.MaxStoredLemons {
		return errors.NewLemonStorageFullError(lemon.DefaultHarvestRules.MaxStoredLemons)
	}

	// 트랜잭션 생성
	tx := &lemon.Transaction{
		ID:         uuid.New().String(),
		UserID:     userID,
		InstanceID: instanceID,
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

func (s *service) DeductLemons(ctx context.Context, userID string, amount int, actionType lemon.ActionType, note string, instanceID *int64) error {
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
		InstanceID: instanceID,
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
	if balance < cost.CreationCost+cost.HourlyLemons {
		return errors.NewInsufficientLemonsError(cost.CreationCost+cost.HourlyLemons, cost.CreationCost+cost.HourlyLemons-balance)
	}

	return nil
}

func (s *service) ProcessInstanceFee(ctx context.Context, userID string, externalInstanceID string, amount int, actionType lemon.ActionType, instanceID *int64) error {
	return s.DeductLemons(ctx, userID, amount, actionType, fmt.Sprintf("인스턴스 %s 유지 비용", externalInstanceID), instanceID)

}

func (s *service) GiveWelcomeLemon(ctx context.Context, userID string) error {
	return s.AddLemons(ctx, userID, lemon.WelcomeBonusAmount, lemon.ActionWelcomeBonus, "회원가입 보너스", nil)
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

func (s *service) DailyHarvestStats(ctx context.Context, userID string, days int) ([]*lemon.DailyHarvest, error) {
	if days <= 0 {
		days = 7 // 기본값
	}
	if days > 365 {
		days = 365 // 최대 1년
	}

	return s.store.DailyHarvestStats(ctx, userID, days)
}
