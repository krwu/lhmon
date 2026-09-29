package main

import (
	"fmt"
	"strings"

	"lighthouse-monitor/notifier"
)

var allNotifyTypes = []NotifyType{
	NotifySCT,
	NotifyWERobot,
	NotifyNotifyx,
	NotifyTelegram,
	NotifySMTP,
	NotifySendGrid,
	NotifyMailgun,
}

// BuildNotifiers resolves notifiers from config.
// If notify_method is set, only that single channel is used (legacy).
// If notify_method is empty, every channel with enabled:true and complete params is used.
func BuildNotifiers(c Config) ([]notifier.Notifier, error) {
	method := NotifyType(strings.TrimSpace(string(c.NotifyType)))
	if method != "" {
		n, err := c.buildChannel(method, false)
		if err != nil {
			return nil, err
		}
		if n == nil {
			return nil, fmt.Errorf("通知渠道[%s]参数不完整", method)
		}
		return []notifier.Notifier{n}, nil
	}

	var out []notifier.Notifier
	for _, t := range allNotifyTypes {
		n, err := c.buildChannel(t, true)
		if err != nil {
			return nil, err
		}
		if n != nil {
			out = append(out, n)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("未配置可用通知渠道：请设置 notify_method，或为渠道设置 enabled: true 并补全参数")
	}
	return out, nil
}

func (c Config) buildChannel(t NotifyType, requireEnabled bool) (notifier.Notifier, error) {
	switch t {
	case NotifySCT:
		return c.buildSCT(requireEnabled)
	case NotifyWERobot:
		return c.buildWERobot(requireEnabled)
	case NotifyNotifyx:
		return c.buildNotifyx(requireEnabled)
	case NotifyTelegram:
		return c.buildTelegram(requireEnabled)
	case NotifySMTP:
		return c.buildSMTP(requireEnabled)
	case NotifySendGrid:
		return c.buildSendGrid(requireEnabled)
	case NotifyMailgun:
		return c.buildMailgun(requireEnabled)
	default:
		return nil, fmt.Errorf("不支持的通知渠道：%s", t)
	}
}

func (c Config) buildSCT(requireEnabled bool) (notifier.Notifier, error) {
	key := strings.TrimSpace(c.SCTKey)
	enabled := false
	if c.SCT != nil {
		enabled = c.SCT.Enabled
		if k := strings.TrimSpace(c.SCT.Key); k != "" {
			key = k
		}
	}
	if requireEnabled && !enabled {
		return nil, nil
	}
	if key == "" {
		if requireEnabled {
			return nil, nil
		}
		return nil, fmt.Errorf("sct 需要 sct_key 或 sct.key")
	}
	return notifier.NewSCT(key), nil
}

func (c Config) buildWERobot(requireEnabled bool) (notifier.Notifier, error) {
	webhook := strings.TrimSpace(c.WERobotWebhook)
	chatID := strings.TrimSpace(c.WERobotChatID)
	enabled := false
	if c.WERobot != nil {
		enabled = c.WERobot.Enabled
		if w := strings.TrimSpace(c.WERobot.Webhook); w != "" {
			webhook = w
		}
		if id := strings.TrimSpace(c.WERobot.ChatID); id != "" {
			chatID = id
		}
	}
	if requireEnabled && !enabled {
		return nil, nil
	}
	if webhook == "" {
		if requireEnabled {
			return nil, nil
		}
		return nil, fmt.Errorf("werobot 需要 werobot_webhook 或 werobot.webhook")
	}
	return notifier.NewWERobot(webhook, chatID), nil
}

func (c Config) buildNotifyx(requireEnabled bool) (notifier.Notifier, error) {
	key := strings.TrimSpace(c.NotifyxKey)
	team := strings.TrimSpace(c.NotifyxTeam)
	enabled := false
	if c.Notifyx != nil {
		enabled = c.Notifyx.Enabled
		if k := strings.TrimSpace(c.Notifyx.Key); k != "" {
			key = k
		}
		if t := strings.TrimSpace(c.Notifyx.Team); t != "" {
			team = t
		}
	}
	if requireEnabled && !enabled {
		return nil, nil
	}
	if key == "" {
		if requireEnabled {
			return nil, nil
		}
		return nil, fmt.Errorf("notifyx 需要 notifyx_key 或 notifyx.key")
	}
	return notifier.NewNotifyx(key, team), nil
}

func (c Config) buildTelegram(requireEnabled bool) (notifier.Notifier, error) {
	token := strings.TrimSpace(c.TelegramBotToken)
	userID := strings.TrimSpace(c.TelegramUserID)
	enabled := false
	if c.Telegram != nil {
		enabled = c.Telegram.Enabled
		if t := strings.TrimSpace(c.Telegram.BotToken); t != "" {
			token = t
		}
		if u := strings.TrimSpace(c.Telegram.UserID); u != "" {
			userID = u
		}
	}
	if requireEnabled && !enabled {
		return nil, nil
	}
	if token == "" || userID == "" {
		if requireEnabled {
			return nil, nil
		}
		return nil, fmt.Errorf("telegram 需要 bot_token 与 user_id")
	}
	return notifier.NewTelegram(token, userID), nil
}

func (c Config) buildSMTP(requireEnabled bool) (notifier.Notifier, error) {
	if c.SMTP == nil {
		if requireEnabled {
			return nil, nil
		}
		return nil, fmt.Errorf("smtp 需要配置 smtp 段")
	}
	if requireEnabled && !c.SMTP.Enabled {
		return nil, nil
	}
	host := strings.TrimSpace(c.SMTP.Host)
	from := strings.TrimSpace(c.SMTP.From)
	to := trimList(c.SMTP.To)
	if host == "" || from == "" || len(to) == 0 {
		if requireEnabled {
			return nil, nil
		}
		return nil, fmt.Errorf("smtp 需要 host/from/to")
	}
	return notifier.NewSMTP(notifier.SMTPSettings{
		Host:     host,
		Port:     c.SMTP.Port,
		Username: strings.TrimSpace(c.SMTP.Username),
		Password: c.SMTP.Password,
		From:     from,
		To:       to,
	}), nil
}

func (c Config) buildSendGrid(requireEnabled bool) (notifier.Notifier, error) {
	if c.SendGrid == nil {
		if requireEnabled {
			return nil, nil
		}
		return nil, fmt.Errorf("sendgrid 需要配置 sendgrid 段")
	}
	if requireEnabled && !c.SendGrid.Enabled {
		return nil, nil
	}
	key := strings.TrimSpace(c.SendGrid.APIKey)
	from := strings.TrimSpace(c.SendGrid.From)
	to := trimList(c.SendGrid.To)
	if key == "" || from == "" || len(to) == 0 {
		if requireEnabled {
			return nil, nil
		}
		return nil, fmt.Errorf("sendgrid 需要 api_key/from/to")
	}
	return notifier.NewSendGrid(key, from, to), nil
}

func (c Config) buildMailgun(requireEnabled bool) (notifier.Notifier, error) {
	if c.Mailgun == nil {
		if requireEnabled {
			return nil, nil
		}
		return nil, fmt.Errorf("mailgun 需要配置 mailgun 段")
	}
	if requireEnabled && !c.Mailgun.Enabled {
		return nil, nil
	}
	domain := strings.TrimSpace(c.Mailgun.Domain)
	key := strings.TrimSpace(c.Mailgun.APIKey)
	from := strings.TrimSpace(c.Mailgun.From)
	to := trimList(c.Mailgun.To)
	if domain == "" || key == "" || from == "" || len(to) == 0 {
		if requireEnabled {
			return nil, nil
		}
		return nil, fmt.Errorf("mailgun 需要 domain/api_key/from/to")
	}
	return notifier.NewMailgun(notifier.MailgunSettings{
		Domain:  domain,
		APIKey:  key,
		From:    from,
		To:      to,
		BaseURL: strings.TrimSpace(c.Mailgun.BaseURL),
	}), nil
}

func trimList(in []string) []string {
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
