package synchronizer

import (
	"context"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage"
)

// EventMessage 定义了同步事件的消息格式
type EventMessage struct {
	Type      string    `json:"type"`      // 事件类型
	Timestamp time.Time `json:"timestamp"` // 事件发生时间
	TxID      string    `json:"tx_id"`     // 事务ID
	Data      any       `json:"data"`      // 事件相关数据
}

type SynchronizerConfig struct {
}

type Synchronizer struct {
	storageEngine *storage.StorageEngine
	channel       Channel
	cancel        context.CancelFunc
	config        SynchronizerConfig
	subscriptions []<-chan any
}

func NewSynchronizer(storageEngine *storage.StorageEngine, channel Channel, config *SynchronizerConfig) *Synchronizer {
	// 使用默认配置
	if config == nil {
		config = &SynchronizerConfig{}
	}

	synchronizer := &Synchronizer{
		storageEngine: storageEngine,
		channel:       channel,
		cancel:        nil,
		config:        *config,
		subscriptions: []<-chan any{},
	}
	return synchronizer
}

func (s *Synchronizer) Start() error {
	// 订阅存储引擎事件
	committedCh := s.storageEngine.Subscribe(storage.STORAGE_ENGINE_EVENT_TRANSACTION_COMMITTED)
	canceledCh := s.storageEngine.Subscribe(storage.STORAGE_ENGINE_EVENT_TRANSACTION_CANCELED)
	rollbackedCh := s.storageEngine.Subscribe(storage.STORAGE_ENGINE_EVENT_TRANSACTION_ROLLBACKED)

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	// 保存订阅通道以便后续清理
	s.subscriptions = []<-chan any{committedCh, canceledCh, rollbackedCh}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-committedCh:
				s.handleTransactionCommitted(event)
			case event := <-canceledCh:
				s.handleTransactionCanceled(event)
			case event := <-rollbackedCh:
				s.handleTransactionRollbacked(event)
			}
		}
	}()

	return nil
}

func (s *Synchronizer) handleTransactionCommitted(event any) {
	committedEvent, ok := event.(*storage.TransactionCommittedEvent)
	if !ok {
		return // TODO 处理错误
	}
}

func (s *Synchronizer) handleTransactionCanceled(event any) {

}

func (s *Synchronizer) handleTransactionRollbacked(event any) {

}

func (s *Synchronizer) Stop() {
	// 发送取消信号
	if s.cancel != nil {
		s.cancel()
	}

	// 取消所有订阅
	if s.subscriptions != nil {
		for _, ch := range s.subscriptions {
			s.storageEngine.Unsubscribe(storage.STORAGE_ENGINE_EVENT_TRANSACTION_COMMITTED, ch)
			s.storageEngine.Unsubscribe(storage.STORAGE_ENGINE_EVENT_TRANSACTION_CANCELED, ch)
			s.storageEngine.Unsubscribe(storage.STORAGE_ENGINE_EVENT_TRANSACTION_ROLLBACKED, ch)
		}
	}
	s.subscriptions = nil
}
