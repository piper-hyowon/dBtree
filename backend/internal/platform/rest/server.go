package rest

import (
	"context"
	"github.com/piper-hyowon/dBtree/internal/platform/config"
	"github.com/piper-hyowon/dBtree/internal/platform/rest/middleware"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	*http.Server
	logger *log.Logger
}

func NewServer(appConfig *config.Config, router http.Handler, logger *log.Logger) *Server {
	loggingMiddleware := middleware.LoggingMiddleware(logger, appConfig.DebugLogging)
	corsMiddleware := middleware.NewCORSMiddleware(middleware.CORSConfig{
		AllowedOrigins:   appConfig.CORS.AllowedOrigins,
		AllowCredentials: appConfig.CORS.AllowCredentials,
	})

	handler := loggingMiddleware(corsMiddleware(router))

	return &Server{
		Server: &http.Server{
			Addr:         ":" + strconv.Itoa(appConfig.Server.Port),
			Handler:      handler,
			ReadTimeout:  time.Duration(appConfig.Server.ReadTimeoutSeconds) * time.Second,
			WriteTimeout: time.Duration(appConfig.Server.WriteTimeoutSeconds) * time.Second,
			IdleTimeout:  time.Duration(appConfig.Server.IdleTimeoutSeconds) * time.Second,
		},
		logger: logger,
	}
}

func (s *Server) Start() error {
	s.logger.Printf("HTTP 서버 시작, 포트: %s\n", s.Addr)
	return s.ListenAndServe()
}

func (s *Server) GracefulShutdown(timeout time.Duration) error {
	s.logger.Println("서버 종료 중...")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.Shutdown(ctx)
}
