package secondary

import "context"

type EmailService interface {
	SendOTP(ctx context.Context, to string, code string) error
	SendWelcome(ctx context.Context, to string) error
	Close() // 리소스 정리
}
