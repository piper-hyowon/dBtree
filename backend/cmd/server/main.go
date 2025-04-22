package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/piper-hyowon/dBtree/internal/adapters/primary/rest/dashboard/middleware"

	"github.com/joho/godotenv"
	"github.com/piper-hyowon/dBtree/internal/adapters/primary/core"
	"github.com/piper-hyowon/dBtree/internal/adapters/primary/rest/dashboard/auth"
	"github.com/piper-hyowon/dBtree/internal/adapters/secondary/db/memory"
	"github.com/piper-hyowon/dBtree/internal/adapters/secondary/email"
	"github.com/piper-hyowon/dBtree/internal/config"
)

// TODO: log 레벨별 출력 구분

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(".env 파일 없음")
		os.Exit(1)
	}

	appConfig, err := config.NewConfig()
	if err != nil {
		log.Fatal("환경 변수 설정 오류")
	}
	fmt.Println(appConfig)

	logger := log.New(os.Stdout, "[dBtree] ", log.LstdFlags|log.Lshortfile)
	logger.Println("서버 시작 중...")

	// 어댑터
	sessionRepo := memory.NewSessionRepo()
	userRepo := memory.NewUserRepo()
	emailService := setupEmailService(appConfig.SMTP)

	// 인증 서비스
	authService := core.NewAuthService(
		sessionRepo,
		emailService,
		userRepo,
		logger,
	)

	authHandler := auth.NewHandler(authService, logger)
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "dBtree")
	})
	mux.HandleFunc("/verify-otp", authHandler.VerifyOTP)

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

	// TODO: 유저 조회 API 작업 후 제거
	mux.HandleFunc("/user/profile", authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user := middleware.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "유저 인증 오류", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"created_at": user.CreatedAt,
		})
	}))

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(appConfig.Server.Port),
		Handler:      mux,
		ReadTimeout:  time.Duration(appConfig.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(appConfig.Server.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(appConfig.Server.IdleTimeoutSeconds) * time.Second,
	}

	go cleanupSessions(sessionRepo, appConfig.Session.CleanupIntervalHours, logger)

	startServer(server)
}

func setupEmailService(smtpConfig config.SMTPConfig) *email.SMTPEmailService {
	return email.NewSMTPEmailService(email.SMTPConfig{
		Host:     smtpConfig.Host,
		Port:     smtpConfig.Port,
		Username: smtpConfig.Username,
		Password: smtpConfig.Password,
		From:     smtpConfig.From,
	})
}

func cleanupSessions(sessionRepo *memory.SessionRepo, intervalHours int, logger *log.Logger) {
	ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		logger.Println("만료된 세션 정리 중...")
		if err := sessionRepo.Cleanup(context.Background()); err != nil {
			logger.Printf("세션 정리 오류: %v", err)
		}
	}
}

func startServer(server *http.Server) {
	log.Printf("HTTP 서버 시작, 포트: %s\n", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}
