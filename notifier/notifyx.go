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
	notifyxURL = "https://www.notifyx.cn/api/v1/send/%s"
)

type NotifyxNotifier struct {
	key  string
	team string
}

var _ Notifier = (*NotifyxNotifier)(nil)

func NewNotifyx(key, team string) *NotifyxNotifier {
	return &NotifyxNotifier{key: key, team: team}
}

func (n *NotifyxNotifier) Send(ctx context.Context, title, message string) error {
	payload := map[string]string{
		"title":       title,
		"content":     message,
		"description": title,
	}
	if n.team != "" {
		payload["team"] = n.team
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[%s] ERROR %v\n", time.Now().Format("20060102 15:04:05"), err)
		return err
	}

	api := fmt.Sprintf(notifyxURL, n.key)
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
