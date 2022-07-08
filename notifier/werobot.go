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

type WERobotNotifier struct {
	url    string
	chatid string
}

func NewWERobot(url, chatid string) *WERobotNotifier {
	return &WERobotNotifier{url: url, chatid: chatid}
}

var _ Notifier = (*WERobotNotifier)(nil)

func (n *WERobotNotifier) Send(ctx context.Context, title, message string) error {
	client := http.Client{}
	body := make(map[string]interface{})
	body["msgtype"] = "text"
	if n.chatid != "" {
		body["chatid"] = n.chatid
	}
	body["text"] = map[string]string{
		"content": fmt.Sprintf("【%s】%s", title, message),
	}
	data, err := json.Marshal(body)
	if err != nil {
		log.Printf("[%s] ERROR %v\n", time.Now().Format("20060102 15:04:05"), err)
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.url, bytes.NewReader(data))
	if err != nil {
		log.Printf("[%s] ERROR %v\n", time.Now().Format("20060102 15:04:05"), err)
		return err
	}
	if req.Header == nil {
		req.Header = http.Header{}
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = client.Do(req)
	if err != nil {
		log.Printf("[%s] ERROR %v\n", time.Now().Format("20060102 15:04:05"), err)
	}
	return err
}
