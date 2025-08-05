package rest

import (
	coredbservice "github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/platform/rest"
	"github.com/piper-hyowon/dBtree/internal/platform/validation"
	"log"
	"net/http"
)

type Handler struct {
	dbService coredbservice.Service
	logger    *log.Logger
}

func NewHandler(dbService coredbservice.Service, logger *log.Logger) *Handler {
	return &Handler{
		dbService: dbService,
		logger:    logger,
	}
}

func (h *Handler) CreateInstance(w http.ResponseWriter, r *http.Request) {
	user, err := rest.GetUserFromContext(r.Context())
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	var dto coredbservice.CreateInstanceRequest
	if !rest.DecodeJSONRequest(w, r, &dto, h.logger) {
		return
	}

	if err := validation.ValidateStruct(&dto); err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	if err := dto.Validate(); err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	resp, err := h.dbService.CreateInstance(r.Context(), user.ID, user.LemonBalance, &dto)
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendJSONResponse(w, http.StatusAccepted, resp)
}
