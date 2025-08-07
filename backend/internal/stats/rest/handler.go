package rest

import (
	"log"
	"net/http"

	corestats "github.com/piper-hyowon/dBtree/internal/core/stats"
	"github.com/piper-hyowon/dBtree/internal/platform/rest"
)

type Handler struct {
	statsService corestats.Service
	logger       *log.Logger
}

func NewHandler(statsService corestats.Service, logger *log.Logger) *Handler {
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
