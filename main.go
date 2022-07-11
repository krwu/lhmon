package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/automaxprocs/maxprocs"

	"lighthouse-monitor/log"
	"lighthouse-monitor/notifier"

	_ "go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
	TB = GB * 1024
)

var (
	Conf   Config
	logger = log.Logger()
)

func main() {
	path := "./conf.yml"
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		path = "/etc/lhmon/conf.yml"
	}
	file := flag.String("conf", "/etc/lhmon/conf.yml", "配置文件路径")
	flag.Parse()
	if file == nil || *file == "" {
		log.Fatalf("必须指定配置文件")
	}

	InitConfig(*file)
	logger.Info("started", zap.Int("帐号数", len(Conf.Accounts)))
	go cronTask()

	interval := time.Duration(Conf.CheckInterval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		go cronTask()
	}
}

func cronTask() {
	var wg sync.WaitGroup
	ch := make(chan string, 1000)
	for _, a := range Conf.Accounts {
		wg.Add(1)
		ac := a
		go checkTraffic(ac, ch, &wg)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	var results []string
	for str := range ch {
		results = append(results, str)
	}

	if len(results) > 0 {
		log.Printf(strings.Join(results, "\n"))
		notify("%s\n", strings.Join(results, "\n\n"))
	}
}

func checkTraffic(a account, ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	id := a.SecretID
	key := a.SecretKey
	for _, r := range a.Regions {
		c := NewLighthouseClient(id, key, r)
		pkgs := c.ListTrafficPackages()
		for _, pkg := range pkgs {
			logger.Info("result",
				zap.String("帐号", a.Name),
				zap.String("区域", r),
				zap.String("实例", pkg.InstanceID),
				zap.String("使用率", fmt.Sprintf("%.2f", pkg.UseRate())),
				zap.String("已用", calcTraffic(pkg.Used)),
				zap.String("总共", calcTraffic(pkg.Total)),
			)

			if pkg.UseRate() == 0 { // 使用率为0直接跳过
				break
			}

			if Conf.ShutdownRate > 0 && pkg.UseRate() > Conf.ShutdownRate {
				c.ShutdownInstance(pkg.InstanceID)
				ch <- fmt.Sprintf("- 账号[%s] 实例[%s] 使用率[%.2f]，执行关机！", a.Name, pkg.InstanceID, pkg.UseRate())
			}

			if Conf.WarnRate > 0 && pkg.UseRate() > Conf.WarnRate {
				ch <- fmt.Sprintf("- 账号[%s] 实例[%s] 使用率[%.2f]，请关注！", a.Name, pkg.InstanceID, pkg.UseRate())
			}
		}
	}
}

func notify(format string, args ...interface{}) {
	desp := fmt.Sprintf(format, args...)
	title := "腾讯云轻量监控通知"
	message := desp
	var client notifier.Notifier
	switch Conf.NotifyType {
	case NotifySCT:
		client = notifier.NewSCT(Conf.SCTKey)
	case NotifyWERobot:
		client = notifier.NewWERobot(Conf.WERobotWebhook, Conf.WERobotChatID)
	case NotifyNextrt:
		client = notifier.NewNextrt(Conf.NextrtType, Conf.NextrtToken)
	default:
		log.Errorf("%s：%v", "不支持的通知渠道", Conf.NotifyType)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := client.Send(ctx, title, message)
	if err != nil {
		log.Errorf("%v", err)
		return
	}
}

func calcTraffic(bytes int64) string {
	var result float64
	if bytes > TB && bytes%TB == 0 {
		result = float64(bytes) / TB
		return fmt.Sprintf("%.02fTB", result)
	}

	if bytes > GB {
		result = float64(bytes) / GB
		return fmt.Sprintf("%.02fGB", result)
	}

	if bytes > MB {
		result = float64(bytes) / MB
		return fmt.Sprintf("%.02fMB", result)
	}

	if bytes > KB {
		result = float64(bytes) / KB
		return fmt.Sprintf("%.02fKB", result)
	}

	return fmt.Sprintf("%dB", bytes)
}

func init() {
	maxprocs.Set(maxprocs.Logger(func(string, ...interface{}) {}))
}
