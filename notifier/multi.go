package notifier

import (
	"context"
	"errors"
)

// Multi sends the same notification to every configured Notifier.
// Individual failures are collected with errors.Join; a nil Multi or empty list is a no-op.
type Multi struct {
	clients []Notifier
}

var _ Notifier = (*Multi)(nil)

func NewMulti(clients ...Notifier) *Multi {
	out := make([]Notifier, 0, len(clients))
	for _, c := range clients {
		if c != nil {
			out = append(out, c)
		}
	}
	return &Multi{clients: out}
}

func (m *Multi) Send(ctx context.Context, title, message string) error {
	if m == nil || len(m.clients) == 0 {
		return nil
	}
	var errs []error
	for _, c := range m.clients {
		if err := c.Send(ctx, title, message); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (m *Multi) Len() int {
	if m == nil {
		return 0
	}
	return len(m.clients)
}
