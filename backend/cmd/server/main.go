package main

import (
	"fmt"
	"log"
	"net/http"
)

// 홈페이지 핸들러
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "안녕하세요! Go 웹서버에 오신 것을 환영합니다!")
}

// 사용자 정보 페이지 핸들러
func userHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "사용자 정보 페이지입니다")
}

// 커스텀 미들웨어 예제
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 요청 정보 로깅
		log.Printf("요청: %s %s %s", r.RemoteAddr, r.Method, r.URL)
		// 다음 핸들러로 요청 전달
		next.ServeHTTP(w, r)
	})
}

func main() {
	// 핸들러 등록
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/user", userHandler)

	// 미들웨어 적용
	handler := loggingMiddleware(mux)

	// 서버 설정
	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	fmt.Println("서버 시작: http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
