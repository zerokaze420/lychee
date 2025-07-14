package journal

import (
	"fmt"
	"hashcowuwu/lychee/internal/monitor"
	"strings"

	"github.com/coreos/go-systemd/v22/sdjournal"
)

// JournalMonitor 从 systemd journal 中读取特定服务的日志
type JournalMonitor struct {
	serviceName string
	keywords    []string
	journal     *sdjournal.Journal
}

// New 创建一个新的 JournalMonitor 实例
func New(serviceName string, keywords []string) (monitor.Monitor, error) {
	// 创建一个新的 journal reader
	j, err := sdjournal.NewJournal()
	if err != nil {
		return nil, fmt.Errorf("无法打开 journal: %w", err)
	}

	// 添加过滤器，只匹配特定服务的日志
	// 这相当于 journalctl -u <serviceName>
	match := sdjournal.Match{
		Field: sdjournal.SD_JOURNAL_FIELD_SYSTEMD_UNIT,
		Value: serviceName,
	}
	if err := j.AddMatch(match.String()); err != nil {
		j.Close()
		return nil, fmt.Errorf("为服务 [%s] 添加 journal 过滤器失败: %w", serviceName, err)
	}

	// 将光标移动到日志的末尾，这样我们只会读取未来的新日志
	if err := j.SeekTail(); err != nil {
		j.Close()
		return nil, fmt.Errorf("无法将 journal 光标移动到末尾: %w", err)
	}
	// 为了让第一次 Check() 不读取旧日志，我们先空读一次，将光标确实移动到最后一条之后
	j.Next()

	return &JournalMonitor{
		serviceName: serviceName,
		keywords:    keywords,
		journal:     j,
	}, nil
}

func (jm *JournalMonitor) Name() string {
	return fmt.Sprintf("journal-%s", jm.serviceName)
}

// Check 从上次的位置开始，检查新的日志条目
func (jm *JournalMonitor) Check() monitor.Result {
	var matchedMessages []string

	// 循环读取自上次检查以来所有新的日志条目
	for {
		// j.Next() 会将光标向前移动一条。如果没有新日志，它会返回 0。
		r, err := jm.journal.Next()
		if err != nil {
			return monitor.Result{Success: false, Message: fmt.Sprintf("服务 [%s] 读取 journal 出错: %v", jm.serviceName, err)}
		}
		if r == 0 {
			// 没有更多新条目了
			break
		}

		// 获取日志条目
		entry, err := jm.journal.GetEntry()
		if err != nil {
			return monitor.Result{Success: false, Message: fmt.Sprintf("服务 [%s] 获取 journal 条目失败: %v", jm.serviceName, err)}
		}

		// 获取日志消息正文
		message, ok := entry.Fields[sdjournal.SD_JOURNAL_FIELD_MESSAGE]
		if !ok {
			continue // 如果没有消息字段，跳过
		}

		// 检查关键字
		for _, keyword := range jm.keywords {
			if strings.Contains(strings.ToLower(message), strings.ToLower(keyword)) {
				msg := fmt.Sprintf("服务 [%s] journal 日志发现关键字 '%s': %s", jm.serviceName, keyword, message)
				matchedMessages = append(matchedMessages, msg)
				break
			}
		}
	}

	// 等待新日志。这使得检查是非阻塞的，并且在没有新日志时能快速返回。
	jm.journal.Wait(0)

	if len(matchedMessages) > 0 {
		return monitor.Result{
			Success: false,
			Message: strings.Join(matchedMessages, "\n"),
		}
	}

	return monitor.Result{Success: true}
}
