package rest

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/piper-hyowon/dBtree/internal/core/errors"
)

func DecodeJSONRequest(w http.ResponseWriter, r *http.Request, v interface{}, logger *log.Logger) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		domainErr := errors.NewError(
			errors.ErrInvalidParameter,
			"잘못된 요청 형식",
			nil,
			err,
		)
		HandleError(w, domainErr, logger)
		return false
	}
	return true
}
