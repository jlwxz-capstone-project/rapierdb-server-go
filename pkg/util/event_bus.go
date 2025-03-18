package util

import (
	"sync"
)

type EventBus[E any] struct {
	subscribers []chan E     // 每个订阅者对应一个 Channel
	lk          sync.RWMutex // 保护并发访问 subscribers 的锁
}

// NewEventBus 创建并返回一个新的事件总线实例
// 类型参数 E 代表事件数据的类型
func NewEventBus[E any]() *EventBus[E] {
	return &EventBus[E]{
		subscribers: make([]chan E, 0),
		lk:          sync.RWMutex{},
	}
}

// Subscribe 订阅事件
// 返回一个接收通道，通过该通道可以获取发布的事件
// 通道带有缓冲区，可以缓存最近的 100 个事件
func (eb *EventBus[E]) Subscribe() <-chan E {
	eb.lk.Lock()
	defer eb.lk.Unlock()

	ch := make(chan E, 100) // 带缓冲的通道
	eb.subscribers = append(eb.subscribers, ch)
	return ch
}

// Unsubscribe 取消订阅
// 会关闭对应的通道并从订阅者列表中移除
func (eb *EventBus[E]) Unsubscribe(ch <-chan E) {
	eb.lk.Lock()
	defer eb.lk.Unlock()

	for i := range eb.subscribers {
		if eb.subscribers[i] == ch {
			// 关闭通道并移除订阅者
			close(eb.subscribers[i])
			eb.subscribers = append(eb.subscribers[:i], eb.subscribers[i+1:]...)
			break
		}
	}
}

// SubscribeCallback 提供基于回调函数的订阅方式
// 内部使用通道订阅，但向调用者提供回调式 API
// 返回一个清理函数，调用它可以取消订阅
// 注意: 调用方应该在回调函数中处理错误！
func (eb *EventBus[E]) SubscribeCallback(cb func(data E)) func() {
	ch := eb.Subscribe()
	cleanup := func() {
		eb.Unsubscribe(ch)
	}

	go func() {
		defer cleanup()
		for data := range ch {
			cb(data)
		}
	}()
	return cleanup
}

// Publish 发布事件
// 所有订阅的通道都会收到事件数据
func (eb *EventBus[E]) Publish(data E) {
	eb.lk.RLock()
	defer eb.lk.RUnlock()

	go func() {
		for _, ch := range eb.subscribers {
			ch <- data
		}
	}()
}
