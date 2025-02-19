package synchronizer

import (
	"context"
	"log"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/synchronizer/message"
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

	msgHandler := func(clientId string, msg []byte) {
		// handle message
	}
	s.channel.SetMsgHandler(msgHandler)

	return nil
}

// handleTransactionCommitted 处理事务提交事件
// 注意，这个方法会捕获所有错误，保证不会因为错误而中断
func (s *Synchronizer) handleTransactionCommitted(event any) {
	// 哪个节点提交的事务，就向这个节点发送 AckTransactionMessage
	committedEvent, ok := event.(*storage.TransactionCommittedEvent)
	if !ok {
		log.Println("handleTransactionCommitted: event is not a *storage.TransactionCommittedEvent")
		return
	}
	ackMsg := message.AckTransactionMessageV1{
		TxID: committedEvent.Transaction.TxID,
	}
	ackMsgBytes, err := ackMsg.Encode()
	if err != nil {
		log.Println("handleTransactionCommitted: failed to encode ack message", err)
	}
	err = s.channel.Send(committedEvent.Committer, ackMsgBytes)
	if err != nil {
		log.Println("handleTransactionCommitted: failed to send ack message", err)
	}
	// 遍历所有节点，将新的文档通过 PostUpdateMessage 节点
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

	// 卸载消息处理器
	s.channel.SetMsgHandler(nil)

	// 关闭所有连接
	s.channel.CloseAll()
}
