package main

import (
	"os"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type NotifyType string

const (
	NotifySCT      NotifyType = "sct"
	NotifyWERobot  NotifyType = "werobot"
	NotifyNotifyx  NotifyType = "notifyx"
	NotifyTelegram NotifyType = "telegram"
)

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
	Accounts         []account  `yaml:"accounts"`
	SSL              SSLConfig  `yaml:"ssl"`
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
