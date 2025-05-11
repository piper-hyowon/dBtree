package combined

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/core/quiz"
	"github.com/piper-hyowon/dBtree/internal/platform/store/redis/keys"
	"github.com/redis/go-redis/v9"
	"runtime/debug"
	"strconv"
	"time"
)

type QuizStore struct {
	cache *redis.Client
	rdb   *sql.DB
}

var _ quiz.Store = (*QuizStore)(nil)

func NewQuizStore(cache *redis.Client, rdb *sql.DB) quiz.Store {
	return &QuizStore{
		cache: cache,
		rdb:   rdb,
	}
}

/* ---------- 진행 중인 퀴즈 ---------- */

func (q QuizStore) InProgress(ctx context.Context, userEmail string) (*quiz.StatusInfo, error) {
	fields, err := q.cache.HGetAll(ctx, keys.InProgressKey(userEmail)).Result()
	if err != nil {
		return nil, errors.NewInternalErrorWithStack(fmt.Errorf("redis 서버 에러: %v", err), string(debug.Stack()))
	}

	if len(fields) == 0 {
		return nil, nil
	}

	quizID, _ := strconv.Atoi(fields["quiz_id"])
	positionID, _ := strconv.Atoi(fields["position_id"])
	startTime, _ := strconv.ParseInt(fields["start_time"], 10, 64)
	attemptID, _ := strconv.Atoi(fields["attempt_id"])

	return &quiz.StatusInfo{
		QuizID:         quizID,
		PositionID:     positionID,
		StartTimestamp: startTime,
		AttemptID:      attemptID,
	}, nil
}

func (q QuizStore) CreateInProgress(ctx context.Context, userEmail string, info *quiz.StatusInfo, timeLimit int) (bool, error) {
	key := keys.InProgressKey(userEmail)

	exists, err := q.cache.Exists(ctx, key).Result()
	if err != nil {
		return false, errors.NewInternalErrorWithStack(fmt.Errorf("redis 서버 에러: %v", err), string(debug.Stack()))
	}
	if exists == 1 {
		return false, errors.NewQuizInProgressError()
	}

	pipe := q.cache.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"quiz_id":     info.QuizID,
		"position_id": info.PositionID,
		"start_time":  info.StartTimestamp,
		"attempt_id":  info.AttemptID,
	})
	pipe.Expire(ctx, key, time.Duration(timeLimit+quiz.TimeBufferSeconds)*time.Second)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, errors.NewInternalErrorWithStack(fmt.Errorf("redis 서버 에러: %v", err), string(debug.Stack()))
	}

	return true, nil
}

func (q QuizStore) UpdateInProgress(ctx context.Context, userEmail string, attemptID int) error {
	key := keys.InProgressKey(userEmail)
	_, err := q.cache.HSet(ctx, key, "attempt_id", attemptID).Result()
	return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
}

func (q QuizStore) DeleteInProgress(ctx context.Context, userEmail string) error {
	_, err := q.cache.Del(ctx, keys.InProgressKey(userEmail)).Result()
	if err != nil {
		return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}
	return nil
}

/* ---------- 퀴즈 관련 ---------- */

func (q QuizStore) ByPositionID(ctx context.Context, positionID int) (*quiz.Quiz, error) {
	query := `
        SELECT qz.id, qz.question, qz.options, qz.correct_option_idx, 
               qz.difficulty, qz.category, qz.explanation, qz.time_limit,
               qz.is_active, qz.usage_count, lq.lemon_position_id
        FROM quizzes qz
        JOIN lemon_quiz lq ON qz.id = lq.quiz_id
        WHERE lq.lemon_position_id = $1
    `

	var quizData quiz.Quiz
	var optionsJSON string
	var difficultyStr, categoryStr string
	var explanation sql.NullString

	err := q.rdb.QueryRowContext(ctx, query, positionID).Scan(
		&quizData.ID,
		&quizData.Question,
		&optionsJSON,
		&quizData.CorrectOptionIdx,
		&difficultyStr,
		&categoryStr,
		&explanation,
		&quizData.TimeLimit,
		&quizData.IsActive,
		&quizData.UsageCount,
		&quizData.PositionID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.NewResourceNotFoundError("quiz_for_lemon", strconv.Itoa(positionID))
		}
		return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	if explanation.Valid {
		quizData.Explanation = explanation.String
	}

	var options []string
	if err = json.Unmarshal([]byte(optionsJSON), &options); err != nil {
		return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}
	quizData.Options = options

	quizData.Difficulty = quiz.Difficulty(difficultyStr)
	quizData.Category = quiz.Category(categoryStr)

	return &quizData, nil
}

func (q QuizStore) Random(ctx context.Context) (*quiz.Quiz, error) {
	query := `
        SELECT id, question, options, correct_option_idx, 
               difficulty, category, explanation, time_limit,
               is_active, usage_count
        FROM quizzes
        WHERE is_active = false
        ORDER BY usage_count ASC, RANDOM()
        LIMIT 1
    `

	var quizData quiz.Quiz
	var optionsJSON string
	var difficultyStr, categoryStr string

	err := q.rdb.QueryRowContext(ctx, query).Scan(
		&quizData.ID,
		&quizData.Question,
		&optionsJSON,
		&quizData.CorrectOptionIdx,
		&difficultyStr,
		&categoryStr,
		&quizData.Explanation,
		&quizData.TimeLimit,
		&quizData.IsActive,
		&quizData.UsageCount,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.NewResourceNotFoundError("available_quiz", "random")
		}
		return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	var options []string
	if err = json.Unmarshal([]byte(optionsJSON), &options); err != nil {
		return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}
	quizData.Options = options

	quizData.Difficulty = quiz.Difficulty(difficultyStr)
	quizData.Category = quiz.Category(categoryStr)

	return &quizData, nil
}

func (q QuizStore) AssignToLemon(ctx context.Context, quizID int, positionID int) error {
	tx, err := q.rdb.BeginTx(ctx, nil)
	if err != nil {
		return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}
	defer tx.Rollback()

	// 1. 기존 퀴즈가 있다면 비활성화
	deactivateQuery := `
        UPDATE quizzes q
        SET is_active = false
        FROM lemon_quiz_mappings m
        WHERE q.id = m.quiz_id AND m.lemon_position_id = $1
    `
	_, err = tx.ExecContext(ctx, deactivateQuery, positionID)
	if err != nil {
		return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	// 2. 매핑 테이블 업데이트 (UPSERT 방식)
	upsertQuery := `
        INSERT INTO lemon_quiz_mappings (lemon_position_id, quiz_id)
        VALUES ($1, $2)
        ON CONFLICT (lemon_position_id)
        DO UPDATE SET quiz_id = EXCLUDED.quiz_id
    `
	_, err = tx.ExecContext(ctx, upsertQuery, positionID, quizID)
	if err != nil {
		return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	// 3. 새 퀴즈 활성화 및 사용 횟수 증가
	activateQuery := `
        UPDATE quizzes
        SET is_active = true, usage_count = usage_count + 1
        WHERE id = $1
    `
	_, err = tx.ExecContext(ctx, activateQuery, quizID)
	if err != nil {
		return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	if err = tx.Commit(); err != nil {
		return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	return nil
}

func (q QuizStore) Deactivate(ctx context.Context, positionID int) error {
	query := `
		UPDATE quizzes qz
		SET is_active = false
		FROM lemon_quiz_mappings m
		WHERE m.quiz_id = m.quiz_id AND m.lemon_position_id = $1`

	_, err := q.rdb.ExecContext(ctx, query, positionID)
	if err != nil {
		return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}
	return nil

}

/* ---------- 퀴즈 시도 관련 ---------- */

func (q QuizStore) CreateAttempt(ctx context.Context, userID string, quizID int, positionID int, startTime time.Time) (int, error) {
	query := `
        INSERT INTO user_quiz_attempts 
        (user_id, quiz_id, lemon_position_id, start_time)
        VALUES ($1, $2, $3, $4)
        RETURNING id`

	var attemptID int
	err := q.rdb.QueryRowContext(ctx, query,
		userID, quizID, positionID, startTime).Scan(&attemptID)
	if err != nil {
		return 0, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	return attemptID, nil
}

func (q QuizStore) AttemptByID(ctx context.Context, id int) (*quiz.Attempt, error) {
	query := `
        SELECT id, user_id, quiz_id, lemon_position_id, is_correct,
               selected_option, start_time, submit_time, time_taken,
               time_taken_clicked, status, harvest_status
        FROM user_quiz_attempts
        WHERE id = $1`

	var a quiz.Attempt
	var userID string
	var isCorrect sql.NullBool
	var selectedOption sql.NullInt32
	var submitTime sql.NullTime
	var timeTaken, timeTakenClicked sql.NullInt32
	var status, harvestStatus string

	err := q.rdb.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &userID, &a.QuizID, &a.LemonPositionID, &isCorrect,
		&selectedOption, &a.StartTime, &submitTime, &timeTaken,
		&timeTakenClicked, &status, &harvestStatus,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	a.UserID = userID
	a.IsCorrect = isCorrect.Bool
	a.SelectedOption = int(selectedOption.Int32)
	a.SubmitTime = submitTime.Time
	a.TimeTaken = int(timeTaken.Int32)
	a.TimeTakenClicked = int(timeTakenClicked.Int32)
	a.Status = quiz.Status(status)
	a.HarvestStatus = quiz.HarvestStatus(harvestStatus)

	return &a, nil
}

func (q QuizStore) UpdateAttemptStatus(ctx context.Context, attemptID int, status quiz.Status, isCorrect *bool, selectedOption *int, submitTime time.Time) error {
	query := `
        UPDATE user_quiz_attempts
        SET status = $1, 
            is_correct = $2,
            selected_option = $3,
            submit_time = $4,
            time_taken = EXTRACT(EPOCH FROM ($4 - start_time))::INTEGER
        WHERE id = $5
    `

	_, err := q.rdb.ExecContext(ctx, query, status, isCorrect, selectedOption, submitTime, attemptID)
	if err != nil {
		return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	return nil
}

func (q QuizStore) UpdateAttemptHarvestStatus(ctx context.Context, attemptID int, harvestStatus quiz.HarvestStatus, clickTime time.Time) error {
	query := `
        UPDATE user_quiz_attempts
        SET harvest_status = $1, 
            time_taken = EXTRACT(EPOCH FROM ($2 - start_time))::INTEGER
        WHERE id = $3
    `

	_, err := q.rdb.ExecContext(ctx, query, harvestStatus, clickTime, attemptID)
	if err != nil {
		return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	return nil
}

func (q QuizStore) DeleteAttempt(ctx context.Context, attemptID int) error {
	_, err := q.rdb.ExecContext(ctx, "DELETE FROM user_quiz_attempts WHERE id = $1", attemptID)
	return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
}
