package email

import "context"

type Service interface {
	SendOTP(ctx context.Context, to string, code string) error
	SendWelcome(ctx context.Context, to string) error
	SendGoodbye(ctx context.Context, to string) error
	Send(ctx context.Context, to string, subject string, body string) error
	SendWithImages(ctx context.Context, to string, subject string, htmlBody string, images map[string][]byte) error
	Close() // 리소스 정리
}
