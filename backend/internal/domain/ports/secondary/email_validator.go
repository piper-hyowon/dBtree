package secondary

type EmailValidator interface {
	Validate(email string, checkMX bool) (bool, error) // validated, error message
	IsDisposable(email string) bool                    // 일회용 이메일 확인
	HaxMXRecord(email string) bool
}
