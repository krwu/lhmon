package notifier

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// MailgunNotifier sends mail via Mailgun Messages API.
type MailgunNotifier struct {
	domain string
	apiKey string
	from   string
	to     []string
	base   string // e.g. https://api.mailgun.net or https://api.eu.mailgun.net
}

var _ Notifier = (*MailgunNotifier)(nil)

// MailgunSettings holds Mailgun delivery settings.
type MailgunSettings struct {
	Domain  string
	APIKey  string
	From    string
	To      []string
	BaseURL string
}

func NewMailgun(s MailgunSettings) *MailgunNotifier {
	base := strings.TrimRight(s.BaseURL, "/")
	if base == "" {
		base = "https://api.mailgun.net"
	}
	return &MailgunNotifier{
		domain: s.Domain,
		apiKey: s.APIKey,
		from:   s.From,
		to:     append([]string(nil), s.To...),
		base:   base,
	}
}

func (n *MailgunNotifier) Send(ctx context.Context, title, message string) error {
	if n == nil {
		return fmt.Errorf("mailgun: notifier is nil")
	}
	if n.domain == "" || n.apiKey == "" || n.from == "" || len(n.to) == 0 {
		return fmt.Errorf("mailgun: domain/api_key/from/to required")
	}

	form := url.Values{}
	form.Set("from", n.from)
	form.Set("subject", title)
	form.Set("text", message)
	for _, addr := range n.to {
		form.Add("to", addr)
	}

	api := fmt.Sprintf("%s/v3/%s/messages", n.base, n.domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, api, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth("api", n.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	return fmt.Errorf("mailgun: status %d: %s", resp.StatusCode, string(body))
}
