package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/piper-hyowon/dBtree/internal/adapters/primary/core"
	"github.com/piper-hyowon/dBtree/internal/domain/model"
)

type Handler struct {
	authService *core.AuthService
	logger      *log.Logger
}

func NewHandler(authService *core.AuthService, logger *log.Logger) *Handler {
	return &Handler{
		authService: authService,
		logger:      logger,
	}
}

type SendOTPRequest struct {
	Email string `json:"email"`
}

type VerifyOTPRequest struct {
	Email   string `json:"email"`
	OTPCode string `json:"otpCode"`
}

type SendOTPResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	IsNewUser bool   `json:"isNewUser"`
}

type VerifyOTPResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	User    *model.User `json:"user,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func (h *Handler) SendOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	otpType := r.URL.Query().Get("type")
	if otpType != "authentication" {
		http.Error(w, "Invalid OTP type", http.StatusBadRequest)
		return
	}

	var req SendOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "잘못된 요청 형식")
		return
	}

	// 세션 존재 여부 확인하여 첫 요청인지 재요청인지 판단
	session, err := h.authService.GetSession(r.Context(), req.Email)

	var isNewUser bool
	var responseMsg string

	if err != nil || session == nil {
		// 첫 OTP 발송
		isNewUser, err = h.authService.StartAuth(r.Context(), req.Email)
		responseMsg = "인증 코드가 이메일로 전송되었습니다."
	} else {
		// OTP 재발송
		err = h.authService.ResendOTP(r.Context(), req.Email)
		responseMsg = "인증 코드가 이메일로 재전송되었습니다."
	}

	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	h.sendJSONResponse(w, http.StatusOK, SendOTPResponse{
		Success:   true,
		Message:   responseMsg,
		IsNewUser: isNewUser,
	})
}

func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "잘못된 요청 형식")
		return
	}

	user, err := h.authService.VerifyOTP(r.Context(), req.Email, req.OTPCode)
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	h.sendJSONResponse(w, http.StatusOK, VerifyOTPResponse{
		Success: true,
		Message: "인증이 완료되었습니다.",
		User:    user,
	})
}

func (h *Handler) handleAuthError(w http.ResponseWriter, err error) {
	var statusCode int
	var message string

	switch {
	case errors.Is(err, core.ErrInvalidEmail):
		statusCode = http.StatusBadRequest
		message = "유효하지 않은 이메일 주소입니다."
	case errors.Is(err, core.ErrTooManyResends):
		statusCode = http.StatusTooManyRequests
		message = "OTP 전송 횟수 제한에 도달했습니다. 나중에 다시 시도해주세요."
	case errors.Is(err, core.ErrTooEarlyResend):
		statusCode = http.StatusTooEarly
		message = "OTP 재전송은 1분 후에 가능합니다."
	case errors.Is(err, core.ErrInvalidOTP):
		statusCode = http.StatusBadRequest
		message = "유효하지 않은 인증 코드입니다."
	case errors.Is(err, core.ErrExpiredOTP):
		statusCode = http.StatusBadRequest
		message = "만료된 인증 코드입니다. 새 인증 코드를 요청해주세요."
	case errors.Is(err, core.ErrSessionNotFound):
		statusCode = http.StatusNotFound
		message = "세션을 찾을 수 없습니다."
	case errors.Is(err, core.ErrInternal):
		statusCode = http.StatusInternalServerError
		message = "내부 서버 오류"
		h.logger.Printf("내부 오류: %v", err)
	default:
		statusCode = http.StatusInternalServerError
		message = "알 수 없는 오류"
		h.logger.Printf("알 수 없는 오류: %v", err)
	}

	h.sendErrorResponse(w, statusCode, message)
}

func (h *Handler) sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Printf("JSON 응답 인코딩 오류: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	h.sendJSONResponse(w, statusCode, ErrorResponse{
		Success: false,
		Error:   message,
	})
}
