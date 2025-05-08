package redis

import (
	"context"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/platform/config"
	"github.com/redis/go-redis/v9"
	"log"
	"runtime/debug"
	"time"
)

type Client interface {
	Close() error
	Redis() *redis.Client
}

type client struct {
	redis  *redis.Client
	config config.RedisConfig
	logger *log.Logger
}

var _ Client = (*client)(nil)

func NewClient(config config.RedisConfig, logger *log.Logger) (Client, error) {
	opt, err := redis.ParseURL(config.ConnectString)
	if err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(opt)

	// 연결 테스트
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, errors.NewInternalErrorWithStack(fmt.Errorf("redis 연결 실패: %w", err), string(debug.Stack()))
	}

	return &client{
		redis:  redisClient,
		config: config,
		logger: logger,
	}, nil
}

func (c *client) Redis() *redis.Client {
	return c.redis
}

func (c *client) Close() error {
	return c.redis.Close()
}
