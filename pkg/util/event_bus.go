package util

import (
	"sync"
)

type EventBus struct {
	subscribers map[string][]chan any
	lk          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan any),
	}
}

func (eb *EventBus) Subscribe(topic string) <-chan any {
	eb.lk.Lock()
	defer eb.lk.Unlock()

	ch := make(chan any, 100) // 带缓冲的通道
	eb.subscribers[topic] = append(eb.subscribers[topic], ch)
	return ch
}

func (eb *EventBus) Unsubscribe(topic string, ch <-chan any) {
	eb.lk.Lock()
	defer eb.lk.Unlock()

	if subscribers, found := eb.subscribers[topic]; found {
		for i := range subscribers {
			if subscribers[i] == ch {
				// 关闭通道并移除订阅者
				close(subscribers[i])
				eb.subscribers[topic] = append(subscribers[:i], subscribers[i+1:]...)
				break
			}
		}
	}
}

func (eb *EventBus) Publish(topic string, data any) {
	eb.lk.RLock()
	defer eb.lk.RUnlock()

	if subscribers, found := eb.subscribers[topic]; found {
		tmp := append([]chan any{}, subscribers...)
		go func() {
			for _, ch := range tmp {
				ch <- data
			}
		}()
	}
}
