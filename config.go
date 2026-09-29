package main

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type NotifyType string

const (
	NotifySCT      NotifyType = "sct"
	NotifyWERobot  NotifyType = "werobot"
	NotifyNotifyx  NotifyType = "notifyx"
	NotifyTelegram NotifyType = "telegram"
	NotifySMTP     NotifyType = "smtp"
	NotifySendGrid NotifyType = "sendgrid"
	NotifyMailgun  NotifyType = "mailgun"
)

// StringList unmarshals a YAML string or sequence of strings.
type StringList []string

func (s *StringList) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		v := strings.TrimSpace(value.Value)
		if v == "" {
			*s = nil
			return nil
		}
		*s = []string{v}
		return nil
	case yaml.SequenceNode:
		var list []string
		if err := value.Decode(&list); err != nil {
			return err
		}
		*s = list
		return nil
	case yaml.AliasNode:
		if value.Alias != nil {
			return s.UnmarshalYAML(value.Alias)
		}
	}
	return nil
}

type SCTChannel struct {
	Enabled bool   `yaml:"enabled"`
	Key     string `yaml:"key"`
}

type WERobotChannel struct {
	Enabled bool   `yaml:"enabled"`
	Webhook string `yaml:"webhook"`
	ChatID  string `yaml:"chatid"`
}

type NotifyxChannel struct {
	Enabled bool   `yaml:"enabled"`
	Key     string `yaml:"key"`
	Team    string `yaml:"team"`
}

type TelegramChannel struct {
	Enabled  bool   `yaml:"enabled"`
	BotToken string `yaml:"bot_token"`
	UserID   string `yaml:"user_id"`
}

type SMTPChannel struct {
	Enabled  bool       `yaml:"enabled"`
	Host     string     `yaml:"host"`
	Port     int        `yaml:"port"`
	Username string     `yaml:"username"`
	Password string     `yaml:"password"`
	From     string     `yaml:"from"`
	To       StringList `yaml:"to"`
}

type SendGridChannel struct {
	Enabled bool       `yaml:"enabled"`
	APIKey  string     `yaml:"api_key"`
	From    string     `yaml:"from"`
	To      StringList `yaml:"to"`
}

type MailgunChannel struct {
	Enabled bool       `yaml:"enabled"`
	Domain  string     `yaml:"domain"`
	APIKey  string     `yaml:"api_key"`
	From    string     `yaml:"from"`
	To      StringList `yaml:"to"`
	BaseURL string     `yaml:"base_url"` // optional; default https://api.mailgun.net (EU: https://api.eu.mailgun.net)
}

type Config struct {
	WarnRate         float64    `yaml:"warn_rate"`
	ShutdownRate     float64    `yaml:"shutdown_rate"`
	CheckInterval    int64      `yaml:"check_interval"`
	NotifyType       NotifyType `yaml:"notify_method"`
	SCTKey           string     `yaml:"sct_key"`
	WERobotWebhook   string     `yaml:"werobot_webhook"`
	WERobotChatID    string     `yaml:"werobot_chatid"`
	NotifyxKey       string     `yaml:"notifyx_key"`
	NotifyxTeam      string     `yaml:"notifyx_team"`
	TelegramBotToken string     `yaml:"telegram_bot_token"`
	TelegramUserID   string     `yaml:"telegram_user_id"`

	SCT      *SCTChannel      `yaml:"sct"`
	WERobot  *WERobotChannel  `yaml:"werobot"`
	Notifyx  *NotifyxChannel  `yaml:"notifyx"`
	Telegram *TelegramChannel `yaml:"telegram"`
	SMTP     *SMTPChannel     `yaml:"smtp"`
	SendGrid *SendGridChannel `yaml:"sendgrid"`
	Mailgun  *MailgunChannel  `yaml:"mailgun"`

	Accounts []account `yaml:"accounts"`
	SSL      SSLConfig `yaml:"ssl"`
}

// SSLConfig controls SSL certificate monitoring and cleanup.
// expire_days: optional. When set, certificates with remaining life <= this many days
// are treated as expiring. When omitted, official FilterExpiring (~30 days) is used.
type SSLConfig struct {
	Enabled         bool  `yaml:"enabled"`
	AutoDelete      *bool `yaml:"auto_delete"`
	DryRun          *bool `yaml:"dry_run"`
	ExpireDays      *int  `yaml:"expire_days"`
	CheckInterval   int64 `yaml:"check_interval"`
	BindTaskTimeout int64 `yaml:"bind_task_timeout"`
	BindUseCache    *bool `yaml:"bind_use_cache"`
}

func (c SSLConfig) autoDelete() bool {
	if c.AutoDelete == nil {
		return true
	}
	return *c.AutoDelete
}

func (c SSLConfig) dryRun() bool {
	if c.DryRun == nil {
		return true
	}
	return *c.DryRun
}

func (c SSLConfig) bindUseCache() bool {
	if c.BindUseCache == nil {
		return true
	}
	return *c.BindUseCache
}

func (c SSLConfig) checkInterval() int64 {
	if c.CheckInterval <= 0 {
		return 86400
	}
	return c.CheckInterval
}

func (c SSLConfig) bindTaskTimeout() int64 {
	if c.BindTaskTimeout <= 0 {
		return 120
	}
	return c.BindTaskTimeout
}

type account struct {
	Name      string   `yaml:"name"`
	SecretID  string   `yaml:"secret_id"`
	SecretKey string   `yaml:"secret_key"`
	Regions   []string `yaml:"regions"`
}

func InitConfig(file string) {
	var conf Config
	yamlByte, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(yamlByte, &conf)
	if err != nil {
		logger.Fatal("failed to load config",
			zap.String("error", err.Error()),
			zap.String("file", file),
		)
	}
	Conf = conf
}
