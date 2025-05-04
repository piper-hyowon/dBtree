package rest

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/piper-hyowon/dBtree/internal/core/errors"
)

func DecodeJSONRequest(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		domainErr := errors.NewError(
			errors.ErrInvalidParameter,
			"잘못된 요청 형식",
			nil,
			err,
		)
		HandleError(w, domainErr, log.Default())
		return false
	}
	return true
}

func ValidateMethod(w http.ResponseWriter, r *http.Request, allowedMethod string) bool {
	if r.Method != allowedMethod {
		w.Header().Set("Allow", allowedMethod)
		HandleError(w, errors.NewMethodNotAllowedError(allowedMethod), log.Default())
		return false
	}
	return true
}

func ValidateMethods(w http.ResponseWriter, r *http.Request, allowedMethods ...string) bool {
	for _, method := range allowedMethods {
		if r.Method == method {
			return true
		}
	}

	allowed := joinMethods(allowedMethods)
	w.Header().Set("Allow", allowed)

	HandleError(w, errors.NewMethodNotAllowedError(allowed), log.Default())
	return false
}

func joinMethods(methods []string) string {
	if len(methods) == 0 {
		return ""
	}

	result := methods[0]
	for i := 1; i < len(methods); i++ {
		result += ", " + methods[i]
	}
	return result
}
