package rest

import (
	"log"
	"net/http"

	coreresource "github.com/piper-hyowon/dBtree/internal/core/resource"
	"github.com/piper-hyowon/dBtree/internal/platform/rest"
)

type Handler struct {
	resourceManager coreresource.Manager
	logger          *log.Logger
}

func NewHandler(resourceManager coreresource.Manager, logger *log.Logger) *Handler {
	return &Handler{
		resourceManager: resourceManager,
		logger:          logger,
	}
}

func (h *Handler) GetSystemResources(w http.ResponseWriter, r *http.Request) {
	status, err := h.resourceManager.GetStatus(r.Context())
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendSuccessResponse(w, http.StatusOK, status)
}
