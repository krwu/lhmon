package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"lighthouse-monitor/log"
)

const (
	telegramURL = "https://api.telegram.org/bot%s/sendMessage"
)

type TelegramNotifier struct {
	botToken string
	userID   string
}

var _ Notifier = (*TelegramNotifier)(nil)

func NewTelegram(botToken, userID string) *TelegramNotifier {
	return &TelegramNotifier{botToken: botToken, userID: userID}
}

func (n *TelegramNotifier) Send(ctx context.Context, title, message string) error {
	payload := map[string]string{
		"chat_id": n.userID,
		"text":    fmt.Sprintf("【%s】\n%s", title, message),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[%s] ERROR %v\n", time.Now().Format("20060102 15:04:05"), err)
		return err
	}

	api := fmt.Sprintf(telegramURL, n.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, api, bytes.NewReader(data))
	if err != nil {
		log.Printf("[%s] ERROR %v\n", time.Now().Format("20060102 15:04:05"), err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	_, err = client.Do(req)
	if err != nil {
		log.Printf("[%s] ERROR %v\n", time.Now().Format("20060102 15:04:05"), err)
	}
	return err
}
