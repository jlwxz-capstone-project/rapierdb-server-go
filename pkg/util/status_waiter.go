package util

import (
	"time"
)

// WaitForStatus 等待目标状态，返回一个通道，当达到目标状态时会收到通知
// currentStatus: 获取当前状态的函数
// targetStatus: 目标状态
// statusCh: 状态变更事件通道
// cleanup: 清理函数，用于取消订阅等操作
// timeout: 等待超时时间，如果为0则永不超时
func WaitForStatus(
	currentStatus func() string,
	targetStatus string,
	statusCh <-chan any,
	cleanup func(),
	timeout time.Duration,
) <-chan struct{} {
	doneCh := make(chan struct{})

	// 如果当前已经是目标状态，直接返回一个已关闭的通道
	if currentStatus() == targetStatus {
		close(doneCh)
		return doneCh
	}

	// 在后台监听状态变更
	go func() {
		defer cleanup()
		defer close(doneCh)

		// 如果没有设置超时，一直等待直到状态变更
		if timeout <= 0 {
			for status := range statusCh {
				if status.(string) == targetStatus {
					return
				}
			}
			return // 通道关闭
		}

		// 设置了超时，使用select等待状态变更或超时
		timeoutCh := time.After(timeout)
		for {
			select {
			case status, ok := <-statusCh:
				if !ok {
					// 通道已关闭
					return
				}
				if status.(string) == targetStatus {
					return
				}
			case <-timeoutCh:
				// 超时
				return
			}
		}
	}()

	return doneCh
}
