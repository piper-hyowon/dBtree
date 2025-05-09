package rest

import (
	"log"
	"net/http"
)

type Handler struct {
	logger *log.Logger // TODO: core.Logger 인터페이스 정의해서 사용
}

func NewHandler(logger *log.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

func (h *Handler) StartQuizWithPosition(w http.ResponseWriter, r *http.Request, positionID int) {}

func (h *Handler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {}
