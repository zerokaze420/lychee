package lark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io" // 导入 io 包用于读取响应体
	"net/http"
)

// LarkNotifier 实现了 notifier.Notifier 接口，用于发送飞书消息
type LarkNotifier struct {
	WebhookURLs []string
	client      *http.Client
}

// New 创建一个新的飞书通知器
func New(webhookURLs []string) *LarkNotifier {
	return &LarkNotifier{
		WebhookURLs: webhookURLs,
		client:      &http.Client{},
	}
}

// 飞书消息卡片的公共结构体
type larkPayload struct {
	MsgType string      `json:"msg_type"`
	Content interface{} `json:"content"` // 使用 interface{} 来适应不同消息类型的内容
}

// 文本消息内容结构体（如果还需要发送纯文本消息）
type textContent struct {
	Text string `json:"text"`
}

// Markdown 消息内容结构体
type postContent struct {
	Post struct {
		ZhCn struct {
			Title   string `json:"title"`
			Content [][]struct {
				Tag  string `json:"tag"`
				Text string `json:"text,omitempty"` // 对于 text 类型的标签
				// 您可以在这里添加其他 Markdown 元素字段，例如：
				// Href string `json:"href,omitempty"` // for a tag
				// UserId string `json:"user_id,omitempty"` // for at tag
				// Email string `json:"email,omitempty"` // for at tag
			} `json:"content"`
		} `json:"zh_cn"`
	} `json:"post"`
}

// Notify 发送通知到飞书
// subject 将作为 Markdown 卡片的标题
// message 将作为 Markdown 文本内容
func (n *LarkNotifier) Notify(ctx context.Context, subject, message string) error {
	// 构造 Markdown 消息内容
	// 这里的示例将整个 message 作为一行简单的 Markdown 文本
	// 如果需要更复杂的 Markdown 结构，你需要根据 message 的内容自行解析和构建 content 数组
	if len(n.WebhookURLs) == 0 {
		return fmt.Errorf("没有配置飞书 Webhook URL，无法发送通知")
	}
	markdownContent := postContent{
		Post: struct {
			ZhCn struct {
				Title   string `json:"title"`
				Content [][]struct {
					Tag  string `json:"tag"`
					Text string `json:"text,omitempty"`
				} `json:"content"`
			} `json:"zh_cn"`
		}{
			ZhCn: struct {
				Title   string `json:"title"`
				Content [][]struct {
					Tag  string `json:"tag"`
					Text string `json:"text,omitempty"`
				} `json:"content"`
			}{
				Title: subject, // 将 subject 用作卡片标题
				Content: [][]struct { // Markdown 内容是一个二维数组
					Tag  string `json:"tag"`
					Text string `json:"text,omitempty"`
				}{
					{ // 每一行是一个数组
						{
							Tag:  "text",  // Markdown 元素的类型，例如 "text", "a", "at", "img" 等
							Text: message, // 您的 Markdown 格式的文本内容
						},
					},
				},
			},
		},
	}

	payload := larkPayload{
		MsgType: "post", // 消息类型设置为 "post" 表示发送富文本（Markdown）消息
		Content: markdownContent,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化飞书消息失败: %w", err)
	}

	var allErrors []error
	for _, url := range n.WebhookURLs {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("创建飞书请求失败 (%s): %w", url, err))
			continue // 继续尝试下一个 URL
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := n.client.Do(req)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("发送飞书消息失败 (%s): %w", url, err))
			continue // 继续尝试下一个 URL
		}
		defer resp.Body.Close() // 确保每次循环都关闭响应体

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			allErrors = append(allErrors, fmt.Errorf("发送飞书消息失败 (%s)，状态码: %d, 响应: %s", url, resp.StatusCode, string(respBody)))
			continue // 继续尝试下一个 URL
		}
		// 如果成功发送到某个 URL，可以根据需求选择是否记录成功
		// fmt.Printf("成功发送消息到 %s\n", url)
	}
	if len(allErrors) > 0 {
		// 返回所有错误的聚合信息
		return fmt.Errorf("向部分或所有飞书 Webhook 发送消息失败: %v", allErrors)
	}

	return nil // 所有 URL 都成功发送

}

// 如果你需要同时支持文本和 Markdown 消息，可以添加一个额外的函数
// 例如：
/*
func (n *LarkNotifier) NotifyText(ctx context.Context, subject, message string) error {
    fullMessage := fmt.Sprintf("【%s】\n%s", subject, message)

    payload := larkPayload{
        MsgType: "text",
        Content: textContent{
            Text: fullMessage,
        },
    }

    body, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("序列化飞书文本消息失败: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", n.webhookURL, bytes.NewBuffer(body))
    if err != nil {
        return fmt.Errorf("创建飞书文本请求失败: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := n.client.Do(req)
    if err != nil {
        return fmt.Errorf("发送飞书文本消息失败: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        respBody, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("发送飞书文本消息失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
    }

    return nil
}
*/
