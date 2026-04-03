package notification

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	domain "github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
)

type SMTPStrategy struct{}

func NewSMTPStrategy() *SMTPStrategy {
	return &SMTPStrategy{}
}

func (s *SMTPStrategy) Provider() domainNotification.Provider {
	return domainNotification.ProviderSMTP
}

func (s *SMTPStrategy) ValidateSettings(settings *domainNotification.SMTPSettings) error {
	ve := domain.NewValidationError()

	if settings == nil {
		return ve.Add("settings", "is required")
	}
	if !settings.Provider.Valid() {
		ve.Add("provider", "must be smtp")
	}
	if strings.TrimSpace(settings.Host) == "" {
		ve.Add("host", "is required")
	}
	if settings.Port < 1 || settings.Port > 65535 {
		ve.Add("port", "must be between 1 and 65535")
	}
	if _, err := mail.ParseAddress(settings.FromEmail); err != nil {
		ve.Add("from_email", "must be a valid email")
	}
	if settings.ReplyTo != "" {
		if _, err := mail.ParseAddress(settings.ReplyTo); err != nil {
			ve.Add("reply_to", "must be a valid email")
		}
	}
	if !settings.Security.Valid() {
		ve.Add("security", "must be one of: none starttls tls")
	}
	if !settings.AuthMode.Valid() {
		ve.Add("auth_mode", "must be one of: none plain")
	}
	if settings.AuthMode == domainNotification.AuthModePlain && strings.TrimSpace(settings.Username) == "" {
		ve.Add("username", "is required when auth_mode is plain")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

func (s *SMTPStrategy) Send(ctx context.Context, settings *domainNotification.SMTPSettings, password string, message domainNotification.EmailMessage) error {
	if err := s.ValidateSettings(settings); err != nil {
		return err
	}
	if len(message.To) == 0 {
		return domain.NewValidationError().Add("to", "is required")
	}
	for _, recipient := range message.To {
		if _, err := mail.ParseAddress(recipient); err != nil {
			return domain.NewValidationError().Add("to", "must contain valid email addresses")
		}
	}
	if strings.TrimSpace(message.Subject) == "" {
		return domain.NewValidationError().Add("subject", "is required")
	}

	payload, err := buildSMTPMessage(settings, message)
	if err != nil {
		return err
	}

	addr := net.JoinHostPort(settings.Host, strconv.Itoa(settings.Port))

	switch settings.Security {
	case domainNotification.SecurityTLS:
		return s.sendImplicitTLS(ctx, settings, password, addr, payload, message.To)
	default:
		return s.sendPlainOrStartTLS(ctx, settings, password, addr, payload, message.To)
	}
}

func (s *SMTPStrategy) sendPlainOrStartTLS(ctx context.Context, settings *domainNotification.SMTPSettings, password, addr string, payload []byte, recipients []string) error {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, settings.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if settings.Security == domainNotification.SecuritySTARTTLS {
		if err := client.StartTLS(s.tlsConfig(settings)); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if err := s.applyAuth(client, settings, password); err != nil {
		return err
	}

	return s.sendData(client, settings.FromEmail, recipients, payload)
}

func (s *SMTPStrategy) sendImplicitTLS(ctx context.Context, settings *domainNotification.SMTPSettings, password, addr string, payload []byte, recipients []string) error {
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: 10 * time.Second},
		Config:    s.tlsConfig(settings),
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, settings.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if err := s.applyAuth(client, settings, password); err != nil {
		return err
	}

	return s.sendData(client, settings.FromEmail, recipients, payload)
}

func (s *SMTPStrategy) applyAuth(client *smtp.Client, settings *domainNotification.SMTPSettings, password string) error {
	switch settings.AuthMode {
	case domainNotification.AuthModeNone:
		return nil
	case domainNotification.AuthModePlain:
		auth := smtp.PlainAuth("", settings.Username, password, settings.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
		return nil
	default:
		return domain.NewValidationError().Add("auth_mode", "unsupported auth mode")
	}
}

func (s *SMTPStrategy) sendData(client *smtp.Client, from string, recipients []string, payload []byte) error {
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("smtp rcpt to: %w", err)
		}
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := writer.Write(payload); err != nil {
		_ = writer.Close()
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}
	return nil
}

func (s *SMTPStrategy) tlsConfig(settings *domainNotification.SMTPSettings) *tls.Config {
	return &tls.Config{
		ServerName:         settings.Host,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: settings.AllowInsecureTLS,
	}
}

func buildSMTPMessage(settings *domainNotification.SMTPSettings, message domainNotification.EmailMessage) ([]byte, error) {
	from := (&mail.Address{Name: settings.FromName, Address: settings.FromEmail}).String()
	to := make([]string, 0, len(message.To))
	for _, recipient := range message.To {
		parsed, err := mail.ParseAddress(recipient)
		if err != nil {
			return nil, err
		}
		to = append(to, parsed.String())
	}

	headers := []string{
		"From: " + from,
		"To: " + strings.Join(to, ", "),
		"Subject: " + message.Subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Date: " + time.Now().UTC().Format(time.RFC1123Z),
	}
	if settings.ReplyTo != "" {
		replyTo, err := mail.ParseAddress(settings.ReplyTo)
		if err != nil {
			return nil, err
		}
		headers = append(headers, "Reply-To: "+replyTo.String())
	}

	body := message.TextBody
	if strings.TrimSpace(body) == "" {
		body = "SMTP test message"
	}

	content := strings.Join(headers, "\r\n") + "\r\n\r\n" + body
	return []byte(content), nil
}
