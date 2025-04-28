package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

func SendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("JSON 응답 인코딩 오류: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func SendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Error:   message,
	}); err != nil {
		log.Printf("JSON 오류 응답 인코딩 오류: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func SendSuccessResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(SuccessResponse{
		Success: true,
		Data:    data,
	}); err != nil {
		log.Printf("JSON 성공 응답 인코딩 오류: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func WithRetryAfter(w http.ResponseWriter, statusCode int, message string, seconds int) {
	w.Header().Set("Retry-After", fmt.Sprintf("%d", seconds))
	SendErrorResponse(w, statusCode, message)
}
