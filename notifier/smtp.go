package notifier

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// SMTPNotifier delivers mail via SMTP (STARTTLS on submission ports by default).
type SMTPNotifier struct {
	host     string
	port     int
	username string
	password string
	from     string
	to       []string
}

var _ Notifier = (*SMTPNotifier)(nil)

// SMTPSettings holds SMTP delivery settings.
type SMTPSettings struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string
}

func NewSMTP(s SMTPSettings) *SMTPNotifier {
	port := s.Port
	if port <= 0 {
		port = 587
	}
	return &SMTPNotifier{
		host:     s.Host,
		port:     port,
		username: s.Username,
		password: s.Password,
		from:     s.From,
		to:       append([]string(nil), s.To...),
	}
}

func (n *SMTPNotifier) Send(ctx context.Context, title, message string) error {
	if n == nil {
		return fmt.Errorf("smtp: notifier is nil")
	}
	if n.host == "" || n.from == "" || len(n.to) == 0 {
		return fmt.Errorf("smtp: host/from/to required")
	}

	addr := fmt.Sprintf("%s:%d", n.host, n.port)
	body := buildMailMessage(n.from, n.to, title, message)

	dialer := &net.Dialer{}
	if deadline, ok := ctx.Deadline(); ok {
		dialer.Deadline = deadline
	} else {
		dialer.Timeout = 10 * time.Second
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, n.host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{ServerName: n.host, MinVersion: tls.VersionTLS12}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if n.username != "" {
		auth := smtp.PlainAuth("", n.username, n.password, n.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(n.from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	for _, rcpt := range n.to {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp rcpt %s: %w", rcpt, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write([]byte(body)); err != nil {
		_ = w.Close()
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}
	return client.Quit()
}

func buildMailMessage(from string, to []string, subject, body string) string {
	var b strings.Builder
	b.WriteString("From: ")
	b.WriteString(from)
	b.WriteString("\r\nTo: ")
	b.WriteString(strings.Join(to, ", "))
	b.WriteString("\r\nSubject: ")
	b.WriteString(encodeSubject(subject))
	b.WriteString("\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n")
	b.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		b.WriteString("\r\n")
	}
	return b.String()
}

func encodeSubject(subject string) string {
	// Keep ASCII subjects plain; wrap non-ASCII with RFC 2047 UTF-8 encoding.
	for i := 0; i < len(subject); i++ {
		if subject[i] >= 0x80 {
			return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(subject)) + "?="
		}
	}
	return subject
}
