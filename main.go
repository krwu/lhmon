package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
	TB = GB * 1024
)

var (
	Conf Config
)

func main() {
	file := flag.String("conf", "/etc/lhmon/conf.yml", "配置文件路径")
	flag.Parse()
	if file == nil || *file == "" {
		log.Fatalf("必须指定配置文件")
	}

	InitConfig(*file)
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
		log.Printf("%s\n", strings.Join(results, "\n"))
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
			log.Printf("[%s] 账号：%s, 区域：%s, 实例：%s, 使用率：%.2f，已用：%s，总共：%s\n",
				time.Now().Format("2006-01-02 15:04:05"),
				a.Name,
				r,
				pkg.InstanceID,
				pkg.UseRate(),
				calcTraffic(pkg.Used),
				calcTraffic(pkg.Total),
			)

			if pkg.UseRate() == 0 { // 使用率为0直接跳过
				break
			}

			if Conf.ShutdownRate > 0 && pkg.UseRate() > Conf.ShutdownRate {
				c.ShutdownInstance(pkg.InstanceID)
				ch<- fmt.Sprintf("- 账号[%s] 实例[%s] 使用率[%.2f]，执行关机！", a.Name, pkg.InstanceID, pkg.UseRate())
			}

			if Conf.WarnRate > 0 && pkg.UseRate() > Conf.WarnRate {
				ch<- fmt.Sprintf("- 账号[%s] 实例[%s] 使用率[%.2f]，请关注！", a.Name, pkg.InstanceID, pkg.UseRate())
			}
		}
	}
}

func notify(format string, args ...interface{}) {
	client := http.Client{
		Timeout:       time.Second * 3,
	}
	desp := fmt.Sprintf(format, args...)
	body := url.Values{}
	body.Add("title", "腾讯云轻量监控通知")
	body.Add("desp", desp)
	api := fmt.Sprintf("https://sctapi.ftqq.com/%s.send", Conf.SCTKey)
	req, err := http.NewRequest(http.MethodPost, api, strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		log.Printf("[%s] ERROR %v\n", time.Now().Format("20060102 15:04:05"), err)
		return
	}
	_, err = client.Do(req)
	if err != nil {
		log.Printf("[%s] ERROR %v\n", time.Now().Format("20060102 15:04:05"), err)
		return
	}
}

func calcTraffic(bytes int64) string {
	var result float64
	if bytes > TB && bytes % TB == 0 {
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
