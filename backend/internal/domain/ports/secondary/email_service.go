package secondary

type EmailService interface {
	SendOTP(to string, code string) error
	SendWelcome(to string) error
}
