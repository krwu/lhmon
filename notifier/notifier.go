package notifier

import "context"

type Notifier interface {
	Send(ctx context.Context, title, message string) error
}
