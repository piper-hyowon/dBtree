package main

import (
	"context"
	"errors"
	"github.com/piper-hyowon/dBtree/internal/auth"
	authRest "github.com/piper-hyowon/dBtree/internal/auth/rest"
	coreauth "github.com/piper-hyowon/dBtree/internal/core/auth"
	"github.com/piper-hyowon/dBtree/internal/email"
	"github.com/piper-hyowon/dBtree/internal/lemon"
	lemonRest "github.com/piper-hyowon/dBtree/internal/lemon/rest"
	"github.com/piper-hyowon/dBtree/internal/platform/config"
	"github.com/piper-hyowon/dBtree/internal/platform/rest"
	"github.com/piper-hyowon/dBtree/internal/platform/store/postgres"
	"github.com/piper-hyowon/dBtree/internal/platform/store/redis"
	"github.com/piper-hyowon/dBtree/internal/user"
	userRest "github.com/piper-hyowon/dBtree/internal/user/rest"
	"os/signal"
	"syscall"

	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// TODO: log 레벨별 출력 구분

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(".env 파일 없음")
	}
	appConfig, err := config.NewConfig()
	if err != nil {
		log.Fatalf("환경 변수 설정 오류: %v", err)
	}

	logger := log.New(os.Stdout, "[dBtree] ", log.LstdFlags|log.Lshortfile)
	logger.Println("서버 시작 중...")

	emailService, err := email.NewService(appConfig.SMTP)
	if err != nil {
		logger.Fatalf("이메일 서비스 초기화 실패: %v", err)
	}
	defer emailService.Close()

	pgClient, err := postgres.NewClient(appConfig.Postgres, logger)
	if err != nil {
		logger.Fatalf("PostgreSQL 초기화 실패: %v", err)
	}
	defer pgClient.Close()

	redisClient, err := redis.NewClient(appConfig.Redis, logger)
	if err != nil {
		logger.Fatalf("Redis 연결 실패: %v", err)
	}
	defer redisClient.Close()

	sessionStore := auth.NewSessionStore(appConfig.UseLocalMemoryStore, pgClient.DB())
	userStore := user.NewStore(appConfig.UseLocalMemoryStore, pgClient.DB())
	lemonStore := lemon.NewLemonStore(appConfig.UseLocalMemoryStore, pgClient.DB())

	authService := auth.NewService(
		sessionStore,
		emailService,
		userStore,
		logger,
	)

	authHandler := authRest.NewHandler(authService, logger)
	authMiddleware := rest.NewAuthMiddleware(authService, logger)

	userService := user.NewService(
		emailService,
		userStore,
		sessionStore,
		logger,
	)

	userHandler := userRest.NewHandler(userService, lemonStore, logger)

	lemonService := lemon.NewService(lemonStore)
	lemonHandler := lemonRest.NewHandler(lemonService, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/verify-otp", func(w http.ResponseWriter, r *http.Request) {
		otpType := r.URL.Query().Get("type")
		if otpType == "authentication" {
			authHandler.VerifyOTP(w, r)
		} else {
			http.Error(w, "Invalid OTP type", http.StatusBadRequest)
		}
	})

	// 발송 or 재발송
	mux.HandleFunc("/send-otp", func(w http.ResponseWriter, r *http.Request) {
		otpType := r.URL.Query().Get("type")
		if otpType == "authentication" {
			authHandler.SendOTP(w, r)
		} else {
			http.Error(w, "Invalid OTP type", http.StatusBadRequest)
		}
	})

	mux.HandleFunc("/logout", authMiddleware.RequireAuth(authHandler.Logout))
	mux.HandleFunc("/user", authMiddleware.RequireAuth(userHandler.Default))
	mux.HandleFunc("/lemon/global-status", lemonHandler.TreeStatus)
	mux.HandleFunc("/lemon/harvestable", authMiddleware.RequireAuth(lemonHandler.CanHarvest))
	mux.HandleFunc("/lemon/harvest", authMiddleware.RequireAuth(lemonHandler.HarvestLemon))

	server := rest.NewServer(appConfig, mux, logger)

	go cleanupSessions(sessionStore, appConfig.Session.CleanupIntervalHours, logger)

	lemonScheduler := lemon.NewScheduler(lemonStore, logger, 1*time.Minute)
	lemonScheduler.Start()

	// 종료 시그널
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("서버 시작 실패: %v", err)
		}
	}()

	// 종료 시그널 대기
	<-stopChan
	logger.Println("종료 신호 수신")
	lemonScheduler.Stop()

	if err := server.GracefulShutdown(5 * time.Second); err != nil {
		logger.Fatalf("서버 종료 중 오류: %v", err)
	}
}

func cleanupSessions(sessionStore coreauth.SessionStore, intervalHours int, logger *log.Logger) {
	ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		logger.Println("만료된 세션 정리 중...")
		if err := sessionStore.Cleanup(context.Background()); err != nil {
			logger.Printf("세션 정리 오류: %v", err)
		}
	}
}
