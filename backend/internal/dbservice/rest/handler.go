package rest

import (
	coredbservice "github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/platform/rest"
	"github.com/piper-hyowon/dBtree/internal/platform/rest/router"
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

func (h *Handler) GetInstanceWithSync(w http.ResponseWriter, r *http.Request) {
	user, err := rest.GetUserFromContext(r.Context())
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	id := router.Param(r, "id")
	if id == "" {
		rest.HandleError(w, errors.NewMissingParameterError("id"), h.logger)
		return
	}

	instance, err := h.dbService.GetInstanceWithSync(r.Context(), user.ID, id)
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	response := instance.ToResponse()

	rest.SendSuccessResponse(w, http.StatusOK, response)
}

func (h *Handler) ListPresets(w http.ResponseWriter, r *http.Request) {
	presets, err := h.dbService.ListPresets(r.Context())
	if err != nil {
		rest.HandleError(w, err, h.logger)
	}

	responses := make([]coredbservice.PresetResponse, len(presets))
	for i, preset := range presets {
		responses[i] = preset.ToResponse()
	}

	rest.SendSuccessResponse(w, http.StatusOK, responses)
}
