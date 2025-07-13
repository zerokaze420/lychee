package lark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// LarkNotifier 实现了 notifier.Notifier 接口，用于发送飞书消息
type LarkNotifier struct {
	webhookURL string
	client     *http.Client
}

// New 创建一个新的飞书通知器
func New(webhookURL string) *LarkNotifier {
	return &LarkNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}
}

// 飞书消息卡片的结构体
type larkPayload struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

// Notify 发送通知到飞书
func (n *LarkNotifier) Notify(ctx context.Context, subject, message string) error {
	fullMessage := fmt.Sprintf("【%s】\n%s", subject, message)
	
	payload := larkPayload{
		MsgType: "text",
		Content: struct {
			Text string `json:"text"`
		}{
			Text: fullMessage,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化飞书消息失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", n.webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("创建飞书请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送飞书消息失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("发送飞书消息失败，状态码: %d", resp.StatusCode)
	}

	return nil
}