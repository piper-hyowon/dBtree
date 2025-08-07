package string

import "strings"

func MaskString(s string, keepStart int, maskChar string) string {
	if len(s) <= keepStart {
		return s + maskChar
	}
	return s[:keepStart] + maskChar
}

func MaskEmail(email string, keepLocal int, maskChar string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return maskChar
	}

	local := parts[0]
	if len(local) <= keepLocal {
		return local + maskChar
	}

	return local[:keepLocal] + maskChar
}
