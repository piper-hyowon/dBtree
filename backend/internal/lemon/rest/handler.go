package rest

import (
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
	"github.com/piper-hyowon/dBtree/internal/platform/rest"
	"log"
	"net/http"
)

type Handler struct {
	lemonService lemon.Service
	logger       *log.Logger // TODO: core.Logger 인터페이스 정의해서 사용
}

func NewHandler(lemonService lemon.Service, logger *log.Logger) *Handler {
	return &Handler{
		lemonService: lemonService,
		logger:       logger,
	}
}

func (h *Handler) TreeStatus(w http.ResponseWriter, r *http.Request) {
	t, err := h.lemonService.TreeStatus(r.Context())
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendSuccessResponse(w, http.StatusOK, t)
}

func (h *Handler) CanHarvest(w http.ResponseWriter, r *http.Request) {
	u := rest.GetUserFromContext(r.Context())
	if u == nil {
		//rest.HandleError(w, errors.NewError(
		//	errors.ErrInternalServer,
		//	"인증 정보 없음", nil, nil), h.logger)
		rest.HandleError(w, errors.NewUnauthorizedError(), h.logger)
		return
	}

	result, err := h.lemonService.CanHarvest(r.Context(), u.ID)
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendSuccessResponse(w, http.StatusOK, result)
}

func (h *Handler) HarvestLemon(w http.ResponseWriter, r *http.Request) {
	var req lemon.HarvestRequest

	if !rest.DecodeJSONRequest(w, r, &req, h.logger) {
		return
	}

	if req.PositionID == nil {
		rest.HandleError(w, errors.NewMissingParameterError("positionID"), h.logger)
		return
	}

	if req.AttemptID == nil {
		rest.HandleError(w, errors.NewMissingParameterError("attemptID"), h.logger)
		return
	}

	if *req.PositionID < 0 || *req.PositionID > lemon.DefaultRegrowthRules.MaxPositions {
		rest.HandleError(w, errors.NewInvalidParameterError("positionID", fmt.Sprintf("positionID 는 %d 부터 %d 사이", 0, lemon.DefaultRegrowthRules.MaxPositions)), h.logger)
		return
	}

	if *req.AttemptID <= 0 {
		rest.HandleError(w, errors.NewInvalidParameterError("attemptID", "attemptID - 양의 정수"), h.logger)
		return
	}

	u := rest.GetUserFromContext(r.Context())
	if u == nil {
		rest.HandleError(w, errors.NewUnauthorizedError(), h.logger)
		return
	}

	result, err := h.lemonService.HarvestLemon(r.Context(), u.ID, *req.PositionID, *req.AttemptID)
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendSuccessResponse(w, http.StatusOK, result)
}
