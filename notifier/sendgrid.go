package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const sendgridURL = "https://api.sendgrid.com/v3/mail/send"

// SendGridNotifier sends mail via SendGrid Web API v3.
type SendGridNotifier struct {
	apiKey string
	from   string
	to     []string
}

var _ Notifier = (*SendGridNotifier)(nil)

func NewSendGrid(apiKey, from string, to []string) *SendGridNotifier {
	return &SendGridNotifier{
		apiKey: apiKey,
		from:   from,
		to:     append([]string(nil), to...),
	}
}

func (n *SendGridNotifier) Send(ctx context.Context, title, message string) error {
	if n == nil {
		return fmt.Errorf("sendgrid: notifier is nil")
	}
	if n.apiKey == "" || n.from == "" || len(n.to) == 0 {
		return fmt.Errorf("sendgrid: api_key/from/to required")
	}

	type emailAddr struct {
		Email string `json:"email"`
	}
	to := make([]emailAddr, 0, len(n.to))
	for _, addr := range n.to {
		to = append(to, emailAddr{Email: addr})
	}
	payload := map[string]any{
		"personalizations": []map[string]any{
			{"to": to},
		},
		"from":    emailAddr{Email: n.from},
		"subject": title,
		"content": []map[string]string{
			{"type": "text/plain", "value": message},
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sendgridURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+n.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	return fmt.Errorf("sendgrid: status %d: %s", resp.StatusCode, string(body))
}
