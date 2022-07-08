package main

import (
	"io/ioutil"

	"go.uber.org/zap"

	"gopkg.in/yaml.v3"
)

type NotifyType string

const (
	NotifySCT     NotifyType = "sct"
	NotifyWERobot NotifyType = "werobot"
	NotifyNextrt  NotifyType = "nextrt"
)

type Config struct {
	WarnRate       float64    `yaml:"warn_rate"`
	ShutdownRate   float64    `yaml:"shutdown_rate"`
	CheckInterval  int64      `yaml:"check_interval"`
	NotifyType     NotifyType `yaml:"notify_method"`
	SCTKey         string     `yaml:"sct_key"`
	WERobotWebhook string     `yaml:"werobot_webhook"`
	WERobotChatID  string     `yaml:"werobot_chatid"`
	NextrtType     string     `yaml:"nextrt_type"`
	NextrtToken    string     `yaml:"nextrt_token"`
	Accounts       []account  `yaml:"accounts"`
}

type account struct {
	Name      string   `yaml:"name"`
	SecretID  string   `yaml:"secret_id"`
	SecretKey string   `yaml:"secret_key"`
	Regions   []string `yaml:"regions"`
}

func InitConfig(file string) {
	var conf Config
	yamlByte, err := ioutil.ReadFile(file)
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
