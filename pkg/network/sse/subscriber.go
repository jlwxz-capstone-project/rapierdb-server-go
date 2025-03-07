package sse

import "net/url"

// Subscriber 订阅者
type Subscriber struct {
	quit       chan *Subscriber
	connection chan *SseEvent
	removed    chan struct{}
	eventid    int
	URL        *url.URL
}

// Close 将通知流客户端连接已终止
func (s *Subscriber) Close() {
	s.quit <- s
	if s.removed != nil {
		<-s.removed
	}
}

func (s *Subscriber) GetEventChannel() chan *SseEvent {
	return s.connection
}
