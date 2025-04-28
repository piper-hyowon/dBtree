package http

import (
	"encoding/json"
	"net/http"
)

// DecodeJSONRequest JSON 요청 본문을 지정된 구조체로 디코딩합니다.
// 실패 시 적절한 오류 응답을 보내고 false를 반환합니다.
func DecodeJSONRequest(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		SendErrorResponse(w, http.StatusBadRequest, "잘못된 요청 형식")
		return false
	}
	return true
}

// ValidateMethod HTTP 메서드가 허용된 메서드와 일치하는지 확인합니다.
// 불일치 시 적절한 오류 응답을 보내고 false를 반환합니다.
func ValidateMethod(w http.ResponseWriter, r *http.Request, allowedMethod string) bool {
	if r.Method != allowedMethod {
		w.Header().Set("Allow", allowedMethod)
		SendErrorResponse(w, http.StatusMethodNotAllowed, "허용되지 않는 메서드")
		return false
	}
	return true
}

// ValidateMethods HTTP 메서드가 허용된 메서드 목록에 포함되는지 확인합니다.
// 불일치 시 적절한 오류 응답을 보내고 false를 반환합니다.
func ValidateMethods(w http.ResponseWriter, r *http.Request, allowedMethods ...string) bool {
	for _, method := range allowedMethods {
		if r.Method == method {
			return true
		}
	}

	// 응답 헤더에 허용된 메서드 목록 설정
	w.Header().Set("Allow", joinMethods(allowedMethods))
	SendErrorResponse(w, http.StatusMethodNotAllowed, "허용되지 않는 메서드")
	return false
}

// 메서드 목록을 쉼표로 구분된 문자열로 변환
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
