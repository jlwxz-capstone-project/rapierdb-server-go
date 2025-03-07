package sse

import (
	"strconv"
	"time"
)

// EventLog 保存所有之前的事件
type EventLog []*SseEvent

// Add 将事件添加到事件日志
func (e *EventLog) Add(ev *SseEvent) {
	if !ev.hasContent() {
		return
	}

	ev.ID = []byte(e.currentindex())
	ev.timestamp = time.Now()
	*e = append(*e, ev)
}

// Clear 清除事件日志中的所有事件
func (e *EventLog) Clear() {
	*e = nil
}

// Replay 向订阅者重播事件
func (e *EventLog) Replay(s *Subscriber) {
	for i := 0; i < len(*e); i++ {
		id, _ := strconv.Atoi(string((*e)[i].ID))
		if id >= s.eventid {
			s.connection <- (*e)[i]
		}
	}
}

func (e *EventLog) currentindex() string {
	return strconv.Itoa(len(*e))
}
