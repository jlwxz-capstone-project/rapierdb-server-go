package synchronizer

import (
	"bytes"
	"fmt"
	"time"
)

// SSEConfig SSE 编码配置
type SSEConfig struct {
	Event string // 事件类型
	ID    string // 消息ID
	Retry int    // 重试时间（毫秒）
}

// EncodeSSE 将消息编码为SSE格式
// 使用 config 参数进行配置，传 nil 时使用默认配置
func EncodeSSE(data []byte, config *SSEConfig) []byte {
	// 设置默认值
	cfg := SSEConfig{
		ID:    fmt.Sprintf("%d", time.Now().UnixNano()),
		Retry: 5000,
	}

	// 合并自定义配置
	if config != nil {
		if config.Event != "" {
			cfg.Event = config.Event
		}
		if config.ID != "" {
			cfg.ID = config.ID
		}
		if config.Retry > 0 {
			cfg.Retry = config.Retry
		}
	}

	var buf bytes.Buffer
	if cfg.Event != "" {
		buf.WriteString(fmt.Sprintf("event: %s\n", cfg.Event))
	}
	buf.WriteString(fmt.Sprintf("id: %s\n", cfg.ID))
	buf.WriteString(fmt.Sprintf("retry: %d\n", cfg.Retry))

	// 处理多行数据
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		buf.WriteString(fmt.Sprintf("data: %s\n", line))
	}
	buf.WriteString("\n")

	return buf.Bytes()
}
