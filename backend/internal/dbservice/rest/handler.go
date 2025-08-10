package rest

import (
	"fmt"
	coredbservice "github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/platform/rest"
	"github.com/piper-hyowon/dBtree/internal/platform/rest/router"
	"github.com/piper-hyowon/dBtree/internal/platform/validation"
	"log"
	"net/http"
)

type Handler struct {
	publicHost string
	dbService  coredbservice.Service
	portStore  coredbservice.PortStore
	logger     *log.Logger
}

func NewHandler(
	publicHost string, dbService coredbservice.Service, portStore coredbservice.PortStore, logger *log.Logger,
) *Handler {
	return &Handler{
		publicHost: publicHost,
		dbService:  dbService,
		portStore:  portStore,
		logger:     logger,
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
	if port, err := h.portStore.GetPort(r.Context(), instance.ExternalID); err == nil && port > 0 {
		response.ExternalHost = h.publicHost
		response.ExternalPort = port

		// URI 템플릿 직접 생성
		if instance.Type == coredbservice.MongoDB {
			response.ExternalURITemplate = fmt.Sprintf(
				"mongodb://{USERNAME}:{PASSWORD}@%s:%d/%s?authSource=admin",
				h.publicHost, port, instance.Name)
		} else if instance.Type == coredbservice.Redis {
			response.ExternalURITemplate = fmt.Sprintf(
				"redis://:{PASSWORD}@%s:%d",
				h.publicHost, port)
		}
	}
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

func (h *Handler) ListInstances(w http.ResponseWriter, r *http.Request) {
	user, err := rest.GetUserFromContext(r.Context())
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	instances, err := h.dbService.ListInstances(r.Context(), user.ID)
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}
	res := make([]coredbservice.InstanceResponse, 0, len(instances))
	for _, v := range instances {
		res = append(res, *v.ToResponse())
	}

	rest.SendSuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) DeleteInstance(w http.ResponseWriter, r *http.Request) {
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
	if err := h.dbService.DeleteInstance(r.Context(), user.ID, id); err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendSuccessResponse(w, http.StatusNoContent, nil)
}

func (h *Handler) UpdateInstanceStatus(w http.ResponseWriter, r *http.Request) {
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

	newStatus := router.Param(r, "status")
	if newStatus == "" {
		rest.HandleError(w, errors.NewEndpointNotFoundError(r.RequestURI), h.logger)
		return
	}

	switch newStatus {
	case "stop":
		if err := h.dbService.StopInstance(r.Context(), user.ID, id); err != nil {
			rest.HandleError(w, err, h.logger)
			return
		}
	case "start":
		if err := h.dbService.StartInstance(r.Context(), user.ID, id); err != nil {
			rest.HandleError(w, err, h.logger)
			return
		}

	case "restart":
		if err := h.dbService.RestartInstance(r.Context(), user.ID, id); err != nil {
			rest.HandleError(w, err, h.logger)
			return
		}
	}

	rest.SendSuccessResponse(w, http.StatusNoContent, nil)
}
