package main

import (
	"context"
	"flag"
	"hashcowuwu/lychee/internal/config"
	"hashcowuwu/lychee/internal/monitor"
	"hashcowuwu/lychee/internal/monitor/journal"
	"hashcowuwu/lychee/internal/monitor/systemd"
	"hashcowuwu/lychee/internal/notifier"
	"hashcowuwu/lychee/internal/notifier/lark"
	"log"
	"strings"
	"time"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to the configuration file")
	flag.Parse()
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}
	log.Printf("加载的 checkInterval: %d", cfg.CheckInterval)

	var notif notifier.Notifier = lark.New(cfg.Lark.WebhookURLs)
	var monitors []monitor.Monitor
	for _, serviceName := range cfg.Systemd.Services {
		monitors = append(monitors, systemd.New(serviceName))
	}

	for _, journalCfg := range cfg.Journal {
		log.Printf("为服务 [%s] 设置 journal 日志监控, 关键字: %v", journalCfg.ServiceName, journalCfg.Keywords)
		m, err := journal.New(journalCfg.ServiceName, journalCfg.Keywords)
		if err != nil {
			log.Printf("警告: 无法为服务 [%s] 创建 journal 监控器: %v", journalCfg.ServiceName, err)
			continue
		}
		monitors = append(monitors, m)
	}

	log.Println("运维监控工具启动...")
	if cfg.CheckInterval <= 0 {
		cfg.CheckInterval = 60
		log.Printf("checkInterval 非正数，使用默认值: %ds", cfg.CheckInterval)
	}
	interval := time.Duration(cfg.CheckInterval) * time.Second
	log.Printf("计时器间隔: %v", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	failedMessages := runChecks(monitors)
	sendAggregateNotification(notif, failedMessages)

	for range ticker.C {
		failedMessages := runChecks(monitors)
		sendAggregateNotification(notif, failedMessages)
	}
}

func runChecks(monitors []monitor.Monitor) []string {
	log.Println("开始执行新一轮监控检查...")
	var failedMessages []string
	for _, m := range monitors {
		result := m.Check()
		log.Printf("监控器 [%s]: 状态=%t, 消息=%s\n", m.Name(), result.Success, result.Message)
		if !result.Success {
			failedMessages = append(failedMessages, result.Message)
		}
	}
	if len(failedMessages) == 0 {
		log.Println("所有服务状态正常。")
	}
	return failedMessages
}

func sendAggregateNotification(notif notifier.Notifier, messages []string) {
	if len(messages) == 0 {
		return
	}
	fullMessage := "以下服务出现异常:\n" + strings.Join(messages, "\n")
	log.Println("发现服务异常，准备发送聚合通知...")
	err := notif.Notify(context.Background(), "🚨 服务异常告警", fullMessage)
	if err != nil {
		log.Printf("发送聚合通知失败: %v\n", err)
	}
}
