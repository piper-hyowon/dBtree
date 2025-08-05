package rest

import (
	"encoding/json"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func SendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("JSON 응답 인코딩 오류: %v", err)

		jsonBytes, _ := json.Marshal(ErrorResponse{
			Code:    int(errors.ErrInternalServer),
			Message: "서버 내부 오류가 발생했습니다",
		})
		w.Write(jsonBytes)
	}
}

func SendSuccessResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	SendJSONResponse(w, statusCode, data)
}
