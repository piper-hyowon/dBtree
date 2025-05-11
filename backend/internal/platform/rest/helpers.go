package rest

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/piper-hyowon/dBtree/internal/core/errors"
)

func DecodeJSONRequest(w http.ResponseWriter, r *http.Request, v interface{}, logger *log.Logger) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		HandleError(w, errors.NewInvalidParameterError("JSON", "Error decoding JSON"), logger)
		return false
	}
	return true
}
