package main

import (
	"context"
	"flag" // 1. 导入 flag 包
	"hashcowuwu/lychee/internal/config"
	"hashcowuwu/lychee/internal/monitor"
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

	var notif notifier.Notifier = lark.New(cfg.Lark.WebhookURL)
	var monitors []monitor.Monitor
	for _, serviceName := range cfg.Systemd.Services {
		monitors = append(monitors, systemd.New(serviceName))
	}

	log.Println("运维监控工具启动...")
	ticker := time.NewTicker(time.Duration(cfg.CheckInterval) * time.Second)
	defer ticker.Stop()

	// 首次立即执行
	failedMessages := runChecks(monitors)
	sendAggregateNotification(notif, failedMessages)

	// 之后按定时器周期执行
	for range ticker.C {
		failedMessages := runChecks(monitors)
		sendAggregateNotification(notif, failedMessages)
	}
}

// runChecks 函数保持不变
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

	err := notif.Notify(context.Background(), "🚨 多个服务异常告警", fullMessage)
	if err != nil {
		log.Printf("发送聚合通知失败: %v\n", err)
	}

}
