package smtp

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
)

type SMTPClient struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewSMTPClient(host string, port int, username, password, from string) *SMTPClient {
	return &SMTPClient{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (c *SMTPClient) SendEmail(to []string, subject, htmlContent string, textContent string, variables map[string]interface{}) error {
	// HTML 템플릿 처리
	htmlTemplate, err := template.New("email").Parse(htmlContent)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %v", err)
	}

	var htmlBuffer bytes.Buffer
	if err := htmlTemplate.Execute(&htmlBuffer, variables); err != nil {
		return fmt.Errorf("failed to execute HTML template: %v", err)
	}

	// Text 템플릿 처리
	textTemplate, err := template.New("email").Parse(textContent)
	if err != nil {
		return fmt.Errorf("failed to parse text template: %v", err)
	}

	var textBuffer bytes.Buffer
	if err := textTemplate.Execute(&textBuffer, variables); err != nil {
		return fmt.Errorf("failed to execute text template: %v", err)
	}

	// 이메일 헤더 구성
	mime := "MIME-version: 1.0;\nContent-Type: multipart/alternative; boundary=\"boundary-string\"\n\n"

	body := fmt.Sprintf("--%s\n"+
		"Content-Type: text/plain; charset=\"UTF-8\"\n"+
		"\n%s\n\n"+
		"--%s\n"+
		"Content-Type: text/html; charset=\"UTF-8\"\n"+
		"\n%s\n\n"+
		"--%s--",
		"boundary-string",
		textBuffer.String(),
		"boundary-string",
		htmlBuffer.String(),
		"boundary-string")

	msg := fmt.Sprintf("Subject: %s\nTo: %s\nFrom: %s\n%s\n%s",
		subject,
		strings.Join(to, ","),
		c.from,
		mime,
		body)

	// SMTP 인증
	auth := smtp.PlainAuth("", c.username, c.password, c.host)

	// 이메일 전송
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	if err := smtp.SendMail(addr, auth, c.from, to, []byte(msg)); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
