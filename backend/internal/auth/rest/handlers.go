package rest

import (
	"errors"
	"github.com/piper-hyowon/dBtree/internal/auth"
	"github.com/piper-hyowon/dBtree/internal/common"
	"github.com/piper-hyowon/dBtree/internal/email"
	httputil "github.com/piper-hyowon/dBtree/internal/platform/rest"
	"github.com/piper-hyowon/dBtree/internal/platform/rest/middleware"
	"github.com/piper-hyowon/dBtree/internal/user"
	"log"
	"net/http"
	"runtime/debug"
)

type Handler struct {
	authService    auth.Service
	logger         *log.Logger // TODO: common.Logger 인터페이스 정의해서 사용
	emailValidator email.Validator
}

func NewHandler(authService auth.Service, logger *log.Logger) *Handler {
	validator, err := email.NewValidator()

	if err != nil {
		logger.Printf("일회용 이메일 검사 생략: %v, 정규식, MX만 검사가능", err)
	}

	return &Handler{
		authService:    authService,
		logger:         logger,
		emailValidator: validator,
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
	IsNewUser bool `json:"isNewUser"`
}

type VerifyOTPResponse struct {
	User      *user.User `json:"user,omitempty"`
	Token     string     `json:"token,omitempty"`
	ExpiresIn int64      `json:"expires_in,omitempty"`
}

func (h *Handler) SendOTP(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, http.MethodPost) {
		return
	}

	var req SendOTPRequest
	if !httputil.DecodeJSONRequest(w, r, &req) {
		return
	}

	if !h.validateEmail(w, req.Email, true) {
		return
	}

	// 세션 존재 여부 확인하여 첫 요청인지 재요청인지 판단
	session, err := h.authService.GetSession(r.Context(), req.Email)

	var isNewUser bool

	if err != nil || session == nil {
		// 첫 OTP 발송
		isNewUser, err = h.authService.StartAuth(r.Context(), req.Email)
	} else {
		// OTP 재발송
		err = h.authService.ResendOTP(r.Context(), req.Email)
	}

	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	httputil.SendSuccessResponse(w, http.StatusOK, SendOTPResponse{isNewUser})
}

func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest
	if !httputil.ValidateMethod(w, r, http.MethodPost) {
		return
	}

	if !httputil.DecodeJSONRequest(w, r, &req) {
		return
	}

	if !h.validateEmail(w, req.Email, true) {
		return
	}

	u, token, err := h.authService.VerifyOTP(r.Context(), req.Email, req.OTPCode)
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	expiresIn := int64(common.TokenExpirationHours * 3600)

	httputil.SendSuccessResponse(w, http.StatusOK, VerifyOTPResponse{
		User:      u,
		Token:     token,
		ExpiresIn: expiresIn,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, http.MethodPost) {
		return
	}

	token := middleware.GetTokenFromContext(r.Context())
	if token == "" {
		httputil.SendErrorResponse(w, http.StatusInternalServerError, "토큰 정보를 불러올 수 없습니다")
		return
	}

	err := h.authService.Logout(r.Context(), token)
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	httputil.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func (h *Handler) handleAuthError(w http.ResponseWriter, err error) {
	var statusCode int
	var message string
	var retryAfter int

	switch {
	case errors.Is(err, common.ErrInvalidEmail):
		statusCode = http.StatusBadRequest
		message = "유효하지 않은 이메일 주소입니다."
	case errors.Is(err, common.ErrTooManyResends):
		statusCode = http.StatusTooManyRequests
		message = "OTP 전송 횟수 제한에 도달했습니다. 나중에 다시 시도해주세요."
		// MaxResendAttempts에 도달한 경우 OTP 만료 시간 이후 새 세션 시작 가능
		retryAfter = common.OTPExpirationMinutes * 60
	case errors.Is(err, common.ErrTooEarlyResend):
		statusCode = http.StatusTooManyRequests
		message = "OTP 재전송은 1분 후에 가능합니다."
		retryAfter = common.ResendWaitSeconds
	case errors.Is(err, common.ErrInvalidOTP):
		statusCode = http.StatusUnauthorized
		message = "유효하지 않은 인증 코드입니다."
	case errors.Is(err, common.ErrExpiredOTP):
		statusCode = http.StatusUnauthorized
		message = "만료된 인증 코드입니다. 새 인증 코드를 요청해주세요."
	case errors.Is(err, common.ErrSessionAlreadyVerified):
		statusCode = http.StatusBadRequest
		message = "이미 인증이 완료된 세션입니다"
	case errors.Is(err, common.ErrSessionNotFound):
		statusCode = http.StatusUnauthorized
		message = "세션을 찾을 수 없습니다."
	case errors.Is(err, common.ErrInternal):
		statusCode = http.StatusInternalServerError
		message = "내부 서버 오류"
		h.logger.Printf("내부 오류: %v", err)
		h.logger.Printf("%s", debug.Stack())
	default:
		statusCode = http.StatusInternalServerError
		message = "알 수 없는 오류"
		h.logger.Printf("알 수 없는 오류: %v", err)
		h.logger.Printf("%s", debug.Stack())
	}

	// Retry-After 헤더 추가
	if retryAfter > 0 {
		httputil.WithRetryAfter(w, statusCode, message, retryAfter)
		return
	}

	httputil.SendErrorResponse(w, statusCode, message)
}

func (h *Handler) validateEmail(w http.ResponseWriter, email string, checkMX bool) bool {
	valid, err := h.emailValidator.Validate(email, checkMX)
	if !valid {
		if err != nil {
			httputil.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		} else {
			httputil.SendErrorResponse(w, http.StatusBadRequest, "유효하지 않은 이메일")
		}
		return false
	}
	return true
}
