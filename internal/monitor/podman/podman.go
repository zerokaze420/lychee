package podman

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	podmanClient "github.com/containers/podman/v4/pkg/api/client" // 注意这里的 v4
	// 注意这里的 v4
	// 保持不变
	// 保持不变
)

type ContainerState struct {
	ContainerID    string
	Name           string
	CPUUsage       float64
	MemoryUsage    float64
	NetworkRxBytes uint64
	NetworkTxBytes uint64
	Timestamp      time.Time
}

type ContainerMonitorConfig struct {
	Interval         time.Duration
	TargetContainers []string
	AlertThresholds  map[string]float64
	SocketPath       string
}

type ContainerMoitor struct {
	config       ContainerMonitorConfig
	mu           sync.RWMutex
	alertCh      chan Alert
	stopCh       chan struct{}
	podmanClient *podmanClient.Client
}

type Alert struct {
	ContainerID string
	Metric      string
	Value       float64
	Threshold   float64
	Message     string
	Timestamp   time.Time
}

func NewContainerMonitor(config ContainerMonitorConfig) (*ContainerMoitor, error) {
	if config.SocketPath == "" {
		config.SocketPath = "unix:/run/podman/podman.sock"
		log.Printf("Podman SocketPath 未指定，尝试使用默认路径: %s。请在 config 中明确指定以避免问题。", config.SocketPath)
	}

}

func (m *ContainerMoitor) Start(ctx context.Context) {
	log.Println("容器监控已经启动")
	ticker := time.NewTicker(m.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("正在收集容器指标...")
			stats, err := m.collectMetrics(ctx)
			if err != nil {
				log.Printf("收集指标时出错: %v", err)
				continue
			}
			m.processMetrics(stats)
		case <-m.stopCh:
			log.Println("容器已经显示停止")
			return
		case <-ctx.Done():
			log.Println("容器已通过上下文取消停止")
			return
		}
	}
}

func (m *ContainerMoitor) Stop() {
	close(m.stopCh)
}

func (m *ContainerMoitor) collectMetrics(ctx context.Context) ([]ContainerState, error) {
	var allStats []ContainerState

	simulatedContainers := []struct{ ID, Name string }{
		{"container1", "my-app-db"},
		{"container2", "my-app-web"},
	}

	for _, sc := range simulatedContainers {
		if len(m.config.AlertThresholds) > 0 {
			found := false
			for _, target := range m.config.TargetContainers {
				if target == sc.ID || target == sc.Name {
					found = true
					break
				}
			}
			if !found {
				continue
			}

		}
		cpu := float64(time.Now().UnixNano() % 100)
		mem := float64(time.Now().UnixNano()%100 + 200)

		allStats = append(allStats, ContainerState{
			CPUUsage:       cpu,
			MemoryUsage:    mem,
			Name:           sc.ID,
			ContainerID:    sc.Name,
			NetworkRxBytes: uint64(time.Now().UnixNano() % 1000),
			NetworkTxBytes: uint64(time.Now().UnixNano() % 1000),
			Timestamp:      time.Now(),
		})
	}
	return allStats, nil

}

func (m *ContainerMoitor) processMetrics(stats []ContainerState) {
	for _, s := range stats {
		log.Printf("容器 [%s]: CPU使用率=%f%%, 内存使用率=%f", s.Name, s.CPUUsage, s.MemoryUsage)

		if threshold, ok := m.config.AlertThresholds["cpu-high"]; ok && s.CPUUsage > threshold {
			m.alertCh <- Alert{
				ContainerID: s.ContainerID,
				Message:     fmt.Sprintf("CPU使用率过高: %f%%", s.CPUUsage),
				Metric:      "CPUUsage",
				Threshold:   threshold,
				Timestamp:   s.Timestamp,
				Value:       s.CPUUsage,
			}
		}
		if threshold, ok := m.config.AlertThresholds["mem-high"]; ok && s.MemoryUsage > threshold {
			m.alertCh <- Alert{
				ContainerID: s.ContainerID,
				Message:     fmt.Sprintf("内存使用率过高: %f%%", s.MemoryUsage),
				Metric:      "MemoryUsageMB",
				Threshold:   threshold,
				Timestamp:   s.Timestamp,
				Value:       s.MemoryUsage,
			}
		}
	}

}

func (m *ContainerMoitor) GetAlerts() <-chan Alert {
	return m.alertCh
}
