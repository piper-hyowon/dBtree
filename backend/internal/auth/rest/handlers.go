package rest

import (
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core/auth"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/core/user"
	"github.com/piper-hyowon/dBtree/internal/email"
	"github.com/piper-hyowon/dBtree/internal/platform/rest"
	"log"
	"net/http"
)

type Handler struct {
	authService    auth.Service
	logger         *log.Logger // TODO: core.Logger 인터페이스 정의해서 사용
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
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type SendOTPResponse struct {
	IsNewUser bool `json:"isNewUser"`
}

type VerifyOTPResponse struct {
	Profile   user.ProfileResponse `json:"profile"`
	Token     string               `json:"token,omitempty"`
	ExpiresIn int64                `json:"expiresIn,omitempty"`
}

func (h *Handler) SendOTP(w http.ResponseWriter, r *http.Request) {
	var req SendOTPRequest
	if !rest.DecodeJSONRequest(w, r, &req, h.logger) {
		return
	}

	if !h.validateEmail(w, req.Email, true) {
		return
	}

	// 세션 존재 여부 확인하여 첫 요청인지 재요청인지 판단
	session, err := h.authService.GetSession(r.Context(), req.Email)
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	var isNewUser bool

	if session == nil {
		// 첫 OTP 발송
		isNewUser, err = h.authService.StartAuth(r.Context(), req.Email)
	} else {
		// OTP 재발송
		err = h.authService.ResendOTP(r.Context(), req.Email)
	}

	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendSuccessResponse(w, http.StatusOK, SendOTPResponse{isNewUser})
}

func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest
	if !rest.DecodeJSONRequest(w, r, &req, h.logger) {
		return
	}

	if req.OTP == "" {
		rest.HandleError(w, errors.NewMissingParameterError("otp"), h.logger)
		return
	}

	if !h.validateEmail(w, req.Email, true) {
		return
	}

	u, token, err := h.authService.VerifyOTP(r.Context(), req.Email, req.OTP)
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	expiresIn := int64(auth.TokenExpirationHours * 3600)

	rest.SendSuccessResponse(w, http.StatusOK, VerifyOTPResponse{
		Profile:   u.ToProfileResponse(),
		Token:     token,
		ExpiresIn: expiresIn,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token := rest.GetTokenFromContext(r.Context())
	if token == "" {
		rest.HandleError(w, errors.NewInternalError(fmt.Errorf("토큰 정보를 불러올 수 없습니다")), h.logger)
		return
	}

	err := h.authService.Logout(r.Context(), token)
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendSuccessResponse(w, http.StatusOK, nil)
}

func (h *Handler) validateEmail(w http.ResponseWriter, email string, checkMX bool) bool {
	valid, validErr := h.emailValidator.Validate(email, checkMX)
	if !valid {
		rest.HandleError(w, validErr, h.logger)
		return false
	}
	return true
}
