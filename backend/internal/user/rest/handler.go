package rest

import (
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
	"github.com/piper-hyowon/dBtree/internal/core/user"
	"github.com/piper-hyowon/dBtree/internal/platform/rest"
	"log"
	"net/http"
)

type Handler struct {
	userService user.Service
	lemonStore  lemon.Store
	logger      *log.Logger // TODO: core.Logger 인터페이스 정의해서 사용
}

func NewHandler(userService user.Service, lemonStore lemon.Store, logger *log.Logger) *Handler {
	return &Handler{
		lemonStore:  lemonStore,
		userService: userService,
		logger:      logger,
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	u, err := rest.GetUserFromContext(r.Context())
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	err = h.userService.Delete(r.Context(), u.ID, u.Email)
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendSuccessResponse(w, http.StatusNoContent, nil)
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	u, err := rest.GetUserFromContext(r.Context())
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendSuccessResponse(w, http.StatusOK, u.ToProfileResponse())
}
