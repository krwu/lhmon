package notifier

import (
	"context"
	"errors"
	"testing"
)

type stubNotifier struct {
	err error
	n   *int
}

func (s stubNotifier) Send(context.Context, string, string) error {
	if s.n != nil {
		*s.n++
	}
	return s.err
}

func TestMultiSendJoinsErrors(t *testing.T) {
	var n int
	m := NewMulti(
		stubNotifier{n: &n},
		stubNotifier{n: &n, err: errors.New("boom")},
		stubNotifier{n: &n},
	)
	err := m.Send(context.Background(), "t", "m")
	if err == nil {
		t.Fatal("want error")
	}
	if n != 3 {
		t.Fatalf("want 3 sends, got %d", n)
	}
	if m.Len() != 3 {
		t.Fatalf("len=%d", m.Len())
	}
}
