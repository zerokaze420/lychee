package main

import (
	"context"
	"flag" // 1. 导入 flag 包
	"log"
	"hashcowuwu/lychee/internal/config"
	"hashcowuwu/lychee/internal/monitor"
	"hashcowuwu/lychee/internal/monitor/systemd"
	"hashcowuwu/lychee/internal/notifier"
	"hashcowuwu/lychee/internal/notifier/lark"
	"time"
)

func main() {
	// 2. 定义一个名为 "config" 的命令行标志，用于接收配置文件路径
	// - 第一个参数是标志名称 ("config")
	// - 第二个参数是默认值 ("config.yaml")，如果用户不提供该标志，则使用此值
	// - 第三个参数是帮助信息
	configPath := flag.String("config", "config.yaml", "path to the configuration file")

	// 3. 解析用户在命令行中提供的标志
	flag.Parse()

	// 4. 使用从标志中获取的路径加载配置 (*configPath)
	// 注意：flag.String 返回的是指针，所以我们需要用 * 来获取它的值
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}

	// 后面的代码保持不变...
	var notif notifier.Notifier = lark.New(cfg.Lark.WebhookURL)
	var monitors []monitor.Monitor
	for _, serviceName := range cfg.Systemd.Services {
		monitors = append(monitors, systemd.New(serviceName))
	}

	log.Println("运维监控工具启动...")
	ticker := time.NewTicker(time.Duration(cfg.CheckInterval) * time.Second)
	defer ticker.Stop()

	runChecks(monitors, notif)

	for range ticker.C {
		runChecks(monitors, notif)
	}
}

// runChecks 函数保持不变
func runChecks(monitors []monitor.Monitor, notif notifier.Notifier) {
	log.Println("开始执行新一轮监控检查...")
	for _, m := range monitors {
		result := m.Check()
		log.Printf("监控器 [%s]: 状态=%t, 消息=%s\n", m.Name(), result.Success, result.Message)

		if !result.Success {
			err := notif.Notify(context.Background(), "🚨 服务异常告警", result.Message)
			if err != nil {
				log.Printf("为监控器 [%s] 发送通知失败: %v\n", m.Name(), err)
			}
		}
	}
}