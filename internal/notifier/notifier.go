package notifier

import "context"

// Notifier 定义了所有通知器的通用接口
type Notifier interface {
	// Notify 发送一条通知
	Notify(ctx context.Context, subject, message string) error
}
