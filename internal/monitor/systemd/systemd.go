package systemd

import (
	"bytes"
	"fmt"
	"op-tool/internal/monitor"
	"os/exec"
)

// ServiceMonitor 监控一个具体的 systemd 服务
type ServiceMonitor struct {
	serviceName string
}

// New 创建一个新的 systemd 服务监控器
func New(serviceName string) *ServiceMonitor {
	return &ServiceMonitor{serviceName: serviceName}
}

// Name 返回监控器名称
func (s *ServiceMonitor) Name() string {
	return fmt.Sprintf("systemd-service(%s)", s.serviceName)
}

// Check 使用 `systemctl is-active` 命令检查服务状态
func (s *ServiceMonitor) Check() monitor.Result {
	cmd := exec.Command("systemctl", "is-active", "--quiet", s.serviceName)
	
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		// `is-active` 命令在服务不活跃时会返回非零退出码
		return monitor.Result{
			Success: false,
			Message: fmt.Sprintf("服务 %s 状态异常或不存在。", s.serviceName),
			Err:     err,
		}
	}

	return monitor.Result{
		Success: true,
		Message: fmt.Sprintf("服务 %s 运行正常。", s.serviceName),
	}
}