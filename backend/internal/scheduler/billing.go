package scheduler

import (
	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
	"log"
	"sync"
	"time"
)

type BillingScheduler struct {
	dbService    dbservice.Service
	lemonService lemon.Service
	logger       *log.Logger

	ticker    *time.Ticker
	done      chan bool
	mutex     sync.Mutex
	isRunning bool
	interval  time.Duration
}
