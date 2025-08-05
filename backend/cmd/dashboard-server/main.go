package main

import (
	"context"
	stdErrors "errors"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/auth"
	authRest "github.com/piper-hyowon/dBtree/internal/auth/rest"
	coreauth "github.com/piper-hyowon/dBtree/internal/core/auth"
	dbsDomain "github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/dbservice"
	"github.com/piper-hyowon/dBtree/internal/email"
	"github.com/piper-hyowon/dBtree/internal/lemon"
	lemonRest "github.com/piper-hyowon/dBtree/internal/lemon/rest"
	"github.com/piper-hyowon/dBtree/internal/platform/k8s"
	"github.com/piper-hyowon/dBtree/internal/platform/rest/router"
	"github.com/piper-hyowon/dBtree/internal/quiz"
	quizRest "github.com/piper-hyowon/dBtree/internal/quiz/rest"
	"regexp"
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

	lemonService := lemon.NewService(lemonStore, quizStore, logger)
	lemonHandler := lemonRest.NewHandler(lemonService, logger)

	quizService := quiz.NewService(quizStore, lemonStore, logger)
	quizHandler := quizRest.NewHandler(quizService, lemonService, logger)

	dbsService := dbservice.NewService(appConfig.Server.PublicHost, dbiStore, presetStore, lemonService,
		userStore, k8sClient, portStore, logger)

	r := router.New(logger)

	// TODO 지워야하ㅑㅁ 테스트

	r.POST("/test", authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, err := rest.GetUserFromContext(r.Context())
		if err != nil {
			rest.HandleError(w, err, logger)
			return
		}

		var dto dbsDomain.CreateInstanceRequest
		if !rest.DecodeJSONRequest(w, r, &dto, logger) {
			return
		}
		if err := validateInstanceName(dto.Name); err != nil {
			rest.HandleError(w, err, logger)
			return
		}

		resp, err := dbsService.CreateInstance(context.Background(), user.ID, user.LemonBalance, &dto)

		if err != nil {
			rest.HandleError(w, err, logger)
			return
		}

		rest.SendJSONResponse(w, http.StatusAccepted, resp)

	}))

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

	server := rest.NewServer(appConfig, r, logger)

	go cleanupSessions(sessionStore, appConfig.Session.CleanupIntervalHours, logger)

	lemonScheduler := lemon.NewScheduler(lemonStore, quizStore, logger, 1*time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := lemonScheduler.InitializeLemons(ctx); err != nil {
		logger.Printf("레몬 초기화 중 오류 발생: %v", err)
	}
	cancel()

	lemonScheduler.Start()

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

// TODO: 적절한 위치로 옮기기 잠시 여기에..
func validateInstanceName(name string) error {
	if name == "" {
		return errors.NewMissingParameterError("name")
	}

	if len(name) < 3 {
		return errors.NewInvalidParameterError("name", "3 글자 이상")
	}

	matched, _ := regexp.MatchString(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`, name)
	if !matched {
		return errors.NewInvalidParameterError("name",
			"인스턴스 이름은 소문자, 숫자, 하이픈(-)만 사용 가능하며, 시작과 끝은 영문자와 숫자만 가능")
	}

	if len(name) > 63 {
		return errors.NewInvalidParameterError("name", "인스턴스 이름은 63자 이하여야 합니다")
	}

	return nil
}
