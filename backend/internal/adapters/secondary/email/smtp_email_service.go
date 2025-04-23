package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/piper-hyowon/dBtree/internal/constants"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type SMTPEmailService struct {
	config SMTPConfig
}

func NewSMTPEmailService(config SMTPConfig) *SMTPEmailService {
	return &SMTPEmailService{
		config: config,
	}
}

func (s *SMTPEmailService) SendOTP(ctx context.Context, to string, otpCode string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("수신자 이메일 주소가 비어 있습니다")
	}

	if otpCode == "" {
		return fmt.Errorf("OTP 코드가 비어 있습니다")
	}

	subject := "dBtree 인증 코드"
	htmlBody := fmt.Sprintf(emailTemplateOTP, otpCode, constants.OTPExpirationMinutes)

	message := "From: " + s.config.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		htmlBody

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	client, err := createSMTPClient(addr, s.config.Host, s.config.Username, s.config.Password)
	if err != nil {
		return fmt.Errorf("SMTP 클라이언트 생성 실패: %w", err)
	}
	defer client.Close()

	if err := client.Mail(s.config.From); err != nil {
		return fmt.Errorf("발신자 설정 실패: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("수신자 설정 실패: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("데이터 쓰기 준비 실패: %w", err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("이메일 내용 쓰기 실패: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("데이터 쓰기 완료 실패: %w", err)
	}

	return nil
}

func createSMTPClient(addr, host, username, password string) (*smtp.Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	client, err := smtp.Dial(addr)
	if err != nil {
		return nil, err
	}

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err = client.StartTLS(tlsConfig); err != nil {
			client.Close()
			return nil, fmt.Errorf("STARTTLS 시작 실패: %w", err)
		}
	}

	if ok, _ := client.Extension("AUTH"); ok && username != "" && password != "" {
		authApi := smtp.PlainAuth("", username, password, host)
		if err = client.Auth(authApi); err != nil {
			client.Close()
			return nil, fmt.Errorf("SMTP 인증 실패: %w", err)
		}
	}

	return client, nil
}

func (s *SMTPEmailService) SendWelcome(ctx context.Context, to string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("수신자 이메일 주소가 비어 있습니다")
	}

	subject := "dBtree에 오신 것을 환영합니다"
	htmlBody := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>dBtree에 오신 것을 환영합니다</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4a86e8; color: white; padding: 10px; text-align: center; }
        .content { padding: 20px; }
        .footer { font-size: 12px; color: #666; text-align: center; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>dBtree에 오신 것을 환영합니다</h1>
        </div>
        <div class="content">
            <p>안녕하세요, dBtree입니다.</p>
            <p>회원가입을 축하합니다! 이제 dBtree의 모든 기능을 사용하실 수 있습니다.</p>
            <p>dBtree를 선택해 주셔서 감사합니다.</p>
        </div>
        <div class="footer">
            <p>본 이메일은 자동으로 발송되었습니다. 회신하지 마세요.</p>
            <p>&copy; 2025 dBtree. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	message := "From: " + s.config.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		htmlBody

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	client, err := createSMTPClient(addr, s.config.Host, s.config.Username, s.config.Password)
	if err != nil {
		return fmt.Errorf("SMTP 클라이언트 생성 실패: %w", err)
	}
	defer client.Close()

	if err := client.Mail(s.config.From); err != nil {
		return fmt.Errorf("발신자 설정 실패: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("수신자 설정 실패: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("데이터 쓰기 준비 실패: %w", err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("이메일 내용 쓰기 실패: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("데이터 쓰기 완료 실패: %w", err)
	}

	return nil
}

const emailTemplateOTP = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>dBtree 인증 코드</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4a86e8; color: white; padding: 10px; text-align: center; }
        .content { padding: 20px; }
        .code { font-size: 32px; font-weight: bold; text-align: center; padding: 20px; 
                background-color: #f0f0f0; margin: 20px 0; letter-spacing: 5px; }
        .footer { font-size: 12px; color: #666; text-align: center; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>dBtree 인증 코드</h1>
        </div>
        <div class="content">
            <p>안녕하세요, dBtree입니다.</p>
            <p>요청하신 인증 코드는 다음과 같습니다:</p>
            <div class="code">%s</div>
            <p>이 코드는 %d분 동안 유효합니다.</p>
            <p>본인이 요청하지 않았다면 이 이메일을 무시해주세요.</p>
        </div>
        <div class="footer">
            <p>본 이메일은 자동으로 발송되었습니다. 회신하지 마세요.</p>
			   <p>&copy; 2025 dBtree. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`
