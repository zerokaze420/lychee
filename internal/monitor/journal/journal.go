package journal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"hashcowuwu/lychee/internal/monitor"
	"log"
	"os/exec"
	"regexp" // 導入 regexp 包
	"strings"
)

// JournalEntry 對應 journalctl -o json 輸出的單條日誌結構。
// 我們只需要 __CURSOR 和 MESSAGE 兩個字段。
type JournalEntry struct {
	Cursor  string `json:"__CURSOR"`
	Message string `json:"MESSAGE"`
}

// JournalMonitor 從 systemd journal 中讀取特定服務的日志
// 它通過執行 journalctl 命令並管理 cursor 來實現，完全不依賴 CGO。
type JournalMonitor struct {
	serviceName string
	keywords    []string
	// cursor 用於記錄上次讀取到的日誌位置，以便下次只讀取新的日誌。
	cursor string
}

// New 創建一個新的 JournalMonitor 實例
func New(serviceName string, keywords []string) (monitor.Monitor, error) {
	// 創建一個基礎的 monitor 實例
	jm := &JournalMonitor{
		serviceName: serviceName,
		keywords:    keywords,
	}

	// 初始化 cursor，將其設置為當前服務最新一條日誌的位置。
	// 這相當於原代碼中的 j.SeekTail() + j.Next()。
	// 我們只獲取最新的一條 (-n 1) 來拿到它的 cursor。
	cmd := exec.Command("journalctl", "-u", serviceName, "-n", "1", "-o", "json", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		// 如果命令執行失敗（例如服務不存在或還沒有任何日誌），我們不將其視為致命錯誤。
		// cursor 將為空，第一次 Check() 會從頭讀取（或讀取最近的日誌）。
		// 這裡可以根據您的需求決定是否返回錯誤。
		// 作為監控，允許服務初期沒有日誌是合理的。
		log.Printf("注意: 為服務 [%s] 初始化 cursor 失敗 (可能是新服務無日誌): %v", serviceName, err)
	} else {
		// 從輸出中解析最後一條日誌的 cursor
		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		if scanner.Scan() {
			var entry JournalEntry
			if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
				jm.cursor = entry.Cursor
			}
		}
	}

	return jm, nil
}

// Name 返回監控器的名稱
func (jm *JournalMonitor) Name() string {
	return fmt.Sprintf("journal-%s", jm.serviceName)
}

// Check 從上次的位置開始，檢查新的日誌條目
func (jm *JournalMonitor) Check() monitor.Result {
	var matchedMessages []string

	// 準備 journalctl 命令的參數
	args := []string{"-u", jm.serviceName, "-o", "json", "--no-pager"}
	if jm.cursor != "" {
		// 如果我們有 cursor，就只讀取它之後的日誌
		args = append(args, "--after-cursor", jm.cursor)
	}

	cmd := exec.Command("journalctl", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return monitor.Result{Success: false, Message: fmt.Sprintf("服務 [%s] 無法創建命令管道: %v", jm.serviceName, err)}
	}

	if err := cmd.Start(); err != nil {
		return monitor.Result{Success: false, Message: fmt.Sprintf("服務 [%s] 無法啟動 journalctl: %v", jm.serviceName, err)}
	}

	// 逐行讀取新日誌
	scanner := bufio.NewScanner(stdout)
	var lastReadCursor string
	for scanner.Scan() {
		line := scanner.Bytes()

		var entry JournalEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			// 忽略無法解析的行
			continue
		}

		// 更新我們在此次檢查中讀到的最後一個 cursor
		lastReadCursor = entry.Cursor

		// 檢查關鍵字 - 使用正則表達式
		for _, keyword := range jm.keywords {
			// 编译正则表达式。(?i) 表示不區分大小寫匹配。
			// 如果您的關鍵字可能包含正則表達式特殊字符，但您希望它們被字面匹配，
			// 則應使用 regexp.QuoteMeta(keyword) 進行轉義。
			// 例如：re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(keyword))
			re, err := regexp.Compile("(?i)" + keyword)
			if err != nil {
				log.Printf("警告: 服務 [%s] 關鍵字 '%s' 正則表達式編譯失敗: %v", jm.serviceName, keyword, err)
				continue // 跳過當前無效的關鍵字，繼續處理下一個
			}

			if re.MatchString(entry.Message) {
				msg := fmt.Sprintf("服務 [%s] journal 日誌發現關鍵字 '%s': %s", jm.serviceName, keyword, entry.Message)
				matchedMessages = append(matchedMessages, msg)
				break // 找到一個關鍵字就足夠了，處理下一條日誌
			}
		}
	}

	// 等待命令結束
	if err := cmd.Wait(); err != nil {
		// journalctl 在沒有新日誌時可能會以非 0 狀態碼退出，這裡可以更寬容地處理
		// 但如果管道讀取正常，通常可以忽略 wait 的錯誤
		// 只有當 err 不是 ExitError 且不是 0 狀態碼時才記錄為錯誤
		if exitErr, ok := err.(*exec.ExitError); ok {
			// 如果是 journalctl 正常退出但沒有新日誌，其退出碼可能為非零
			// 這裡可以根據需要調整錯誤處理邏輯
			log.Printf("注意: 服務 [%s] journalctl 命令退出，狀態碼: %d", jm.serviceName, exitErr.ExitCode())
		} else {
			log.Printf("錯誤: 服務 [%s] journalctl 命令等待失敗: %v", jm.serviceName, err)
		}
	}

	// 如果我們讀到了新日誌，就更新 monitor 的 cursor 狀態
	if lastReadCursor != "" {
		jm.cursor = lastReadCursor
	}

	if len(matchedMessages) > 0 {
		return monitor.Result{
			Success: false, // 發現關鍵字通常表示非成功狀態
			Message: strings.Join(matchedMessages, "\n"),
		}
	}

	return monitor.Result{Success: true}
}
