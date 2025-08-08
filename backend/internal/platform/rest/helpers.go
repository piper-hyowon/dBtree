package rest

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/piper-hyowon/dBtree/internal/core/errors"
)

func DecodeJSONRequest(w http.ResponseWriter, r *http.Request, v interface{}, logger *log.Logger) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		HandleError(w, errors.NewInvalidParameterError("JSON", "Error decoding JSON"), logger)
		return false
	}
	return true
}

func GetIntQuery(r *http.Request, key string, defaultVal int) int {
	if str := r.URL.Query().Get(key); str != "" {
		if val, err := strconv.Atoi(str); err == nil {
			return val
		}
	}
	return defaultVal
}

func GetStringQuery(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

func GetStringQueryPtr(r *http.Request, key string) *string {
	if str := r.URL.Query().Get(key); str != "" {
		return &str
	}
	return nil
}

func GetBoolQuery(r *http.Request, key string, defaultVal bool) bool {
	str := r.URL.Query().Get(key)
	if str == "" {
		return defaultVal
	}

	switch str {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return defaultVal
	}
}
