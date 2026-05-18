package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

// SMTPConfig holds connection details for an SMTP server.
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// SMTPMailer sends email via SMTP. Port 465 uses implicit TLS; all other ports
// use STARTTLS via smtp.SendMail.
type SMTPMailer struct {
	cfg SMTPConfig
}

func NewSMTPMailer(cfg SMTPConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) Send(to, subject, htmlBody string) error {
	addr := net.JoinHostPort(m.cfg.Host, m.cfg.Port)
	msg := buildMIMEMessage(m.cfg.From, to, subject, htmlBody)

	if m.cfg.Port == "465" {
		return m.sendImplicitTLS(addr, to, msg)
	}

	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	if err := smtp.SendMail(addr, auth, m.cfg.From, []string{to}, msg); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}

	return nil
}

func (m *SMTPMailer) sendImplicitTLS(addr, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: m.cfg.Host})
	if err != nil {
		return fmt.Errorf("smtp tls dial: %w", err)
	}

	client, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err = client.Mail(m.cfg.From); err != nil {
		return fmt.Errorf("smtp MAIL FROM: %w", err)
	}

	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp RCPT TO: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA: %w", err)
	}

	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("smtp write body: %w", err)
	}

	return w.Close()
}

func buildMIMEMessage(from, to, subject, htmlBody string) []byte {
	header := strings.Join([]string{
		"MIME-Version: 1.0",
		`Content-Type: text/html; charset="UTF-8"`,
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
	}, "\r\n")

	return fmt.Appendf(nil, "%s\r\n\r\n%s", header, htmlBody)
}
