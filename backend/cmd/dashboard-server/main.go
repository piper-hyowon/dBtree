package main

import (
	"context"
	stdErrors "errors"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/auth"
	authRest "github.com/piper-hyowon/dBtree/internal/auth/rest"
	coreauth "github.com/piper-hyowon/dBtree/internal/core/auth"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/dbservice"
	dbsRest "github.com/piper-hyowon/dBtree/internal/dbservice/rest"
	"github.com/piper-hyowon/dBtree/internal/email"
	"github.com/piper-hyowon/dBtree/internal/lemon"
	lemonRest "github.com/piper-hyowon/dBtree/internal/lemon/rest"
	"github.com/piper-hyowon/dBtree/internal/platform/k8s"
	"github.com/piper-hyowon/dBtree/internal/platform/rest/router"
	"github.com/piper-hyowon/dBtree/internal/quiz"
	quizRest "github.com/piper-hyowon/dBtree/internal/quiz/rest"
	"github.com/piper-hyowon/dBtree/internal/resource"
	resourceRest "github.com/piper-hyowon/dBtree/internal/resource/rest"
	"github.com/piper-hyowon/dBtree/internal/scheduler"
	"github.com/piper-hyowon/dBtree/internal/stats"
	statsRest "github.com/piper-hyowon/dBtree/internal/stats/rest"
	"strconv"

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
		//log.Fatal(".env 파일 없음")
		log.Println(".env 파일 없음 - 환경 변수 사용")
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

	k8sClient, err := k8s.NewClient(appConfig.K8s, logger)
	if err != nil {
		logger.Fatalf("K8S 연결 실패: %v", err)
	}
	fmt.Println(k8sClient.RESTConfig())

	sessionStore := auth.NewSessionStore(appConfig.UseLocalMemoryStore, pgClient.DB())
	userStore := user.NewStore(appConfig.UseLocalMemoryStore, pgClient.DB())
	lemonStore := lemon.NewLemonStore(appConfig.UseLocalMemoryStore, pgClient.DB())
	quizStore := quiz.NewStore(redisClient.Redis(), pgClient.DB())
	dbiStore := dbservice.NewDBIStore(appConfig.UseLocalMemoryStore, pgClient.DB())
	presetStore := dbservice.NewPresetStore(appConfig.UseLocalMemoryStore, pgClient.DB())
	portStore := dbservice.NewPortStore(appConfig.UseLocalMemoryStore, pgClient.DB())

	lemonService := lemon.NewService(lemonStore, quizStore, logger)

	// TODO: authService 에 lemonService 의존성 검토..
	authService := auth.NewService(
		sessionStore,
		lemonService,
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

	lemonHandler := lemonRest.NewHandler(lemonService, logger)

	quizService := quiz.NewService(quizStore, lemonStore, logger)
	quizHandler := quizRest.NewHandler(quizService, lemonService, logger)

	resourceManager := resource.NewManager(dbiStore, logger)
	resourceHandler := resourceRest.NewHandler(resourceManager, logger)

	dbsService := dbservice.NewService(appConfig.Server.PublicDBHost, dbiStore, presetStore, lemonService,
		userStore, k8sClient, portStore, resourceManager, logger)
	dbsHandler := dbsRest.NewHandler(appConfig.Server.PublicDBHost, dbsService, portStore, logger)

	statsService := stats.NewService(lemonStore, userStore, dbiStore, quizStore, logger)
	statsHandler := statsRest.NewHandler(statsService, logger)

	r := router.New(logger)

	r.GET("/system/resources", resourceHandler.GetSystemResources)

	r.POST("/db/instances", authMiddleware.RequireAuth(dbsHandler.CreateInstance))
	r.GET("/db/instances", authMiddleware.RequireAuth(dbsHandler.ListInstances))
	r.GET("/db/instances/:id", authMiddleware.RequireAuth(dbsHandler.GetInstanceWithSync))
	r.DELETE("/db/instances/:id", authMiddleware.RequireAuth(dbsHandler.DeleteInstance))
	r.POST("/db/instances/:id/:status", authMiddleware.RequireAuth(dbsHandler.UpdateInstanceStatus))
	r.GET("/db/presets", dbsHandler.ListPresets)

	r.POST("/verify-otp", func(w http.ResponseWriter, r *http.Request) {
		otpType := r.URL.Query().Get("type")
		if otpType == "authentication" {
			authHandler.VerifyOTP(w, r)
		} else {
			http.Error(w, "Invalid OTP type", http.StatusBadRequest)
		}
	})

	// 발송 or 재발송
	r.POST("/send-otp", func(w http.ResponseWriter, r *http.Request) {
		otpType := r.URL.Query().Get("type")
		if otpType == "authentication" {
			authHandler.SendOTP(w, r)
		} else {
			http.Error(w, "Invalid OTP type", http.StatusBadRequest)
		}
	})

	r.POST("/logout", authMiddleware.RequireAuth(authHandler.Logout))

	r.GET("/user", authMiddleware.RequireAuth(userHandler.Profile))
	r.DELETE("/user", authMiddleware.RequireAuth(userHandler.Delete))

	r.GET("/lemon/global-status", lemonHandler.TreeStatus)
	r.GET("/lemon/harvestable", authMiddleware.RequireAuth(lemonHandler.CanHarvest))
	r.POST("/lemon/harvest", authMiddleware.RequireAuth(lemonHandler.HarvestLemon))

	r.GET("/quiz/:positionID", authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		positionID := router.Param(r, "positionID")
		if positionID == "" {
			rest.HandleError(w, errors.NewMissingParameterError("positionID"), logger)
			return
		}

		if len(positionID) != 1 || positionID[0] < '0' || positionID[0] > '9' {
			rest.HandleError(w, errors.NewInvalidParameterError("positionID", "0-9"), logger)
			return
		}

		positionIDInt, err := strconv.Atoi(positionID)
		if err != nil {
			rest.HandleError(w, errors.NewInvalidParameterError("positionID", "int"), logger)
			return
		}

		quizHandler.StartQuiz(w, r, positionIDInt)
	}))
	r.POST("/quiz/answer", authMiddleware.RequireAuth(quizHandler.SubmitAnswer))

	r.GET("/stats/global", statsHandler.GetGlobalStats)
	r.GET("/leaderboard/mini", statsHandler.GetMiniLeaderboard)
	r.GET("/stats/daily-harvest", authMiddleware.RequireAuth(statsHandler.GetUserDailyHarvest))
	r.GET("/stats/transactions", authMiddleware.RequireAuth(statsHandler.GetUserTransactions))
	r.GET("/stats/summary/instances", authMiddleware.RequireAuth(statsHandler.GetUserInstances))

	r.POST("/support/inquiry", authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Category string `json:"category"`
			Subject  string `json:"subject"`
			Message  string `json:"message"`
			Context  struct {
				UserEmail    string `json:"userEmail"`
				LemonBalance int    `json:"lemonBalance"`
				Timestamp    string `json:"timestamp"`
				UserAgent    string `json:"userAgent"`
				CurrentPage  string `json:"currentPage"`
			} `json:"context"`
		}

		if !rest.DecodeJSONRequest(w, r, &req, logger) {
			return
		}

		u, err := rest.GetUserFromContext(r.Context())
		if err != nil {
			rest.HandleError(w, err, logger)
			return
		}

		subject := fmt.Sprintf("[dBtree 문의] %s - %s", req.Category, req.Subject)

		htmlBody := fmt.Sprintf(`
        <h2>새로운 문의가 접수되었습니다</h2>
        <hr>
        <h3>문의 정보</h3>
        <p><strong>카테고리:</strong> %s</p>
        <p><strong>제목:</strong> %s</p>
        <p><strong>사용자:</strong> %s (ID: %s)</p>
        <p><strong>레몬 잔액:</strong> %d</p>
        <hr>
        <h3>문의 내용</h3>
        <p>%s</p>
        <hr>
        <h3>컨텍스트 정보</h3>
        <p><strong>접수 시간:</strong> %s</p>
        <p><strong>현재 페이지:</strong> %s</p>
        <p><strong>User Agent:</strong> %s</p>
    `, req.Category, req.Subject, u.Email, u.ID,
			req.Context.LemonBalance, req.Message,
			req.Context.Timestamp, req.Context.CurrentPage,
			req.Context.UserAgent)

		// 이메일 발송
		ctx := r.Context()
		if err := emailService.Send(ctx, appConfig.AdminEmail, subject, htmlBody); err != nil {
			logger.Printf("문의 이메일 발송 실패: %v", err)
			rest.HandleError(w, err, logger)
			return
		}

		// 관리자에게 발송 후, 사용자에게도 확인 메일 발송
		userSubject := "dBtree 문의가 접수되었습니다"
		userHtmlBody := fmt.Sprintf(`
		<h2>문의가 접수되었습니다</h2>
        <p>안녕하세요, %s님</p>
        <p>문의해 주셔서 감사합니다. 24시간 내에 답변 드리겠습니다.</p>
        <hr>
        <h3>문의 내용</h3>
        <p><strong>제목:</strong> %s</p>
        <p>%s</p>
        `, u.Email, req.Subject, req.Message)

		go func() {
			if err := emailService.Send(context.Background(), u.Email, userSubject, userHtmlBody); err != nil {
				logger.Printf("사용자 확인 이메일 발송 실패: %v", err)
			}
		}()

		rest.SendSuccessResponse(w, http.StatusNoContent, nil)
	}))

	server := rest.NewServer(appConfig, r, logger)

	go cleanupSessions(sessionStore, appConfig.Session.CleanupIntervalHours, logger)

	lemonScheduler := scheduler.NewLemonScheduler(lemonStore, quizStore, logger, 1*time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := lemonScheduler.InitializeLemons(ctx); err != nil {
		logger.Printf("레몬 초기화 중 오류 발생: %v", err)
	}
	cancel()

	billingScheduler := scheduler.NewBillingScheduler(
		dbiStore,
		lemonStore,
		lemonService,
		k8sClient,
		logger,
		1*time.Hour, // 1시간마다 실행
	)

	lemonScheduler.Start()
	billingScheduler.Start()

	// 종료 시그널
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil && !stdErrors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("서버 시작 실패: %v", err)
		}
	}()

	// 종료 시그널 대기
	<-stopChan
	logger.Println("종료 신호 수신")
	lemonScheduler.Stop()
	billingScheduler.Stop()

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
