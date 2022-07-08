package notifier

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"lighthouse-monitor/log"
)

type NextrtNotifier struct {
	typ   string
	token string
}

const (
	nextrtURL = "https://api.nextrt.com/api/push/send"
)

var _ Notifier = (*NextrtNotifier)(nil)

func NewNextrt(typ, token string) *NextrtNotifier {
	return &NextrtNotifier{typ: typ, token: token}
}

func (n *NextrtNotifier) Send(ctx context.Context, title, message string) error {
	uri := fmt.Sprintf(
		"%s?title=%s&content=%s&token=%s&type=%s",
		nextrtURL, title, message, n.token, n.typ,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		log.Printf("[%s] ERROR %v\n", time.Now().Format("20060102 15:04:05"), err)
		return err
	}
	client := http.DefaultClient
	_, err = client.Do(req)
	if err != nil {
		log.Printf("[%s] ERROR %v\n", time.Now().Format("20060102 15:04:05"), err)
	}
	return err
}
