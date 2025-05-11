package rest

import (
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"log"
	"net/http"

	stdErrors "errors"
)

func errorCodeToStatusCode(code errors.ErrorCode) int {
	switch code {
	case errors.ErrInvalidParameter, errors.ErrMissingParameter:
		return http.StatusBadRequest

	case errors.ErrMethodNotAllowed:
		return http.StatusMethodNotAllowed

	case errors.ErrInvalidOTP, errors.ErrExpiredOTP, errors.ErrSessionNotFound,
		errors.ErrInvalidToken, errors.ErrUnauthorized:
		return http.StatusUnauthorized

	case errors.ErrSessionExpired, errors.ErrAlreadyVerified,
		errors.ErrResourceConflict, errors.ErrInsufficientLemons, errors.ErrHarvestCooldown,
		errors.ErrLemonStorageFull, errors.ErrNoQuizInProgress, errors.ErrHarvestAlreadyProcessed,
		errors.ErrLemonAlreadyHarvested:
		return http.StatusConflict

	case errors.ErrResourceNotFound:
		return http.StatusNotFound

	case errors.ErrTooManyResends, errors.ErrTooEarlyResend:
		return http.StatusTooManyRequests

	case errors.ErrInternalServer:
		return http.StatusInternalServerError

	default:
		return http.StatusInternalServerError
	}
}

func HandleError(w http.ResponseWriter, err error, logger *log.Logger) {
	var domainErr errors.DomainError

	if !stdErrors.As(err, &domainErr) {
		// stack 을 알 수 없음..(return errors.ErrInternalServer 가 아니라 그냥 return err 한 곳 인듯)
		// TODO: 정적 검사 필요
		SendJSONResponse(w, http.StatusInternalServerError, ErrorResponse{
			Code:    int(errors.ErrInternalServer),
			Message: "알수 없는 오류가 발생했습니다",
		})
		return
	}

	logger.Printf("도메인 에러: %s (코드: %d)", domainErr.Error(), domainErr.Code())
	var message string

	if domainErr.Code() == errors.ErrInternalServer {
		if cause := errors.Unwrap(domainErr); cause != nil {
			logger.Printf("에러: %v", cause)
		}
		logger.Printf("스택 트레이스:\n%s", domainErr.Stack())
		message = "서버 내부 오류가 발생했습니다"
	} else {
		message = domainErr.Error()

	}

	if domainErr.Code() == errors.ErrTooEarlyResend || domainErr.Code() == errors.ErrTooManyResends {
		if data := domainErr.ErrorData(); data != nil {
			if m, ok := data.(map[string]int); ok {
				if seconds, ok := m["waitSeconds"]; ok {
					w.Header().Set("Retry-After", fmt.Sprintf("%d", seconds))
				} else if maxResends, ok := m["maxResends"]; ok {
					w.Header().Set("Retry-After", fmt.Sprintf("%d", maxResends))
				}
			}
		}
	}

	response := ErrorResponse{
		Code:    int(domainErr.Code()),
		Message: "[" + domainErr.Code().String() + "] " + message,
	}

	if data := domainErr.ErrorData(); data != nil {
		response.Data = data
	}

	SendJSONResponse(w, errorCodeToStatusCode(domainErr.Code()), response)
}
