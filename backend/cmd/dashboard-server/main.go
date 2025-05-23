package main

import (
	"context"
	"errors"
	"github.com/piper-hyowon/dBtree/internal/auth"
	authRest "github.com/piper-hyowon/dBtree/internal/auth/rest"
	"github.com/piper-hyowon/dBtree/internal/auth/store"
	"github.com/piper-hyowon/dBtree/internal/email"
	"github.com/piper-hyowon/dBtree/internal/platform/config"
	"github.com/piper-hyowon/dBtree/internal/platform/rest"
	middleware "github.com/piper-hyowon/dBtree/internal/platform/rest/middleware"
	"github.com/piper-hyowon/dBtree/internal/user"
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
		log.Fatal("환경 변수 설정 오류")
	}

	logger := log.New(os.Stdout, "[dBtree] ", log.LstdFlags|log.Lshortfile)
	logger.Println("서버 시작 중...")

	sessionStore := store.NewSessionStore()
	userStore := user.NewStore()
	emailService := setupEmailService(appConfig.SMTP)
	defer emailService.Close()

	authService := auth.NewService(
		sessionStore,
		emailService,
		userStore,
		logger,
	)

	authHandler := authRest.NewHandler(authService, logger)
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)

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

	mux.HandleFunc("/profile", authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		u := middleware.GetUserFromContext(r.Context())
		if u == nil {
			rest.SendErrorResponse(w, http.StatusInternalServerError, "유저 인증 오류")
			return
		}

		rest.SendSuccessResponse(w, http.StatusOK, map[string]interface{}{
			"user": u,
		})
	}))
	server := rest.NewServer(appConfig, mux, logger)

	go cleanupSessions(sessionStore, appConfig.Session.CleanupIntervalHours, logger)

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
	if err := server.GracefulShutdown(5 * time.Second); err != nil {
		logger.Fatalf("서버 종료 중 오류: %v", err)
	}
}

func setupEmailService(smtpConfig config.SMTPConfig) email.Service {
	return email.NewSmtpService(email.SMTPConfig{
		Host:     smtpConfig.Host,
		Port:     smtpConfig.Port,
		Username: smtpConfig.Username,
		Password: smtpConfig.Password,
		From:     smtpConfig.From,
	})
}

func cleanupSessions(sessionStore auth.SessionStore, intervalHours int, logger *log.Logger) {
	ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		logger.Println("만료된 세션 정리 중...")
		if err := sessionStore.Cleanup(context.Background()); err != nil {
			logger.Printf("세션 정리 오류: %v", err)
		}
	}
}
