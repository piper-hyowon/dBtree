package middleware

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const RequestIDKey contextKey = "requestID"

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	buffer     *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.buffer.Write(b)
	return r.ResponseWriter.Write(b)
}

func LoggingMiddleware(logger *log.Logger, debugLogging bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 요청 ID 생성
			requestID := generateRequestID()
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
			r = r.WithContext(ctx)

			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				buffer:         &bytes.Buffer{},
			}

			logger.Printf("[%s] → %s %s", requestID, r.Method, r.URL.String())

			// 요청 바디는 디버깅 모드에서만
			if debugLogging && r.ContentLength > 0 {
				if r.ContentLength < 1024*1024 {
					bodyBytes, _ := io.ReadAll(r.Body)
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

					if len(bodyBytes) > 0 {
						logger.Printf("[%s] REQUEST BODY: %s", requestID, string(bodyBytes))
					}
				} else {
					logger.Printf("[%s] REQUEST BODY: [TOO LARGE - %d bytes]", requestID, r.ContentLength)
				}
			}

			next.ServeHTTP(recorder, r)

			durationMs := float64(time.Since(start)) / float64(time.Millisecond)

			logger.Printf("[%s] ← %d (%d bytes, %.2f ms)",
				requestID, recorder.statusCode,
				recorder.buffer.Len(), durationMs)

			// 응답 바디(디버그 모드)
			if debugLogging {
				responseBody := recorder.buffer.String()
				responseBodyLength := recorder.buffer.Len()

				if responseBodyLength > 0 {
					const maxLogSize = 4096 // 4KB 제한
					if responseBodyLength > maxLogSize {
						logger.Printf("[%s] RESPONSE BODY: (First %d of %d bytes)%s...",
							requestID, maxLogSize, responseBodyLength, responseBody[:maxLogSize])
					} else {
						logger.Printf("[%s] RESPONSE BODY: %s", requestID, responseBody)
					}
				} else {
					logger.Printf("[%s] RESPONSE BODY: [EMPTY]", requestID)
				}
			}
		})
	}
}

func generateRequestID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func GetRequestIDFromContext(ctx context.Context) string {
	id, ok := ctx.Value(RequestIDKey).(string)
	if !ok {
		return "unknown"
	}
	return id
}
