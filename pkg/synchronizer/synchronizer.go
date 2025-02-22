package synchronizer

import (
	"bytes"
	"context"
	"log"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/permissions"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/synchronizer/message/v1"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

type SynchronizerConfig struct {
}

type Synchronizer struct {
	storageEngine           *storage.StorageEngine
	storageEngineEvents     *StorageEngineEvents
	channel                 Channel
	cancel                  context.CancelFunc
	config                  SynchronizerConfig
	permission              *permissions.Permission
	collectionSubscriptions map[string]map[string]struct{}
}

type StorageEngineEvents struct {
	CommittedCh  <-chan any
	CanceledCh   <-chan any
	RollbackedCh <-chan any
}

func NewSynchronizer(storageEngine *storage.StorageEngine, channel Channel, config *SynchronizerConfig, permission *permissions.Permission) *Synchronizer {
	// 使用默认配置
	if config == nil {
		config = &SynchronizerConfig{}
	}

	synchronizer := &Synchronizer{
		storageEngine:       storageEngine,
		storageEngineEvents: &StorageEngineEvents{},
		channel:             channel,
		cancel:              nil,
		config:              *config,
		permission:          permission,
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
	s.storageEngineEvents = &StorageEngineEvents{
		CommittedCh:  committedCh,
		CanceledCh:   canceledCh,
		RollbackedCh: rollbackedCh,
	}

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
		buf := bytes.NewBuffer(msg)
		msgType, err := util.ReadVarUint(buf)
		if err != nil {
			log.Println("msgHandler: failed to read message type", err)
			return
		}
	}
	s.channel.SetMsgHandler(msgHandler)

	return nil
}

func (s *Synchronizer) Subscribe(clientId string, collection string) {
	if _, ok := s.collectionSubscriptions[clientId]; !ok {
		s.collectionSubscriptions[clientId] = make(map[string]struct{})
	}
	s.collectionSubscriptions[clientId][collection] = struct{}{}
}

func (s *Synchronizer) Unsubscribe(clientId string, collection string) {
	if subscriptions, ok := s.collectionSubscriptions[clientId]; ok {
		delete(subscriptions, collection)
		// 如果该客户端没有订阅任何集合了,删除该客户端的记录
		if len(subscriptions) == 0 {
			delete(s.collectionSubscriptions, clientId)
		}
	}
}

// handleTransactionCommitted 处理事务提交事件
// 注意，这个方法会捕获所有错误，保证不会因为错误而中断
func (s *Synchronizer) handleTransactionCommitted(event_ any) {
	event, ok := event_.(*storage.TransactionCommittedEvent)
	if !ok {
		log.Println("handleTransactionCommitted: event is not a *storage.TransactionCommittedEvent")
		return
	}
	// 哪个节点提交的事务，就向这个节点发送 AckTransactionMessage
	ackMsg := message.AckTransactionMessageV1{
		TxID: event.Transaction.TxID,
	}
	ackMsgBytes, err := ackMsg.Encode()
	if err != nil {
		log.Println("handleTransactionCommitted: failed to encode ack message", err)
	}
	err = s.channel.Send(event.Committer, ackMsgBytes)
	if err != nil {
		log.Println("handleTransactionCommitted: failed to send ack message", err)
	}
	// 遍历所有节点，对每个节点，检查本次事务改动的文档是否在这个节点订阅的集合中
	// 并且对这个节点可见。如果对一个节点，存在这样的文档，则将这些文档通过
	// PostUpdateMessage 发送给这个节点
	clientIds := s.channel.GetAllConnectedClientIds()
	for _, clientId := range clientIds {
		if cs, ok := s.collectionSubscriptions[clientId]; ok {
			for collection := range cs {

			}
		}
	}
}

func (s *Synchronizer) handleTransactionCanceled(event any) {
	// 哪个节点发送的事务被取消，就向这个节点发送 TransactionFailedMessage
}

func (s *Synchronizer) handleTransactionRollbacked(event any) {
	// 哪个节点发送的事务被回滚，就向这个节点发送 TransactionFailedMessage
}

func (s *Synchronizer) Stop() {
	// 发送取消信号
	if s.cancel != nil {
		s.cancel()
	}

	// 取消所有订阅
	if s.storageEngineEvents != nil {
		s.storageEngine.Unsubscribe(storage.STORAGE_ENGINE_EVENT_TRANSACTION_COMMITTED, s.storageEngineEvents.CommittedCh)
		s.storageEngine.Unsubscribe(storage.STORAGE_ENGINE_EVENT_TRANSACTION_CANCELED, s.storageEngineEvents.CanceledCh)
		s.storageEngine.Unsubscribe(storage.STORAGE_ENGINE_EVENT_TRANSACTION_ROLLBACKED, s.storageEngineEvents.RollbackedCh)
	}
	s.storageEngineEvents = nil

	// 卸载消息处理器
	s.channel.SetMsgHandler(nil)

	// 关闭所有连接
	s.channel.CloseAll()
}
