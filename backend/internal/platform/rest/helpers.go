package rest

import (
	"encoding/json"
	"net/http"
)

func DecodeJSONRequest(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		SendErrorResponse(w, http.StatusBadRequest, "잘못된 요청 형식")
		return false
	}
	return true
}

func ValidateMethod(w http.ResponseWriter, r *http.Request, allowedMethod string) bool {
	if r.Method != allowedMethod {
		w.Header().Set("Allow", allowedMethod)
		SendErrorResponse(w, http.StatusMethodNotAllowed, "허용되지 않는 메서드")
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

	w.Header().Set("Allow", joinMethods(allowedMethods))
	SendErrorResponse(w, http.StatusMethodNotAllowed, "허용되지 않는 메서드")
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
