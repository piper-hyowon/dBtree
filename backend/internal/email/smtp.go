package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"embed"
	"encoding/base64"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core/email"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"math/rand"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"github.com/piper-hyowon/dBtree/internal/core/auth"
	"github.com/piper-hyowon/dBtree/internal/platform/config"
)

//go:embed images/*.png
var emailFS embed.FS

// 이미지 CID (Content-ID)
const (
	otpHeaderImageCID = "otp_header_image@dbtree.com"
	welcomeImageCID   = "welcome_image@dbtree.com"
	goodbyeImageCID   = "goodbye_image@dbtree.com"
)

// 이미지 파일 이름 매핑
var imageFilenames = map[string]string{
	otpHeaderImageCID: "images/otp_header.png",
	welcomeImageCID:   "images/welcome.png",
	goodbyeImageCID:   "images/goodbye.png",
}

type service struct {
	config   config.SMTPConfig
	client   *smtp.Client
	clientMu sync.Mutex
	lastUsed time.Time
}

const clientExpirationTime = 30 * time.Minute

var _ email.Service = (*service)(nil)

func NewService(config config.SMTPConfig) (email.Service, error) {
	return &service{
		config: config,
	}, nil
}

func GetImage(id string) ([]byte, error) {
	filename, ok := imageFilenames[id]
	if !ok {
		return nil, errors.Wrap(fmt.Errorf("알 수 없는 이미지 ID: %s", id))
	}

	data, err := emailFS.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "임베딩된 이미지 읽기 실패 %s", filename)
	}

	return data, nil
}

func (s *service) getClient() (*smtp.Client, error) {
	s.clientMu.Lock()
	defer s.clientMu.Unlock()

	now := time.Now()

	if s.client != nil {
		if now.Sub(s.lastUsed) < clientExpirationTime {
			if err := s.client.Noop(); err == nil {
				s.lastUsed = now
				return s.client, nil
			}
		}
		s.client.Close()
		s.client = nil
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	client, err := createSMTPClient(addr, s.config.Host, s.config.Username, s.config.Password)
	if err != nil {
		return nil, err
	}

	s.client = client
	s.lastUsed = now
	return client, nil
}

func createSMTPClient(addr, host, username, password string) (*smtp.Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	client, err := smtp.Dial(addr)
	if err != nil {
		return nil, errors.Wrapf(err, "SMTP 서버 연결 실패")
	}

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err = client.StartTLS(tlsConfig); err != nil {
			client.Close()
			return nil, errors.Wrapf(err, "STARTTLS 시작 실패")
		}
	}

	if ok, _ := client.Extension("AUTH"); ok && username != "" && password != "" {
		authApi := smtp.PlainAuth("", username, password, host)
		if err = client.Auth(authApi); err != nil {
			client.Close()
			return nil, errors.Wrapf(err, "SMTP 인증 실패")
		}
	}

	return client, nil
}

func (s *service) SendOTP(ctx context.Context, to string, otp string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if otp == "" {
		return errors.Wrap(errors.New("otp is empty"))

	}

	subject := "dBtree 인증 코드"

	htmlTemplate := `
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="UTF-8">
        <title>dBtree 인증 코드</title>
    </head>
    <body style="font-family: Arial, sans-serif; line-height: 1.6; margin: 0; padding: 0;">
        <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
            <div style="background-color: #4a86e8; color: white; padding: 10px; text-align: center;">
                <h1 style="margin: 0; padding: 0;">dBtree 인증 코드</h1>
            </div>
            <div style="padding: 20px;">
                <p>안녕하세요, dBtree입니다.</p>
                <img src="cid:%s" alt="dBtree" style="max-width: 100%%; height: auto; display: block; margin: 20px auto;">
                <p>요청하신 인증 코드는 다음과 같습니다:</p>
                <div style="font-size: 32px; font-weight: bold; text-align: center; padding: 20px; 
                        background-color: #f0f0f0; margin: 20px 0; letter-spacing: 5px;">%s</div>
                <p>이 코드는 %d분 동안 유효합니다.</p>
                <p>본인이 요청하지 않았다면 이 이메일을 무시해주세요.</p>
            </div>
            <div style="font-size: 12px; color: #666; text-align: center; margin-top: 20px;">
                <p>본 이메일은 자동으로 발송되었습니다. 회신하지 마세요.</p>
                <p>&copy; 2025 dBtree. All rights reserved.</p>
            </div>
        </div>
    </body>
    </html>
    `

	htmlBody := fmt.Sprintf(
		htmlTemplate,
		otpHeaderImageCID,
		otp,
		auth.OTPExpirationMinutes)

	images := make(map[string][]byte)
	imgData, err := GetImage(otpHeaderImageCID)
	if err == nil {
		images[otpHeaderImageCID] = imgData
	} else {
		fmt.Printf("경고: OTP 헤더 이미지 로드 실패: %v\n", err)
	}

	return s.SendWithImages(ctx, to, subject, htmlBody, images)
}

func (s *service) SendWelcome(ctx context.Context, to string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	subject := "dBtree에 오신 것을 환영합니다"

	htmlTemplate := `
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="UTF-8">
        <title>dBtree에 오신 것을 환영합니다</title>
    </head>
    <body style="font-family: Arial, sans-serif; line-height: 1.6; margin: 0; padding: 0;">
        <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
            <div style="background-color: #4a86e8; color: white; padding: 10px; text-align: center;">
                <h1 style="margin: 0; padding: 0;">dBtree에 오신 것을 환영합니다</h1>
            </div>
            <div style="padding: 20px; text-align: center;">
                <p>안녕하세요, dBtree입니다.</p>
                <img src="cid:%s" alt="Welcome" style="max-width: 100%%; height: auto; display: block; margin: 20px auto;">
                <p>회원가입을 축하합니다! 이제 dBtree의 모든 기능을 사용하실 수 있습니다.</p>
                <p>dBtree를 선택해 주셔서 감사합니다.</p>
            </div>
            <div style="font-size: 12px; color: #666; text-align: center; margin-top: 20px;">
                <p>본 이메일은 자동으로 발송되었습니다. 회신하지 마세요.</p>
                <p>&copy; 2025 dBtree. All rights reserved.</p>
            </div>
        </div>
    </body>
    </html>
    `

	htmlBody := fmt.Sprintf(htmlTemplate, welcomeImageCID)
	images := make(map[string][]byte)
	imgData, err := GetImage(welcomeImageCID)
	if err == nil {
		images[welcomeImageCID] = imgData
	} else {
		fmt.Printf("경고: 환영 이미지 로드 실패: %v\n", err)
	}

	return s.SendWithImages(ctx, to, subject, htmlBody, images)
}

func (s *service) SendGoodbye(ctx context.Context, to string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	subject := "dBtree 회원 탈퇴 완료"

	htmlTemplate := `
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="UTF-8">
        <title>dBtree 회원 탈퇴 완료</title>
    </head>
    <body style="font-family: Arial, sans-serif; line-height: 1.6; margin: 0; padding: 0;">
        <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
            <div style="background-color: #4a86e8; color: white; padding: 10px; text-align: center;">
                <h1 style="margin: 0; padding: 0;">dBtree 회원 탈퇴 완료</h1>
            </div>
            <div style="padding: 20px; text-align: center;">
                <p>안녕하세요, dBtree입니다.</p>
                <p>회원님의 계정이 성공적으로 탈퇴 처리되었습니다.</p>
                <img src="cid:%s" alt="Goodbye" style="max-width: 100%%; height: auto; margin: 20px auto; display: block;">
                <p>그동안 dBtree 서비스를 이용해 주셔서 감사합니다.</p>
                <p>언제든지 다시 돌아오실 수 있습니다.</p>
            </div>
            <div style="font-size: 12px; color: #666; text-align: center; margin-top: 20px;">
                <p>본 이메일은 자동으로 발송되었습니다. 회신하지 마세요.</p>
                <p>&copy; 2025 dBtree. All rights reserved.</p>
            </div>
        </div>
    </body>
    </html>
    `

	htmlBody := fmt.Sprintf(htmlTemplate, goodbyeImageCID)
	images := make(map[string][]byte)
	imgData, err := GetImage(goodbyeImageCID)
	if err == nil {
		images[goodbyeImageCID] = imgData
	} else {
		fmt.Printf("경고: 탈퇴 이미지 로드 실패: %v\n", err)
	}

	return s.SendWithImages(ctx, to, subject, htmlBody, images)
}

func (s *service) Send(ctx context.Context, to string, subject string, htmlBody string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	client, err := s.getClient()
	if err != nil {
		return errors.Wrapf(err, "SMTP 클라이언트 가져오기 실패")
	}

	if err := client.Reset(); err != nil {
		s.clientMu.Lock()
		s.client.Close()
		s.client = nil
		s.clientMu.Unlock()

		client, err = s.getClient()
		if err != nil {
			return errors.Wrapf(err, "SMTP 클라이언트 재연결 실패")

		}
	}

	message := "From: " + s.config.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		htmlBody

	if err := s.sendEmailContent(client, to, []byte(message)); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// SendWithImages 이미지 포함해서 발송 (MIME 인코딩)
func (s *service) SendWithImages(ctx context.Context, to string, subject string, htmlBody string, images map[string][]byte) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	validImages := make(map[string][]byte)
	for imgID, imgData := range images {
		if imgData == nil || len(imgData) == 0 {
			fmt.Printf("경고: 이미지 데이터가 없습니다: %s\n", imgID)
			continue
		}
		validImages[imgID] = imgData
	}

	if len(validImages) == 0 {
		return s.Send(ctx, to, subject, htmlBody)
	}

	client, err := s.getClient()
	if err != nil {
		return errors.Wrapf(err, "SMTP 클라이언트 가져오기 실패")
	}

	if err := client.Reset(); err != nil {
		s.clientMu.Lock()
		s.client.Close()
		s.client = nil
		s.clientMu.Unlock()

		client, err = s.getClient()
		if err != nil {
			return errors.Wrapf(err, "SMTP 클라이언트 재연결 실패")
		}
	}

	mixedBoundary := generateBoundary()
	relatedBoundary := generateBoundary()
	altBoundary := generateBoundary()

	var buf bytes.Buffer

	// 헤더
	buf.WriteString(fmt.Sprintf("From: %s\r\n", s.config.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", mixedBoundary))

	// mixed 파트 시작
	buf.WriteString(fmt.Sprintf("--%s\r\n", mixedBoundary))
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/related; boundary=\"%s\"\r\n\r\n", relatedBoundary))

	// related 파트 시작
	buf.WriteString(fmt.Sprintf("--%s\r\n", relatedBoundary))
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", altBoundary))

	// alternative 파트 (텍스트)
	buf.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
	buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	buf.WriteString("이 이메일은 HTML 형식으로 작성되었습니다. HTML을 지원하는 이메일 클라이언트에서 확인해 주세요.\r\n\r\n")

	// alternative 파트 (HTML)
	buf.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
	buf.WriteString(htmlBody)
	buf.WriteString("\r\n\r\n")

	// alternative 종료
	buf.WriteString(fmt.Sprintf("--%s--\r\n\r\n", altBoundary))

	// 이미지
	for cid, data := range validImages {
		buf.WriteString(fmt.Sprintf("--%s\r\n", relatedBoundary))
		buf.WriteString("Content-Type: image/png\r\n")
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString(fmt.Sprintf("Content-ID: <%s>\r\n", cid))
		buf.WriteString("Content-Disposition: inline\r\n\r\n")

		// Base64 인코딩
		encodedData := base64.StdEncoding.EncodeToString(data)
		for i := 0; i < len(encodedData); i += 76 {
			end := i + 76
			if end > len(encodedData) {
				end = len(encodedData)
			}
			buf.WriteString(encodedData[i:end])
			buf.WriteString("\r\n")
		}
		buf.WriteString("\r\n")
	}

	buf.WriteString(fmt.Sprintf("--%s--\r\n\r\n", relatedBoundary))
	buf.WriteString(fmt.Sprintf("--%s--\r\n", mixedBoundary))

	if err := s.sendEmailContent(client, to, buf.Bytes()); err != nil {
		return errors.Wrap(err)

	}

	return nil
}

func (s *service) sendEmailContent(client *smtp.Client, to string, messageData []byte) error {
	if err := client.Mail(s.config.From); err != nil {
		s.clientMu.Lock()
		if s.client == client {
			s.client.Close()
			s.client = nil
		}
		s.clientMu.Unlock()

		return errors.Wrapf(err, "발신자 설정 실패")
	}

	if err := client.Rcpt(to); err != nil {
		if strings.Contains(err.Error(), "Invalid recipient") ||
			strings.Contains(err.Error(), "not verified") ||
			strings.Contains(err.Error(), "Message rejected") {
			return errors.NewInvalidEmailError(err.Error())

		}
		return errors.Wrapf(err, "수신자 설정 실패")
	}

	wc, err := client.Data()
	if err != nil {
		return errors.Wrapf(err, "data 준비 실패")
	}

	_, err = wc.Write(messageData)
	if err != nil {
		return errors.Wrapf(err, "이메일 내용 Write 실패")
	}

	if err := wc.Close(); err != nil {
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "Email address is not verified") ||
			strings.Contains(errorMsg, "Message rejected") {
			return errors.NewInvalidEmailError(string(errorMsg))
		}

		return errors.Wrapf(err, "데이터 쓰기 Close 실패")
	}

	s.clientMu.Lock()
	s.lastUsed = time.Now()
	s.clientMu.Unlock()

	return nil
}

func (s *service) Close() {
	s.clientMu.Lock()
	defer s.clientMu.Unlock()

	if s.client != nil {
		s.client.Quit()
		s.client.Close()
		s.client = nil
	}
}

func generateBoundary() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := "----=_NextPart_"

	for i := 0; i < 16; i++ {
		result += string(chars[r.Intn(len(chars))])
	}

	return result
}
