package quiz

import (
	"context"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
	"github.com/piper-hyowon/dBtree/internal/core/quiz"
	"log"
	"strconv"
	"time"
)

type service struct {
	quizStore  quiz.Store
	lemonStore lemon.Store
	logger     *log.Logger
}

var _ quiz.Service = (*service)(nil)

func NewService(quizStore quiz.Store, lemonStore lemon.Store, logger *log.Logger) quiz.Service {
	return &service{
		quizStore:  quizStore,
		lemonStore: lemonStore,
		logger:     logger,
	}
}

func (s *service) StartQuiz(ctx context.Context, positionID int, userID string, userEmail string) (*quiz.StartQuizWithPositionResponse, error) {
	// 이미 퀴즈 진행중이면 새로운 퀴즈 진행 불가
	inProgress, err := s.quizStore.InProgress(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	if inProgress != nil {
		return nil, errors.NewQuizInProgressError()
	}

	// 레몬 수확가능 상태 확인
	l, err := s.lemonStore.ByPositionID(ctx, positionID)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	if !l.IsAvailable {
		return nil, errors.NewResourceNotFoundError("available_lemon", strconv.Itoa(positionID))
	}

	// 퀴즈 가져오기
	q, err := s.quizStore.ByPositionID(ctx, positionID)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	startTime := time.Now()

	// 퀴즈 시작 정보 저장
	created, err := s.quizStore.CreateInProgress(ctx, userEmail, &quiz.StatusInfo{
		QuizID:         q.ID,
		PositionID:     positionID,
		StartTimestamp: startTime.Unix(),
		AttemptID:      0, // 임시값
	}, q.TimeLimit)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	if !created {
		return nil, errors.NewQuizInProgressError()
	}

	// 퀴즈 시작 기록 (PostgresSQL)
	attemptID, err := s.quizStore.CreateAttempt(ctx, userID, q.ID, positionID, startTime)
	if err != nil {
		_ = s.quizStore.DeleteInProgress(ctx, userEmail)
		return nil, errors.Wrap(err)
	}

	err = s.quizStore.UpdateInProgress(ctx, userEmail, attemptID)
	if err != nil {
		_ = s.quizStore.DeleteInProgress(ctx, userEmail)

		created, retryErr := s.quizStore.CreateInProgress(ctx, userEmail, &quiz.StatusInfo{
			QuizID:         q.ID,
			PositionID:     positionID,
			StartTimestamp: startTime.Unix(),
			AttemptID:      attemptID,
		}, q.TimeLimit)

		if retryErr != nil || !created {
			_ = s.quizStore.DeleteAttempt(ctx, attemptID)
			return nil, errors.Wrap(err)
		}
	}

	return &quiz.StartQuizWithPositionResponse{
		Question:  q.Question,
		Options:   q.Options,
		TimeLimit: q.TimeLimit,
		AttemptID: attemptID,
	}, nil
}

func (s *service) SubmitAnswer(ctx context.Context, userEmail string, attemptID int, optionIdx int) (*quiz.SubmitAnswerResponse, error) {
	// 진행중인 퀴즈 있는지 확인
	inProgress, err := s.quizStore.InProgress(ctx, userEmail)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	if inProgress == nil {
		return nil, errors.NewNoQuizInProgressError()
	}

	if attemptID != inProgress.AttemptID {
		return nil, errors.NewNoQuizInProgressError()
	}

	// 퀴즈 데이터 조회, 시간 체크
	quizData, err := s.quizStore.ByPositionID(ctx, inProgress.PositionID)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	now := time.Now()
	startTime := time.Unix(inProgress.StartTimestamp, 0)
	limitTime := startTime.Add(time.Duration(quizData.TimeLimit) * time.Second)
	isCorrect := quizData.CheckAnswer(optionIdx)

	// 시간 초과 체크
	var status quiz.Status
	if now.After(limitTime) {
		status = quiz.StatusTimeout
	} else {
		status = quiz.StatusDone
	}

	// 퀴즈 시도 기록 업데이트(통계용 데이터)
	err = s.quizStore.UpdateAttemptStatus(ctx, attemptID, status, &isCorrect, &optionIdx, now)
	if err != nil {
		s.logger.Printf("퀴즈 로그 업데이트 실패: %v", err) // 수확 API에 영향은 없지만 로그 정확성 떨어짐
	}

	// 퀴즈 종료(진행 상태 삭제)
	err = s.quizStore.DeleteInProgress(ctx, userEmail)
	if err != nil {
		s.logger.Printf("퀴즈 진행 상태 삭제 실패(Redis TTL에 의해 결국 삭제됨): %v", err)
	}

	response := &quiz.SubmitAnswerResponse{
		IsCorrect:     isCorrect,
		CorrectOption: quizData.CorrectOptionIdx,
		Status:        status,
		AttemptID:     attemptID,
	}

	if isCorrect && status != quiz.StatusTimeout {
		// 수확 상태 업데이트
		err = s.quizStore.UpdateAttemptHarvestStatus(ctx, attemptID, quiz.HarvestStatusInProgress, now)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		harvestTimeout := now.Add(time.Duration(quiz.HarvestTimeSeconds) * time.Second)
		response.HarvestEnabled = true
		response.HarvestTimeoutAt = &harvestTimeout
	}

	return response, nil
}
