package notifier

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"lighthouse-monitor/log"
)

const (
	scfURL = "https://sctapi.ftqq.com/"
)

type ScfNotifier struct {
	token string
}

var _ Notifier = (*ScfNotifier)(nil)

func NewSCT(token string) *ScfNotifier {
	return &ScfNotifier{token: token}
}

func (n *ScfNotifier) Send(ctx context.Context, title, message string) error {
	client := http.Client{}
	body := url.Values{}
	body.Add("title", title)
	body.Add("desp", message)
	api := scfURL + n.token + ".Send"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, api, strings.NewReader(body.Encode()))
	if err != nil {
		log.Printf("[%s] ERROR %v\n", time.Now().Format("20060102 15:04:05"), err)
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err = client.Do(req)
	if err != nil {
		log.Printf("[%s] ERROR %v\n", time.Now().Format("20060102 15:04:05"), err)
		return err
	}

	return nil
}
