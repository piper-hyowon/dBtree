package quiz

import (
	"database/sql"
	"github.com/piper-hyowon/dBtree/internal/core/quiz"
	"github.com/piper-hyowon/dBtree/internal/platform/store/combined"
	"github.com/redis/go-redis/v9"
)

func NewStore(cache *redis.Client, db *sql.DB) quiz.Store {
	return combined.NewQuizStore(cache, db)
}
