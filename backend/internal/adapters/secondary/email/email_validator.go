package email

import (
	"bufio"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
var blacklistPath = "internal/resources/disposable_domains.txt"

type Validator struct {
	disposableDomains map[string]bool
	blocklist         map[string]bool
	mu                sync.RWMutex
}

func NewValidator() (*Validator, error) {
	v := &Validator{
		disposableDomains: make(map[string]bool),
	}

	if err := v.loadDomainsFromFile(); err != nil {
		return nil, err
	}

	return v, nil
}

func (v *Validator) loadDomainsFromFile() error {
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
func (v *Validator) IsValidBasicFormat(email string) bool {
	return email != "" && emailRegex.MatchString(email)
}

func (v *Validator) IsDisposable(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := strings.ToLower(parts[1])

	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.disposableDomains[domain] || v.blocklist[domain]
}

func (v *Validator) HasMXRecord(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := parts[1]
	mxRecords, err := net.LookupMX(domain)
	return err == nil && len(mxRecords) > 0
}

func (v *Validator) AddToBlacklist(domain string) error {
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

func (v *Validator) Validate(email string, checkMX bool) (bool, string) {
	// 1차: 정규식 검사
	if !v.IsValidBasicFormat(email) {
		return false, "이메일 형식이 올바르지 않습니다"
	}

	// 2단계: 일회용 이메일 검사
	if v.IsDisposable(email) {
		return false, "일회용 이메일은 사용할 수 없습니다"
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

			return false, "유효하지 않은 이메일 도메인입니다"
		}
	}

	return true, ""
}
