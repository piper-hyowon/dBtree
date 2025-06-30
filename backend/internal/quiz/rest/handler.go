package rest

import (
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
	"github.com/piper-hyowon/dBtree/internal/core/quiz"
	"github.com/piper-hyowon/dBtree/internal/platform/rest"
	"log"
	"net/http"
)

type Handler struct {
	quizService  quiz.Service
	lemonService lemon.Service
	logger       *log.Logger // TODO: core.Logger 인터페이스 정의해서 사용
}

func NewHandler(quizService quiz.Service, lemonService lemon.Service, logger *log.Logger) *Handler {
	return &Handler{
		quizService:  quizService,
		lemonService: lemonService,
		logger:       logger,
	}
}

func (h *Handler) StartQuiz(w http.ResponseWriter, r *http.Request, positionID int) {
	u := rest.GetUserFromContext(r.Context())
	if u == nil {
		rest.HandleError(w, errors.NewUnauthorizedError(), h.logger)
		return
	}

	// 유저 수확 쿨타임 체크
	availability, err := h.lemonService.CanHarvest(r.Context(), u.ID)
	if err != nil {
		rest.HandleError(w, errors.Wrap(err), h.logger)
		return
	}
	if !availability.CanHarvest {
		rest.HandleError(w, errors.NewHarvestCooldownError(availability.WaitTime), h.logger)
		return
	}

	res, err := h.quizService.StartQuiz(r.Context(), positionID, u.ID, u.Email)
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendSuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	u := rest.GetUserFromContext(r.Context())
	if u == nil {
		rest.HandleError(w, errors.NewUnauthorizedError(), h.logger)
		return
	}

	var req quiz.SubmitAnswerRequest
	if !rest.DecodeJSONRequest(w, r, &req, h.logger) {
		return
	}

	if req.AttemptID == nil {
		rest.HandleError(w, errors.NewMissingParameterError("attemptID"), h.logger)
		return
	}

	if *req.AttemptID <= 0 {
		rest.HandleError(w, errors.NewInvalidParameterError("attemptID", "attemptID - 양의 정수"), h.logger)
		return
	}

	res, err := h.quizService.SubmitAnswer(r.Context(), u.Email, *req.AttemptID, *req.OptionIdx)
	if err != nil {
		rest.HandleError(w, err, h.logger)
		return
	}

	rest.SendSuccessResponse(w, http.StatusOK, res)
}
