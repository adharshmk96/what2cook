package mail

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"what2cook-api/internal/config"
)

// Mailer sends transactional email, or logs when SMTP is not configured.
type Mailer struct {
	smtp      config.SMTPConfig
	publicURL string
}

// New creates a Mailer. Empty smtp.Host enables log-fallback for reset links.
func New(smtpCfg config.SMTPConfig, publicURL string) *Mailer {
	return &Mailer{smtp: smtpCfg, publicURL: strings.TrimRight(publicURL, "/")}
}

// SendPasswordReset delivers a reset link via SMTP, or logs it when SMTP host is empty.
func (m *Mailer) SendPasswordReset(toEmail, rawToken string) error {
	link := fmt.Sprintf("%s/app/reset-password?token=%s", m.publicURL, rawToken)

	if strings.TrimSpace(m.smtp.Host) == "" {
		log.Printf("Reset link: %s", link)
		return nil
	}

	from := m.smtp.From
	if from == "" {
		from = m.smtp.User
	}
	if from == "" {
		return fmt.Errorf("smtp.from (or smtp.user) is required when smtp.host is set")
	}

	subject := "Reset your what2cook password"
	body := fmt.Sprintf("Reset your password using this link (expires soon):\r\n\r\n%s\r\n", link)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, toEmail, subject, body,
	))

	addr := fmt.Sprintf("%s:%d", m.smtp.Host, m.smtp.Port)
	var auth smtp.Auth
	if m.smtp.User != "" {
		auth = smtp.PlainAuth("", m.smtp.User, m.smtp.Password, m.smtp.Host)
	}

	if err := smtp.SendMail(addr, auth, from, []string{toEmail}, msg); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	log.Printf("password reset email sent to %s", toEmail)
	return nil
}
