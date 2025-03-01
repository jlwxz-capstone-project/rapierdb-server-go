package db

import (
	"bytes"
	"context"
	"errors"
	"log"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/message/v1"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

var (
	ErrInvalidStorageEvent = errors.New("invalid storage event")
)

// 消息类型
const (
	MSG_TYPE_POST_TRANSACTION_V1    uint64 = 1
	MSG_TYPE_SUBSCRIPTION_UPDATE_V1 uint64 = 2
	MSG_TYPE_SYNC_V1                uint64 = 3
	MSG_TYPE_ACK_TRANSACTION_V1     uint64 = 4
	MSG_TYPE_TRANSACTION_FAILED_V1  uint64 = 5
)

type SynchronizerConfig struct {
}

type Synchronizer struct {
	storageEngine           *storage_engine.StorageEngine
	storageEngineEvents     *StorageEngineEvents
	channel                 network.Channel
	cancel                  context.CancelFunc
	config                  SynchronizerConfig
	permission              *query.Permissions
	collectionSubscriptions map[string]map[string]struct{}
}

type StorageEngineEvents struct {
	CommittedCh  <-chan any
	CanceledCh   <-chan any
	RollbackedCh <-chan any
}

func NewSynchronizer(storageEngine *storage_engine.StorageEngine, channel network.Channel, config *SynchronizerConfig) *Synchronizer {
	// 使用默认配置
	if config == nil {
		config = &SynchronizerConfig{}
	}

	permission, err := query.NewPermissionFromJs(storageEngine.GetPermissionsJs())
	if err != nil {
		log.Fatalf("NewSynchronizer: failed to create permission: %v", err)
	}

	synchronizer := &Synchronizer{
		storageEngine:           storageEngine,
		storageEngineEvents:     &StorageEngineEvents{},
		channel:                 channel,
		cancel:                  nil,
		config:                  *config,
		permission:              permission,
		collectionSubscriptions: make(map[string]map[string]struct{}),
	}
	return synchronizer
}

func (s *Synchronizer) Start() error {
	// 订阅存储引擎事件
	committedCh := s.storageEngine.Subscribe(storage_engine.STORAGE_ENGINE_EVENT_TRANSACTION_COMMITTED)
	canceledCh := s.storageEngine.Subscribe(storage_engine.STORAGE_ENGINE_EVENT_TRANSACTION_CANCELED)
	rollbackedCh := s.storageEngine.Subscribe(storage_engine.STORAGE_ENGINE_EVENT_TRANSACTION_ROLLBACKED)

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

	// 负责处理收到的客户端消息
	msgHandler := func(clientId string, msg []byte) {
		buf := bytes.NewBuffer(msg)
		// 先读取消息类型
		msgType, err := util.ReadVarUint(buf)
		if err != nil {
			log.Println("msgHandler: failed to read message type", err)
			return
		}
		switch msgType {
		case MSG_TYPE_POST_TRANSACTION_V1: // PostTransactionMessageV1
			// 解码事务消息
			msg, err := message.DecodePostTransactionMessageV1(buf.Bytes())
			if err != nil {
				log.Println("msgHandler: failed to decode PostTransactionMessageV1", err)
				return
			}

			// 提交事务到存储引擎
			// 为事务设置提交者ID
			msg.Transaction.Committer = clientId
			err = s.storageEngine.Commit(msg.Transaction)
			if err != nil {
				log.Println("msgHandler: failed to commit transaction", err)
				return
			}

		case MSG_TYPE_SUBSCRIPTION_UPDATE_V1: // SubscriptionUpdateMessageV1
			// 解码订阅更新消息
			msg, err := message.DecodeSubscriptionUpdateMessageV1(buf)
			if err != nil {
				log.Println("msgHandler: failed to decode SubscriptionUpdateMessageV1", err)
				return
			}

			// 处理添加的订阅
			for _, collection := range msg.Added {
				s.Subscribe(clientId, collection)
			}

			// 处理移除的订阅
			for _, collection := range msg.Removed {
				s.Unsubscribe(clientId, collection)
			}
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
	event, ok := event_.(*storage_engine.TransactionCommittedEvent)
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
		// 跳过提交事务的客户端，因为它已经收到了确认消息
		if clientId == event.Committer {
			continue
		}

		// 获取客户端订阅的集合
		subscriptions, ok := s.collectionSubscriptions[clientId]
		if !ok || len(subscriptions) == 0 {
			continue // 该客户端没有订阅任何集合
		}

		// 遍历事务中的所有操作，查找修改的文档
		upsertDocs := make(map[string][]byte)
		deleteDocs := make([]string, 0)

		for _, op := range event.Transaction.Operations {
			switch operation := op.(type) {
			case storage_engine.InsertOp:
				// 检查该集合是否被订阅
				if _, subscribed := subscriptions[operation.Collection]; !subscribed {
					continue
				}

				doc, err := s.storageEngine.LoadDoc(operation.Collection, operation.DocID)
				if err != nil {
					log.Printf("handleTransactionCommitted: failed to load doc %s/%s: %v",
						operation.Collection, operation.DocID, err)
					continue
				}

				// 检查客户端是否有权限查看此文档
				canView := s.permission.CanView(operation.Collection, operation.DocID, doc, clientId)
				if !canView {
					continue
				}

				// 文档可见，添加到upsert列表
				docKeyBytes, err := storage_engine.CalcDocKey(operation.Collection, operation.DocID)
				if err != nil {
					log.Printf("handleTransactionCommitted: failed to calc doc key for %s/%s: %v",
						operation.Collection, operation.DocID, err)
					continue
				}
				docKey := util.Bytes2String(docKeyBytes)
				upsertDocs[docKey] = operation.Snapshot

			case storage_engine.UpdateOp:
				// 检查该集合是否被订阅
				if _, subscribed := subscriptions[operation.Collection]; !subscribed {
					continue
				}

				doc, err := s.storageEngine.LoadDoc(operation.Collection, operation.DocID)
				if err != nil {
					log.Printf("handleTransactionCommitted: failed to load doc %s/%s: %v",
						operation.Collection, operation.DocID, err)
					continue
				}

				// 检查客户端是否有权限查看此文档
				canView := s.permission.CanView(operation.Collection, operation.DocID, doc, clientId)
				if !canView {
					continue
				}

				// 文档可见，添加到upsert列表
				docKeyBytes, err := storage_engine.CalcDocKey(operation.Collection, operation.DocID)
				if err != nil {
					log.Printf("handleTransactionCommitted: failed to calc doc key for %s/%s: %v",
						operation.Collection, operation.DocID, err)
					continue
				}
				docKey := util.Bytes2String(docKeyBytes)
				snapshot := doc.ExportSnapshot()
				upsertDocs[docKey] = snapshot.Bytes()

			case storage_engine.DeleteOp:
				// 检查该集合是否被订阅
				if _, subscribed := subscriptions[operation.Collection]; !subscribed {
					continue
				}

				// 被删除的文档添加到 delete 列表
				docKeyBytes, err := storage_engine.CalcDocKey(operation.Collection, operation.DocID)
				if err != nil {
					log.Printf("handleTransactionCommitted: failed to calc doc key for %s/%s: %v",
						operation.Collection, operation.DocID, err)
					continue
				}
				docKey := util.Bytes2String(docKeyBytes)
				deleteDocs = append(deleteDocs, docKey)
			}
		}

		// 如果有需要同步的文档，则发送同步消息
		if len(upsertDocs) > 0 || len(deleteDocs) > 0 {
			syncMsg := message.SyncMessageV1{
				Upsert: upsertDocs,
				Delete: deleteDocs,
			}

			syncMsgBytes, err := syncMsg.Encode()
			if err != nil {
				log.Printf("handleTransactionCommitted: failed to encode sync message for client %s: %v",
					clientId, err)
				continue
			}

			err = s.channel.Send(clientId, syncMsgBytes)
			if err != nil {
				log.Printf("handleTransactionCommitted: failed to send sync message to client %s: %v",
					clientId, err)
			}
		}
	}
}

func (s *Synchronizer) handleTransactionCanceled(event any) {
	// 哪个节点发送的事务被取消，就向这个节点发送 TransactionFailedMessage
	canceledEvent, ok := event.(*storage_engine.TransactionCanceledEvent)
	if !ok {
		log.Println("handleTransactionCanceled: event is not a *storage.TransactionCanceledEvent")
		return
	}

	failedMsg := message.TransactionFailedMessageV1{
		TxID:   canceledEvent.Transaction.TxID,
		Reason: canceledEvent.Reason,
	}

	failedMsgBytes, err := failedMsg.Encode()
	if err != nil {
		log.Println("handleTransactionCanceled: failed to encode failed message", err)
		return
	}

	err = s.channel.Send(canceledEvent.Committer, failedMsgBytes)
	if err != nil {
		log.Println("handleTransactionCanceled: failed to send failed message", err)
	}
}

func (s *Synchronizer) handleTransactionRollbacked(event any) {
	// 哪个节点发送的事务被回滚，就向这个节点发送 TransactionFailedMessage
	rollbackedEvent, ok := event.(*storage_engine.TransactionRollbackedEvent)
	if !ok {
		log.Println("handleTransactionRollbacked: event is not a *storage.TransactionRollbackedEvent")
		return
	}

	failedMsg := message.TransactionFailedMessageV1{
		TxID:   rollbackedEvent.Transaction.TxID,
		Reason: rollbackedEvent.Reason,
	}

	failedMsgBytes, err := failedMsg.Encode()
	if err != nil {
		log.Println("handleTransactionRollbacked: failed to encode failed message", err)
		return
	}

	err = s.channel.Send(rollbackedEvent.Committer, failedMsgBytes)
	if err != nil {
		log.Println("handleTransactionRollbacked: failed to send failed message", err)
	}
}

func (s *Synchronizer) Stop() {
	// 发送取消信号
	if s.cancel != nil {
		s.cancel()
	}

	// 取消所有订阅
	if s.storageEngineEvents != nil {
		s.storageEngine.Unsubscribe(storage_engine.STORAGE_ENGINE_EVENT_TRANSACTION_COMMITTED, s.storageEngineEvents.CommittedCh)
		s.storageEngine.Unsubscribe(storage_engine.STORAGE_ENGINE_EVENT_TRANSACTION_CANCELED, s.storageEngineEvents.CanceledCh)
		s.storageEngine.Unsubscribe(storage_engine.STORAGE_ENGINE_EVENT_TRANSACTION_ROLLBACKED, s.storageEngineEvents.RollbackedCh)
	}
	s.storageEngineEvents = nil

	// 卸载消息处理器
	s.channel.SetMsgHandler(nil)

	// 关闭所有连接
	s.channel.CloseAll()
}
