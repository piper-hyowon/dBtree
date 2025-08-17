package errors

import "fmt"

// NewResourceExhaustedError 시스템 리소스 부족
func NewResourceExhaustedError(resourceType string, available, requested interface{}) DomainError {
	return NewError(
		ErrResourceExhausted,
		fmt.Sprintf("%s 리소스 부족 (가용: %v, 요청: %v)",
			resourceType, available, requested),
		map[string]interface{}{
			"resourceType": resourceType,
			"available":    available,
			"requested":    requested,
		},
		nil,
	)
}

// NewSystemCapacityError 시스템 용량 초과
func NewSystemCapacityError(message string) DomainError {
	return NewError(
		ErrSystemCapacity,
		message,
		nil,
		nil,
	)
}
