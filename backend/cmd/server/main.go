package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
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

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "dBtree")
	})

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(appConfig.Server.Port),
		Handler:      mux,
		ReadTimeout:  time.Duration(appConfig.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(appConfig.Server.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(appConfig.Server.IdleTimeoutSeconds) * time.Second,
	}

	startServer(server)
}

func startServer(server *http.Server) {
	log.Printf("HTTP 서버 시작, 포트: %s\n", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}
