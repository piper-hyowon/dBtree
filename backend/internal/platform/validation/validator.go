package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
)

var (
	validate *validator.Validate
	once     sync.Once
)

func GetValidator() *validator.Validate {
	once.Do(func() {
		validate = validator.New()

		// JSON 태그를 필드 이름으로 사용
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		// 커스텀 validation 등록
		registerCustomValidations()
	})

	return validate
}

func registerCustomValidations() {
	// 인스턴스 이름: 쿠버네티스 리소스 이름 규칙
	// 소문자, 숫자, 하이픈만 허용, 시작과 끝은 영문자와 숫자만
	validate.RegisterValidation("instancename", func(fl validator.FieldLevel) bool {
		name := fl.Field().String()
		if name == "" {
			return true // required는 별도로 체크
		}
		matched, _ := regexp.MatchString(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`, name)
		return matched
	})

	// Cron 표현식 검증
	validate.RegisterValidation("cronschedule", func(fl validator.FieldLevel) bool {
		schedule := fl.Field().String()
		if schedule == "" {
			return true // required는 별도로 체크
		}

		// 표준 cron 형식 검증 (5개 필드: 분 시 일 월 요일)
		// 예: "0 2 * * *" (매일 새벽 2시)
		// 예: "*/30 * * * *" (30분마다)
		// 예: "0 0 * * 0" (매주 일요일 자정)
		parts := strings.Fields(schedule)
		if len(parts) != 5 {
			return false
		}

		// 각 필드의 유효성 검증
		for i, part := range parts {
			if !isValidCronField(part, i) {
				return false
			}
		}

		return true
	})
}

// isValidCronField cron 필드 유효성 검증 (TODO: ...임시
func isValidCronField(field string, position int) bool {
	if field == "*" {
		return true
	}

	// 스텝 값 (*/n)
	if strings.HasPrefix(field, "*/") {
		return true
	}

	// 범위 (n-m)
	if strings.Contains(field, "-") {
		return true
	}

	// 리스트 (n,m,o)
	if strings.Contains(field, ",") {
		return true
	}

	// 단일 숫자
	switch position {
	case 0: // 분 (0-59)
		return isInRange(field, 0, 59)
	case 1: // 시 (0-23)
		return isInRange(field, 0, 23)
	case 2: // 일 (1-31)
		return isInRange(field, 1, 31)
	case 3: // 월 (1-12)
		return isInRange(field, 1, 12)
	case 4: // 요일 (0-6, 0=일요일)
		return isInRange(field, 0, 6)
	}

	return false
}

func isInRange(field string, min, max int) bool {
	var num int
	_, err := fmt.Sscanf(field, "%d", &num)
	if err != nil {
		return false
	}
	return num >= min && num <= max
}

func ValidateStruct(s interface{}) error {
	if err := GetValidator().Struct(s); err != nil {
		return ParseValidationError(err)
	}
	return nil
}

func ParseValidationError(err error) error {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return errors.NewInvalidParameterError("request", "유효하지 않은 요청입니다")
	}

	// 첫 번째 에러만 처리
	if len(validationErrors) > 0 {
		e := validationErrors[0]
		field := e.Field()
		tag := e.Tag()
		param := e.Param()

		switch tag {
		case "required":
			return errors.NewMissingParameterError(field)

		case "min":
			if e.Type().Kind() == reflect.String {
				return errors.NewInvalidParameterError(field,
					fmt.Sprintf("최소 %s자 이상이어야 합니다", param))
			}
			return errors.NewInvalidParameterError(field,
				fmt.Sprintf("최소 %s 이상이어야 합니다", param))

		case "max":
			if e.Type().Kind() == reflect.String {
				return errors.NewInvalidParameterError(field,
					fmt.Sprintf("최대 %s자 이하여야 합니다", param))
			}
			return errors.NewInvalidParameterError(field,
				fmt.Sprintf("최대 %s 이하여야 합니다", param))

		case "instancename":
			return errors.NewInvalidParameterError(field,
				"소문자, 숫자, 하이픈(-)만 사용 가능하며, 시작과 끝은 영문자와 숫자만 가능합니다")

		case "cronschedule":
			return errors.NewInvalidParameterError(field,
				"올바른 cron 형식이 아닙니다 (예: '0 2 * * *')")

		case "email":
			return errors.NewInvalidParameterError(field,
				"올바른 이메일 형식이 아닙니다")

		case "url":
			return errors.NewInvalidParameterError(field,
				"올바른 URL 형식이 아닙니다")

		case "datetime":
			return errors.NewInvalidParameterError(field,
				fmt.Sprintf("올바른 날짜 형식이 아닙니다 (%s)", param))

		case "oneof":
			return errors.NewInvalidParameterError(field,
				fmt.Sprintf("다음 중 하나여야 합니다: %s", param))

		case "eqfield":
			return errors.NewInvalidParameterError(field,
				fmt.Sprintf("%s와 일치해야 합니다", param))

		case "required_if":
			parts := strings.Split(param, " ")
			if len(parts) >= 2 {
				return errors.NewInvalidParameterError(field,
					fmt.Sprintf("%s가 %s일 때 필수입니다", parts[0], parts[1]))
			}
			return errors.NewInvalidParameterError(field, "조건부 필수 필드입니다")

		case "dive":
			// 중첩된 구조체 validation 에러
			return errors.NewInvalidParameterError(field, "유효하지 않은 값입니다")

		default:
			return errors.NewInvalidParameterError(field, "유효하지 않은 값입니다")
		}
	}

	return errors.NewInvalidParameterError("request", "유효하지 않은 요청입니다")
}

// CollectAllValidationErrors 모든 validation error 수집
func CollectAllValidationErrors(err error) error {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return errors.NewInvalidParameterError("request", "유효하지 않은 요청입니다")
	}

	fieldErrors := make(map[string]string)

	for _, e := range validationErrors {
		field := e.Field()
		tag := e.Tag()
		param := e.Param()

		var message string
		switch tag {
		case "required":
			message = "필수 입력 항목입니다"
		case "min":
			if e.Type().Kind() == reflect.String {
				message = fmt.Sprintf("최소 %s자 이상이어야 합니다", param)
			} else {
				message = fmt.Sprintf("최소 %s 이상이어야 합니다", param)
			}
		case "max":
			if e.Type().Kind() == reflect.String {
				message = fmt.Sprintf("최대 %s자 이하여야 합니다", param)
			} else {
				message = fmt.Sprintf("최대 %s 이하여야 합니다", param)
			}
		case "instancename":
			message = "소문자, 숫자, 하이픈(-)만 사용 가능합니다"
		case "cronschedule":
			message = "올바른 cron 형식이 아닙니다 (예: '0 2 * * *')"
		case "email":
			message = "올바른 이메일 형식이 아닙니다"
		default:
			message = "유효하지 않은 값입니다"
		}

		fieldErrors[field] = message
	}

	if len(fieldErrors) > 0 {
		return errors.NewInvalidParameterError("validation", "입력값 검증 실패").
			WithData(map[string]interface{}{
				"fields": fieldErrors,
			})
	}

	return errors.NewInvalidParameterError("request", "유효하지 않은 요청입니다")
}
