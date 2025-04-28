package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/auth"
	authHttp "github.com/piper-hyowon/dBtree/internal/auth/http"
	"github.com/piper-hyowon/dBtree/internal/auth/store"
	"github.com/piper-hyowon/dBtree/internal/email"
	"github.com/piper-hyowon/dBtree/internal/platform/config"
	"github.com/piper-hyowon/dBtree/internal/platform/middleware"
	"github.com/piper-hyowon/dBtree/internal/user"

	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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
	fmt.Println(appConfig)

	logger := log.New(os.Stdout, "[dBtree] ", log.LstdFlags|log.Lshortfile)
	logger.Println("서버 시작 중...")

	loggingMiddleware := middleware.LoggingMiddleware(logger, appConfig.DebugLogging)
	corsMiddleware := middleware.NewCORSMiddleware(middleware.CORSConfig{
		AllowedOrigins:   appConfig.CORS.AllowedOrigins,
		AllowCredentials: appConfig.CORS.AllowCredentials,
	})

	// 어댑터

	sessionStore := store.NewSessionStore()
	userStore := user.NewStore()
	emailService := setupEmailService(appConfig.SMTP)

	// 리소스 정리
	defer emailService.Close()

	// 인증 서비스

	authService := auth.NewService(
		sessionStore,
		emailService,
		userStore,
		logger,
	)

	authHandler := authHttp.NewHandler(authService, logger)
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
			http.Error(w, "유저 인증 오류", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"user":    u,
		})
	}))

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(appConfig.Server.Port),
		Handler:      loggingMiddleware(corsMiddleware(mux)),
		ReadTimeout:  time.Duration(appConfig.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(appConfig.Server.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(appConfig.Server.IdleTimeoutSeconds) * time.Second,
	}

	go cleanupSessions(sessionStore, appConfig.Session.CleanupIntervalHours, logger)

	// 서버 우아한 종료
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stopChan
		log.Println("종료 신호 수신, 서버를 종료합니다...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("서버 종료 중 오류: %v", err)
		}
	}()

	startServer(server)
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

func startServer(server *http.Server) {
	log.Printf("HTTP 서버 시작, 포트: %s\n", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}
