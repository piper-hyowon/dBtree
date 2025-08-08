package rest

import (
	"github.com/piper-hyowon/dBtree/internal/common"
	"github.com/piper-hyowon/dBtree/internal/core/stats"
	"github.com/piper-hyowon/dBtree/internal/platform/rest"
	"log"
	"net/http"
)

type Handler struct {
	statsService stats.Service
	logger       *log.Logger
}

func NewHandler(statsService stats.Service, logger *log.Logger) *Handler {
	return &Handler{
		statsService: statsService,
		logger:       logger,
	}
}

func (h *Handler) GetGlobalStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.statsService.GetGlobalStats(r.Context())
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendSuccessResponse(w, http.StatusOK, stats)
}

func (h *Handler) GetMiniLeaderboard(w http.ResponseWriter, r *http.Request) {
	leaderboard, err := h.statsService.GetMiniLeaderboard(r.Context())
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendSuccessResponse(w, http.StatusOK, leaderboard)
}

func (h *Handler) GetUserDailyHarvest(w http.ResponseWriter, r *http.Request) {
	u, err := rest.GetUserFromContext(r.Context())
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	req := &stats.DailyHarvestRequest{
		Days: rest.GetIntQuery(r, "days", 7),
	}

	response, err := h.statsService.GetUserDailyHarvest(r.Context(), u.ID, req)
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendSuccessResponse(w, http.StatusOK, response)
}

func (h *Handler) GetUserTransactions(w http.ResponseWriter, r *http.Request) {
	u, err := rest.GetUserFromContext(r.Context())
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	req := &stats.TransactionsRequest{
		PaginationParams: common.PaginationParams{
			Page:  rest.GetIntQuery(r, "page", 1),
			Limit: rest.GetIntQuery(r, "limit", 31),
		},
		InstanceName: rest.GetStringQueryPtr(r, "instanceName"),
	}

	response, err := h.statsService.GetUserTransactions(r.Context(), u.ID, req)
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendSuccessResponse(w, http.StatusOK, response)
}

func (h *Handler) GetUserInstances(w http.ResponseWriter, r *http.Request) {
	u, err := rest.GetUserFromContext(r.Context())
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	response, err := h.statsService.GetUserInstances(r.Context(), u.ID)
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendSuccessResponse(w, http.StatusOK, response)
}
