package email

import (
	"bufio"
	"github.com/piper-hyowon/dBtree/internal/core"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
)

type Validator interface {
	Validate(email string, checkMX bool) (bool, error) // validated
	IsDisposable(email string) bool                    // 일회용 이메일 확인
	HasMXRecord(email string) bool
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
var blacklistPath = "internal/resources/disposable_domains.txt"

type validator struct {
	disposableDomains map[string]bool
	blocklist         map[string]bool
	mu                sync.RWMutex
}

var _ Validator = (*validator)(nil)

func NewValidator() (Validator, error) {
	v := &validator{
		disposableDomains: make(map[string]bool),
		blocklist:         make(map[string]bool),
	}

	err := v.loadDomainsFromFile()
	if err != nil {
		return v, err
	}

	return v, nil
}

func (v *validator) loadDomainsFromFile() error {
	file, err := os.Open(blacklistPath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(blacklistPath, []byte{}, 0644)
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			v.disposableDomains[line] = true
		}
	}

	return scanner.Err()
}
func (v *validator) IsValidBasicFormat(email string) bool {
	return email != "" && emailRegex.MatchString(email)
}

func (v *validator) IsDisposable(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := strings.ToLower(parts[1])

	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.disposableDomains[domain] || v.blocklist[domain]
}

func (v *validator) HasMXRecord(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := parts[1]
	mxRecords, err := net.LookupMX(domain)
	return err == nil && len(mxRecords) > 0
}

func (v *validator) AddToBlacklist(domain string) error {
	domain = strings.ToLower(domain)

	v.mu.Lock()
	defer v.mu.Unlock()

	// 이미 블랙리스트
	if v.disposableDomains[domain] {
		return nil
	}

	v.disposableDomains[domain] = true

	file, err := os.OpenFile(blacklistPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(domain + "\n")
	return err
}

func (v *validator) Validate(email string, checkMX bool) (bool, error) {
	// 1차: 정규식 검사
	if !v.IsValidBasicFormat(email) {
		return false, core.NewEmailValidationError(" 이메일 형식이 올바르지 않습니다")
	}

	// 2단계: 일회용 이메일 검사
	if v.IsDisposable(email) {
		return false, core.NewEmailValidationError("일회용 이메일은 사용할 수 없습니다")

	}

	// 3단계: MX 레코드 검사
	if checkMX {
		hasMX := v.HasMXRecord(email)
		if !hasMX {
			parts := strings.Split(email, "@")
			if len(parts) == 2 {
				domain := parts[1]
				_ = v.AddToBlacklist(domain)
			}

			return false, core.NewEmailValidationError("유효하지 않은 이메일 도메인입니다")
		}
	}

	return true, nil
}
