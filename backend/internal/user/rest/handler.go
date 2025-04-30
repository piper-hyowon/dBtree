package rest

import (
	"github.com/piper-hyowon/dBtree/internal/core/user"
	httputil "github.com/piper-hyowon/dBtree/internal/platform/rest"
	"github.com/piper-hyowon/dBtree/internal/platform/rest/middleware"
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

	u := middleware.GetUserFromContext(r.Context())
	if u == nil {
		httputil.SendErrorResponse(w, http.StatusInternalServerError, "서버 에러")
		return
	}

	err := h.userService.Delete(r.Context(), u.ID)
	if err != nil {
		httputil.SendErrorResponse(w, http.StatusInternalServerError, "유저 탈퇴 실패")
		return
	}

	httputil.SendSuccessResponse(w, http.StatusNoContent, nil)
}
