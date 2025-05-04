package rest

import (
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/core/user"
	httputil "github.com/piper-hyowon/dBtree/internal/platform/rest"
	"log"
	"net/http"
)

type Handler struct {
	userService user.Service
	logger      *log.Logger // TODO: core.Logger 인터페이스 정의해서 사용
}

func NewHandler(userService user.Service, logger *log.Logger) *Handler {
	return &Handler{
		userService: userService,
		logger:      logger,
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, http.MethodDelete) {
		return
	}

	u := httputil.GetUserFromContext(r.Context())
	if u == nil {
		httputil.HandleError(w, errors.NewError(
			errors.ErrInternalServer,
			"인증 정보 없음", nil, nil), h.logger)
		return
	}

	err := h.userService.Delete(r.Context(), u.ID, u.Email)
	if err != nil {
		httputil.HandleError(w, err, h.logger)
		return
	}

	httputil.SendSuccessResponse(w, http.StatusNoContent, nil)
}
