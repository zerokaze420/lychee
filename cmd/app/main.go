package main

import (
	"context"
	"log"
	"op-tool/internal/config"
	"op-tool/internal/monitor"
	"op-tool/internal/monitor/systemd"
	"op-tool/internal/notifier"
	"op-tool/internal/notifier/lark"
	"time"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}

	// 2. 初始化通知器
	// 由于我们只有一个通知器，所以直接创建。
	// 如果未来有多个，可以根据配置来选择创建哪个。
	var notif notifier.Notifier = lark.New(cfg.Lark.WebhookURL)

	// 3. 根据配置创建所有监控器
	var monitors []monitor.Monitor
	for _, serviceName := range cfg.Systemd.Services {
		monitors = append(monitors, systemd.New(serviceName))
	}

	// 4. 启动定时检查
	log.Println("运维监控工具启动...")
	ticker := time.NewTicker(time.Duration(cfg.CheckInterval) * time.Second)
	defer ticker.Stop()

	// 立即执行一次检查，不等第一个 ticker
	runChecks(monitors, notif)

	for range ticker.C {
		runChecks(monitors, notif)
	}
}

// runChecks 遍历所有监控器并执行检查
func runChecks(monitors []monitor.Monitor, notif notifier.Notifier) {
	log.Println("开始执行新一轮监控检查...")
	for _, m := range monitors {
		result := m.Check()
		log.Printf("监控器 [%s]: 状态=%t, 消息=%s\n", m.Name(), result.Success, result.Message)

		// 如果检查失败，发送通知
		if !result.Success {
			err := notif.Notify(context.Background(), "🚨 服务异常告警", result.Message)
			if err != nil {
				log.Printf("为监控器 [%s] 发送通知失败: %v\n", m.Name(), err)
			}
		}
	}
}