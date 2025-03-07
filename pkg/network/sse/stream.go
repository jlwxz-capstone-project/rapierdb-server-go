package sse

import (
	"net/url"
	"sync"
	"sync/atomic"
)

// SseStream 表示一个 SSE 流，用于管理订阅者和事件分发。
// 它负责处理订阅者的注册和注销，将事件广播给所有订阅者，
// 并可选择性地保存事件历史记录以便新订阅者可以重放错过的事件。
type SseStream struct {
	ID              string                                 // 流的唯一标识符
	event           chan *SseEvent                         // 用于接收要广播的事件的通道
	quit            chan struct{}                          // 用于通知流关闭的通道
	quitOnce        sync.Once                              // 确保quit通道只关闭一次
	register        chan *Subscriber                       // 用于注册新订阅者的通道
	deregister      chan *Subscriber                       // 用于注销订阅者的通道
	subscribers     []*Subscriber                          // 当前活跃的订阅者列表
	Eventlog        EventLog                               // 存储历史事件用于重放
	subscriberCount int32                                  // 当前订阅者数量的原子计数器
	AutoReplay      bool                                   // 是否自动重放历史事件给新订阅者
	isAutoStream    bool                                   // 是否自动创建流
	OnSubscribe     func(streamID string, sub *Subscriber) // 订阅者注册时的回调函数
	OnUnsubscribe   func(streamID string, sub *Subscriber) // 订阅者注销时的回调函数
}

func NewStream(id string, buffSize int, replay, isAutoStream bool, onSubscribe, onUnsubscribe func(string, *Subscriber)) *SseStream {
	return &SseStream{
		ID:            id,
		AutoReplay:    replay,
		subscribers:   make([]*Subscriber, 0),
		isAutoStream:  isAutoStream,
		register:      make(chan *Subscriber),
		deregister:    make(chan *Subscriber),
		event:         make(chan *SseEvent, buffSize),
		quit:          make(chan struct{}),
		Eventlog:      make(EventLog, 0),
		OnSubscribe:   onSubscribe,
		OnUnsubscribe: onUnsubscribe,
	}
}

func (str *SseStream) Run() {
	go func(str *SseStream) {
		for {
			select {
			// 添加新订阅者
			case subscriber := <-str.register:
				str.subscribers = append(str.subscribers, subscriber)
				if str.AutoReplay {
					str.Eventlog.Replay(subscriber)
				}

			// 移除已关闭的订阅者
			case subscriber := <-str.deregister:
				i := str.GetSubIndex(subscriber)
				if i != -1 {
					str.RemoveSubscriber(i)
				}

				if str.OnUnsubscribe != nil {
					go str.OnUnsubscribe(str.ID, subscriber)
				}

			// 向订阅者发布事件
			case event := <-str.event:
				if str.AutoReplay {
					str.Eventlog.Add(event)
				}
				for i := range str.subscribers {
					str.subscribers[i].connection <- event
				}

			// 如果服务器关闭则关闭
			case <-str.quit:
				// 移除连接
				str.RemoveAllSubscribers()
				return
			}
		}
	}(str)
}

func (str *SseStream) Close() {
	str.quitOnce.Do(func() {
		close(str.quit)
	})
}

func (str *SseStream) GetSubIndex(sub *Subscriber) int {
	for i := range str.subscribers {
		if str.subscribers[i] == sub {
			return i
		}
	}
	return -1
}

// AddSubscriber 将在流上创建一个新的订阅者
func (str *SseStream) AddSubscriber(eventid int, url *url.URL) *Subscriber {
	atomic.AddInt32(&str.subscriberCount, 1)
	sub := &Subscriber{
		eventid:    eventid,
		quit:       str.deregister,
		connection: make(chan *SseEvent, 64),
		URL:        url,
	}

	if str.isAutoStream {
		sub.removed = make(chan struct{}, 1)
	}

	str.register <- sub

	if str.OnSubscribe != nil {
		go str.OnSubscribe(str.ID, sub)
	}

	return sub
}

func (str *SseStream) RemoveSubscriber(i int) {
	atomic.AddInt32(&str.subscriberCount, -1)
	close(str.subscribers[i].connection)
	if str.subscribers[i].removed != nil {
		str.subscribers[i].removed <- struct{}{}
		close(str.subscribers[i].removed)
	}
	str.subscribers = append(str.subscribers[:i], str.subscribers[i+1:]...)
}

func (str *SseStream) RemoveAllSubscribers() {
	for i := 0; i < len(str.subscribers); i++ {
		close(str.subscribers[i].connection)
		if str.subscribers[i].removed != nil {
			str.subscribers[i].removed <- struct{}{}
			close(str.subscribers[i].removed)
		}
	}
	atomic.StoreInt32(&str.subscriberCount, 0)
	str.subscribers = str.subscribers[:0]
}

func (str *SseStream) GetSubscriberCount() int {
	return int(atomic.LoadInt32(&str.subscriberCount))
}
