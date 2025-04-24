package secondary

type Validator interface {
	Validate(email string, checkMX bool) (bool, error) // validated
	IsDisposable(email string) bool                    // 일회용 이메일 확인
	HasMXRecord(email string) bool
}
