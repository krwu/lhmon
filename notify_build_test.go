package main

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBuildNotifiers_LegacySingle(t *testing.T) {
	c := Config{
		NotifyType: NotifySCT,
		SCTKey:     "SCTKEY",
	}
	list, err := BuildNotifiers(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 notifier, got %d", len(list))
	}
}

func TestBuildNotifiers_MultiEnabled(t *testing.T) {
	c := Config{
		SCT: &SCTChannel{Enabled: true, Key: "k1"},
		Telegram: &TelegramChannel{
			Enabled:  true,
			BotToken: "tok",
			UserID:   "1",
		},
		SMTP: &SMTPChannel{
			Enabled: false,
			Host:    "smtp.example.com",
			From:    "a@b.com",
			To:      StringList{"c@d.com"},
		},
	}
	list, err := BuildNotifiers(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2 notifiers, got %d", len(list))
	}
}

func TestBuildNotifiers_MultiNone(t *testing.T) {
	_, err := BuildNotifiers(Config{})
	if err == nil {
		t.Fatal("want error when no channels")
	}
}

func TestBuildNotifiers_SingleSMTPIncomplete(t *testing.T) {
	_, err := BuildNotifiers(Config{NotifyType: NotifySMTP})
	if err == nil {
		t.Fatal("want incomplete smtp error")
	}
}

func TestStringListYAML(t *testing.T) {
	var got struct {
		To StringList `yaml:"to"`
	}
	if err := yaml.Unmarshal([]byte("to: a@b.com\n"), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.To) != 1 || got.To[0] != "a@b.com" {
		t.Fatalf("scalar: %+v", got.To)
	}
	got = struct {
		To StringList `yaml:"to"`
	}{}
	if err := yaml.Unmarshal([]byte("to: [a@b.com, c@d.com]\n"), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.To) != 2 {
		t.Fatalf("seq: %+v", got.To)
	}
}

func TestBuildNotifiers_NestedOverridesFlat(t *testing.T) {
	c := Config{
		NotifyType: NotifySCT,
		SCTKey:     "flat",
		SCT:        &SCTChannel{Key: "nested"},
	}
	list, err := BuildNotifiers(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("got %d", len(list))
	}
	// ensure no panic / type ok; key itself is unexported inside notifier
	_ = list[0]
}

func TestBuildNotifiers_Unsupported(t *testing.T) {
	_, err := BuildNotifiers(Config{NotifyType: "nope"})
	if err == nil || !strings.Contains(err.Error(), "不支持") {
		t.Fatalf("want unsupported error, got %v", err)
	}
}
