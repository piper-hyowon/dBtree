package combined

import (
	"context"
	"database/sql"
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

func (q QuizStore) CheckInProgress(ctx context.Context, userEmail string) (*quiz.StatusInfo, error) {
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

	return &quiz.StatusInfo{
		QuizID:         quizID,
		PositionID:     positionID,
		StartTimestamp: startTime,
	}, nil
}

func (q QuizStore) SaveInProgress(ctx context.Context, userEmail string, info *quiz.StatusInfo, ttl time.Duration) (bool, error) {
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
	})
	pipe.Expire(ctx, key, ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, errors.NewInternalErrorWithStack(fmt.Errorf("redis 서버 에러: %v", err), string(debug.Stack()))
	}

	return true, nil
}

func (q QuizStore) DeleteInProgress(ctx context.Context, userEmail string) error {
	_, err := q.cache.Del(ctx, keys.InProgressKey(userEmail)).Result()
	if err != nil {
		return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}
	return nil
}

func (q QuizStore) SavePassed(ctx context.Context, userEmail string, positionID int, passedTime int64, ttl time.Duration) error {
	_, err := q.cache.Set(ctx, keys.UserPassKey(userEmail, positionID), passedTime, ttl).Result()
	if err != nil {
		return errors.NewInternalErrorWithStack(fmt.Errorf("redis <UNK> <UNK>: %v", err), string(debug.Stack()))
	}
	return nil
}

func (q QuizStore) PassedTime(ctx context.Context, userEmail string, positionID int) (int64, error) {

	passedTimeStr, err := q.cache.Get(ctx, keys.UserPassKey(userEmail, positionID)).Result()
	if err == redis.Nil {
		return 0, errors.NewQuizInProgressError()
	}
	if err != nil {
		return 0, err
	}

	passedTime, err := strconv.ParseInt(passedTimeStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return passedTime, nil
}

func (q QuizStore) DeletePassed(ctx context.Context, userEmail string, positionID int) error {
	_, err := q.cache.Del(ctx, keys.UserPassKey(userEmail, positionID)).Result()
	if err != nil {
		return errors.NewInternalErrorWithStack(fmt.Errorf("redis <UNK> <UNK>: %v", err), string(debug.Stack()))
	}
	return nil
}
