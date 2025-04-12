package synchronizer

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"sync"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/message/v1"
	network_server "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/server"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

var (
	ErrInvalidStorageEvent = errors.New("invalid storage event")
)

// 同步器状态常量
type SynchronizerStatus string

const (
	SynchronizerStatusStopped  SynchronizerStatus = "stopped"
	SynchronizerStatusStarting SynchronizerStatus = "starting"
	SynchronizerStatusRunning  SynchronizerStatus = "running"
	SynchronizerStatusStopping SynchronizerStatus = "stopping"
)

type SynchronizerConfig struct {
}

type Synchronizer struct {
	storageEngine       *storage_engine.StorageEngine
	storageEngineEvents *StorageEngineEvents
	queryExecutor       *query.QueryExecutor
	channel             network_server.Channel
	config              SynchronizerConfig
	permission          *query.Permissions
	ctx                 context.Context
	cancel              context.CancelFunc
	queryManager        *QueryManager

	// 状态相关字段
	status     SynchronizerStatus
	statusLock sync.RWMutex
	eventBus   *util.EventBus[SynchronizerStatus]
}

type StorageEngineEvents struct {
	CommittedCh  <-chan *storage_engine.TransactionCommittedEvent
	CanceledCh   <-chan *storage_engine.TransactionCanceledEvent
	RollbackedCh <-chan *storage_engine.TransactionRollbackedEvent
}

func NewSynchronizer(storageEngine *storage_engine.StorageEngine, channel network_server.Channel, config *SynchronizerConfig) *Synchronizer {
	return NewSynchronizerWithContext(context.Background(), storageEngine, channel, config)
}

func NewSynchronizerWithContext(ctx context.Context, storageEngine *storage_engine.StorageEngine, channel network_server.Channel, config *SynchronizerConfig) *Synchronizer {
	// 使用默认配置
	if config == nil {
		config = &SynchronizerConfig{}
	}

	log.Debugf("NewSynchronizer: 正在创建同步器")
	permission, err := query.NewPermissionFromJs(storageEngine.GetPermissionsJs())
	if err != nil {
		log.Errorf("NewSynchronizer: 创建权限失败: %v", err)
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(ctx)

	queryExecutor := query.NewQueryExecutor(storageEngine)
	synchronizer := &Synchronizer{
		storageEngine:       storageEngine,
		storageEngineEvents: &StorageEngineEvents{},
		queryExecutor:       queryExecutor,
		channel:             channel,
		config:              *config,
		permission:          permission,
		queryManager:        NewQueryManager(queryExecutor, permission),
		ctx:                 ctx,
		cancel:              cancel,
		status:              SynchronizerStatusStopped,
		statusLock:          sync.RWMutex{},
		eventBus:            util.NewEventBus[SynchronizerStatus](),
	}
	log.Debugf("NewSynchronizer: 同步器创建成功")
	return synchronizer
}

// GetStatus 获取同步器当前状态
func (s *Synchronizer) GetStatus() SynchronizerStatus {
	s.statusLock.RLock()
	defer s.statusLock.RUnlock()
	return s.status
}

// setStatus 设置同步器状态（内部使用）
func (s *Synchronizer) setStatus(status SynchronizerStatus) {
	s.statusLock.Lock()
	oldStatus := s.status
	s.status = status
	s.statusLock.Unlock()

	// 只有状态发生变化时才发布事件
	if oldStatus != status {
		s.eventBus.Publish(status)
	}
}

// SubscribeStatusChange 订阅状态变更事件
func (s *Synchronizer) SubscribeStatusChange() <-chan SynchronizerStatus {
	return s.eventBus.Subscribe()
}

// UnsubscribeStatusChange 取消订阅状态变更事件
func (s *Synchronizer) UnsubscribeStatusChange(ch <-chan SynchronizerStatus) {
	s.eventBus.Unsubscribe(ch)
}

// WaitForStatus 等待同步器达到指定状态
func (s *Synchronizer) WaitForStatus(targetStatus SynchronizerStatus) <-chan struct{} {
	statusCh := s.SubscribeStatusChange()
	cleanup := func() {
		s.UnsubscribeStatusChange(statusCh)
	}
	return util.WaitForStatus(s.GetStatus, targetStatus, statusCh, cleanup, 0)
}

func (s *Synchronizer) Start() error {
	log.Debugf("Synchronizer.Start: 正在启动同步器")
	s.setStatus(SynchronizerStatusStarting)

	// 订阅存储引擎事件
	committedCh := s.storageEngine.SubscribeCommitted()
	canceledCh := s.storageEngine.SubscribeCanceled()
	rollbackedCh := s.storageEngine.SubscribeRollbacked()
	log.Debugf("Synchronizer.Start: 已订阅存储引擎事件")

	// 保存订阅通道以便后续清理
	s.storageEngineEvents = &StorageEngineEvents{
		CommittedCh:  committedCh,
		CanceledCh:   canceledCh,
		RollbackedCh: rollbackedCh,
	}

	go func() {
		log.Debugf("Synchronizer: 事件处理协程已启动")
		for {
			select {
			case <-s.ctx.Done():
				log.Debugf("Synchronizer: 收到取消信号，事件处理协程退出")
				return
			case event := <-committedCh:
				log.Debugf("Synchronizer: 收到事务提交事件")
				s.handleTransactionCommitted(event)
			case event := <-canceledCh:
				log.Debugf("Synchronizer: 收到事务取消事件")
				s.handleTransactionCanceled(event)
			case event := <-rollbackedCh:
				log.Debugf("Synchronizer: 收到事务回滚事件")
				s.handleTransactionRollbacked(event)
			}
		}
	}()

	// 负责处理收到的客户端消息
	msgHandler := func(clientId string, msgBytes []byte) {
		log.Debugf("Synchronizer.msgHandler: 收到来自客户端 %s 的消息，长度 %d 字节", clientId, len(msgBytes))
		buf := bytes.NewBuffer(msgBytes)
		msg, err := message.DecodeMessage(buf)
		if err != nil {
			log.Errorf("msgHandler: 解码消息失败: %#+v", err)
			return
		}
		switch msg := msg.(type) {
		case *message.PostTransactionMessageV1:
			// 提交事务到存储引擎
			// 为事务设置提交者ID
			log.Debugf("msgHandler: 收到 %s 来自 %s", msg.DebugPrint(), clientId)
			msg.Transaction.Committer = clientId
			log.Debugf("msgHandler: 正在提交事务 %s 到存储引擎", msg.Transaction.TxID)
			err = s.storageEngine.Commit(msg.Transaction)
			if err != nil {
				log.Errorf("msgHandler: 提交事务失败: %v", err)
				return
			}
			log.Debugf("msgHandler: 事务 %s 提交成功", msg.Transaction.TxID)

		case *message.SubscriptionUpdateMessageV1:
			log.Debugf("msgHandler: 收到 %s 来自 %s", msg.DebugPrint(), clientId)
			// 处理移除的订阅
			for _, q := range msg.Removed {
				log.Debugf("msgHandler: 客户端 %s 取消订阅查询 %s", clientId, q.DebugPrint())
				err := s.queryManager.RemoveSubscriptedQuery(clientId, q)
				if err != nil {
					log.Errorf("msgHandler: 移除订阅失败: %v", err)
				}
			}

			// 处理添加的订阅
			vqm := message.NewVersionQueryMessageV1()
			for _, q := range msg.Added {
				log.Debugf("msgHandler: 客户端 %s 订阅查询 %s", clientId, q.DebugPrint())
				err := s.queryManager.SubscribeNewQuery(clientId, q)
				if err != nil {
					log.Errorf("msgHandler: 添加订阅失败: %v", err)
				}

				var queryCollection string
				switch q := q.(type) {
				case *query.FindManyQuery:
					queryCollection = q.Collection
				case *query.FindOneQuery:
					queryCollection = q.Collection
				}

				// TODO: 按理说这里应该只将查询结果中包含的文档放入 vqm 中
				// 但简单起见，这里将查询对应集合里的所有文档都放入 vqm 中
				allCollections := vqm.GetAllCollections()
				if !slices.Contains(allCollections, queryCollection) {
					docs, err := s.storageEngine.LoadAllDocsInCollection(queryCollection, true)
					if err != nil {
						log.Errorf("msgHandler: 获取集合 %s 所有文档失败: %v", queryCollection, err)
						continue
					}
					for docId := range docs {
						vqm.AddDoc(queryCollection, docId)
					}
				}
			}

			vqmBytes, err := vqm.Encode()
			if err != nil {
				log.Errorf("msgHandler: 编码版本查询消息失败: %v", err)
				return
			}

			log.Debugf("msgHandler: 向客户端 %s 发送版本查询消息 %s", clientId, vqm.DebugPrint())
			s.channel.Send(clientId, vqmBytes)

		case *message.VersionQueryRespMessageV1:
			log.Debugf("msgHandler: 收到 %s 来自 %s", msg.DebugPrint(), clientId)
			toUpsert := make(map[string][]byte)
			toDelete := make([]string, 0)
			for docKey, vvBytes := range msg.Responses {
				docKeyBytes := util.String2Bytes(docKey)
				collection := storage_engine.GetCollectionNameFromKey(docKeyBytes)
				docId := storage_engine.GetDocIdFromKey(docKeyBytes)
				if vvBytes == nil || len(vvBytes) == 0 {
					doc, err := s.storageEngine.LoadDoc(collection, docId)
					if err != nil {
						log.Errorf("msgHandler: 加载文档 %s/%s 失败: %v", collection, docId, err)
						continue
					}
					docBytes := doc.ExportSnapshot().Bytes()
					toUpsert[docKey] = docBytes
				} else {
					doc, err := s.storageEngine.LoadDoc(collection, docId)
					if err != nil {
						log.Errorf("msgHandler: 加载文档 %s/%s 失败: %v", collection, docId, err)
						continue
					}
					vv := loro.NewVvFromBytes(loro.NewRustBytesVec(vvBytes))
					updateBytesVec := doc.ExportUpdatesFrom(vv)
					updateBytes := updateBytesVec.Bytes()
					toUpsert[docKey] = updateBytes
				}
			}
			if len(toUpsert) > 0 {
				syncMsg := message.PostDocMessageV1{
					Upsert: toUpsert,
					Delete: toDelete,
				}
				log.Debugf("msgHandler: 向客户端 %s 发送同步消息 %s", clientId, syncMsg.DebugPrint())
				syncMsgBytes, err := syncMsg.Encode()
				if err != nil {
					log.Errorf("msgHandler: 编码同步消息失败: %v", err)
					return
				}
				s.channel.Send(clientId, syncMsgBytes)
			}
		}
	}
	s.channel.SetMsgHandler(msgHandler)

	// 设置状态为运行中
	s.setStatus(SynchronizerStatusRunning)
	log.Debugf("Synchronizer.Start: 同步器启动完成")

	return nil
}

// handleTransactionCommitted 处理事务提交事件
// 注意，这个方法会捕获所有错误，保证不会因为错误而中断
func (s *Synchronizer) handleTransactionCommitted(event_ any) {
	log.Debugf("handleTransactionCommitted: 开始处理事务提交事件")
	event, ok := event_.(*storage_engine.TransactionCommittedEvent)
	if !ok {
		log.Errorf("handleTransactionCommitted: 事件不是 *storage.TransactionCommittedEvent 类型")
		return
	}
	log.Debugf("handleTransactionCommitted: 处理事务 %s 提交事件，提交者 %s", event.Transaction.TxID, event.Committer)

	// 哪个节点提交的事务，就向这个节点发送 AckTransactionMessage
	ackMsg := message.AckTransactionMessageV1{
		TxID: event.Transaction.TxID,
	}
	ackMsgBytes, err := ackMsg.Encode()
	if err != nil {
		log.Errorf("handleTransactionCommitted: 编码确认消息失败", err)
	}
	log.Debugf("handleTransactionCommitted: 向提交者 %s 发送确认消息", event.Committer)
	err = s.channel.Send(event.Committer, ackMsgBytes)
	if err != nil {
		log.Errorf("handleTransactionCommitted: 发送确认消息失败", err)
	}

	// 遍历所有节点，对每个节点，检查本次事务改动的文档是否在这个节点订阅的集合中
	// 并且对这个节点可见。如果对一个节点，存在这样的文档，则将这些文档通过
	// PostUpdateMessage 发送给这个节点
	clientIds := s.channel.GetAllConnectedClientIds()
	log.Debugf("handleTransactionCommitted: 当前连接的客户端数量: %d", len(clientIds))
	for _, clientId := range clientIds {
		// 跳过提交事务的客户端，因为它已经收到了确认消息
		if clientId == event.Committer {
			log.Debugf("handleTransactionCommitted: 跳过提交者 %s", clientId)
			continue
		}

		// 获取客户端订阅的集合
		queries := s.activeSet.GetSubscriptions(clientId)
		// TODO: 这里之后应该采用 Event Reduce 算法
		// 现在暂时找出订阅的查询涉及的所有集合
		collections := make(map[string]struct{})
		for _, q := range queries {
			switch q := q.(type) {
			case *query.FindManyQuery:
				collections[q.Collection] = struct{}{}
			case *query.FindOneQuery:
				collections[q.Collection] = struct{}{}
			}
		}

		// 遍历事务中的所有操作，查找修改的文档
		upsertDocs := make(map[string][]byte)
		deleteDocs := make([]string, 0)

		for _, op := range event.Transaction.Operations {
			switch operation := op.(type) {
			case *storage_engine.InsertOp:
				// 检查该集合是否被订阅
				if _, subscribed := collections[operation.Collection]; !subscribed {
					log.Debugf("handleTransactionCommitted: 客户端 %s 未订阅集合 %s，跳过", clientId, operation.Collection)
					continue
				}
				log.Debugf("handleTransactionCommitted: 处理插入操作 %s/%s", operation.Collection, operation.DocID)

				doc, err := s.storageEngine.LoadDoc(operation.Collection, operation.DocID)
				if err != nil {
					log.Errorf("handleTransactionCommitted: 加载文档 %s/%s 失败: %v",
						operation.Collection, operation.DocID, err)
					continue
				}

				// 检查客户端是否有权限查看此文档
				canViewParams := query.CanViewParams{
					Collection: operation.Collection,
					DocId:      operation.DocID,
					Doc:        doc,
					ClientId:   clientId,
				}
				canView := s.permission.CanView(canViewParams)
				if !canView {
					log.Debugf("handleTransactionCommitted: 客户端 %s 无权查看文档 %s/%s", clientId, operation.Collection, operation.DocID)
					continue
				}
				log.Debugf("handleTransactionCommitted: 客户端 %s 可以查看文档 %s/%s", clientId, operation.Collection, operation.DocID)

				// 文档可见，添加到upsert列表
				docKeyBytes, err := storage_engine.CalcDocKey(operation.Collection, operation.DocID)
				if err != nil {
					log.Errorf("handleTransactionCommitted: 计算文档键 %s/%s 失败: %v",
						operation.Collection, operation.DocID, err)
					continue
				}
				docKey := util.Bytes2String(docKeyBytes)
				upsertDocs[docKey] = operation.Snapshot
				log.Debugf("handleTransactionCommitted: 添加文档 %s/%s 到 upsert 列表", operation.Collection, operation.DocID)

			case *storage_engine.UpdateOp:
				// 检查该集合是否被订阅
				if _, subscribed := collections[operation.Collection]; !subscribed {
					log.Debugf("handleTransactionCommitted: 客户端 %s 未订阅集合 %s，跳过", clientId, operation.Collection)
					continue
				}
				log.Debugf("handleTransactionCommitted: 处理更新操作 %s/%s", operation.Collection, operation.DocID)

				doc, err := s.storageEngine.LoadDoc(operation.Collection, operation.DocID)
				if err != nil {
					log.Errorf("handleTransactionCommitted: 加载文档 %s/%s 失败: %v",
						operation.Collection, operation.DocID, err)
					continue
				}

				// 检查客户端是否有权限查看此文档
				canViewParams := query.CanViewParams{
					Collection: operation.Collection,
					DocId:      operation.DocID,
					Doc:        doc,
					ClientId:   clientId,
				}
				canView := s.permission.CanView(canViewParams)
				if !canView {
					log.Debugf("handleTransactionCommitted: 客户端 %s 无权查看文档 %s/%s", clientId, operation.Collection, operation.DocID)
					continue
				}
				log.Debugf("handleTransactionCommitted: 客户端 %s 可以查看文档 %s/%s", clientId, operation.Collection, operation.DocID)

				// 文档可见，添加到upsert列表
				docKeyBytes, err := storage_engine.CalcDocKey(operation.Collection, operation.DocID)
				if err != nil {
					log.Errorf("handleTransactionCommitted: 计算文档键 %s/%s 失败: %v",
						operation.Collection, operation.DocID, err)
					continue
				}
				docKey := util.Bytes2String(docKeyBytes)
				snapshot := doc.ExportSnapshot()
				upsertDocs[docKey] = snapshot.Bytes()
				log.Debugf("handleTransactionCommitted: 添加文档 %s/%s 到 upsert 列表", operation.Collection, operation.DocID)

			case *storage_engine.DeleteOp:
				// 检查该集合是否被订阅
				if _, subscribed := collections[operation.Collection]; !subscribed {
					log.Debugf("handleTransactionCommitted: 客户端 %s 未订阅集合 %s，跳过", clientId, operation.Collection)
					continue
				}
				log.Debugf("handleTransactionCommitted: 处理删除操作 %s/%s", operation.Collection, operation.DocID)

				// 被删除的文档添加到 delete 列表
				docKeyBytes, err := storage_engine.CalcDocKey(operation.Collection, operation.DocID)
				if err != nil {
					log.Errorf("handleTransactionCommitted: 计算文档键 %s/%s 失败: %v",
						operation.Collection, operation.DocID, err)
					continue
				}
				docKey := util.Bytes2String(docKeyBytes)
				deleteDocs = append(deleteDocs, docKey)
				log.Debugf("handleTransactionCommitted: 添加文档 %s/%s 到 delete 列表", operation.Collection, operation.DocID)
			}
		}

		// 如果有需要同步的文档，则发送同步消息
		if len(upsertDocs) > 0 || len(deleteDocs) > 0 {
			log.Debugf("handleTransactionCommitted: 向客户端 %s 发送同步消息，upsert: %d, delete: %d",
				clientId, len(upsertDocs), len(deleteDocs))
			syncMsg := message.PostDocMessageV1{
				Upsert: upsertDocs,
				Delete: deleteDocs,
			}

			syncMsgBytes, err := syncMsg.Encode()
			if err != nil {
				log.Errorf("handleTransactionCommitted: 为客户端 %s 编码同步消息失败: %v",
					clientId, err)
				continue
			}

			err = s.channel.Send(clientId, syncMsgBytes)
			if err != nil {
				log.Errorf("handleTransactionCommitted: 向客户端 %s 发送同步消息失败: %v",
					clientId, err)
			} else {
				log.Debugf("handleTransactionCommitted: 成功向客户端 %s 发送同步消息", clientId)
			}
		} else {
			log.Debugf("handleTransactionCommitted: 客户端 %s 没有需要同步的文档", clientId)
		}
	}
	log.Debugf("handleTransactionCommitted: 事务 %s 处理完成", event.Transaction.TxID)
}

func (s *Synchronizer) handleTransactionCanceled(event any) {
	log.Debugf("handleTransactionCanceled: 开始处理事务取消事件")
	// 哪个节点发送的事务被取消，就向这个节点发送 TransactionFailedMessage
	canceledEvent, ok := event.(*storage_engine.TransactionCanceledEvent)
	if !ok {
		log.Errorf("handleTransactionCanceled: 事件不是 *storage.TransactionCanceledEvent 类型")
		return
	}
	log.Debugf("handleTransactionCanceled: 处理事务 %s 取消事件，提交者 %s，原因: %s",
		canceledEvent.Transaction.TxID, canceledEvent.Committer, canceledEvent.Reason)

	failedMsg := message.TransactionFailedMessageV1{
		TxID:   canceledEvent.Transaction.TxID,
		Reason: canceledEvent.Reason,
	}

	failedMsgBytes, err := failedMsg.Encode()
	if err != nil {
		log.Errorf("handleTransactionCanceled: 编码失败消息失败", err)
		return
	}

	log.Debugf("handleTransactionCanceled: 向提交者 %s 发送失败消息", canceledEvent.Committer)
	err = s.channel.Send(canceledEvent.Committer, failedMsgBytes)
	if err != nil {
		log.Errorf("handleTransactionCanceled: 发送失败消息失败", err)
	} else {
		log.Debugf("handleTransactionCanceled: 成功发送失败消息")
	}
}

func (s *Synchronizer) handleTransactionRollbacked(event any) {
	log.Debugf("handleTransactionRollbacked: 开始处理事务回滚事件")
	// 哪个节点发送的事务被回滚，就向这个节点发送 TransactionFailedMessage
	rollbackedEvent, ok := event.(*storage_engine.TransactionRollbackedEvent)
	if !ok {
		log.Errorf("handleTransactionRollbacked: 事件不是 *storage.TransactionRollbackedEvent 类型")
		return
	}
	log.Debugf("handleTransactionRollbacked: 处理事务 %s 回滚事件，提交者 %s，原因: %s",
		rollbackedEvent.Transaction.TxID, rollbackedEvent.Committer, rollbackedEvent.Reason)

	failedMsg := message.TransactionFailedMessageV1{
		TxID:   rollbackedEvent.Transaction.TxID,
		Reason: rollbackedEvent.Reason,
	}

	failedMsgBytes, err := failedMsg.Encode()
	if err != nil {
		log.Errorf("handleTransactionRollbacked: 编码失败消息失败", err)
		return
	}

	log.Debugf("handleTransactionRollbacked: 向提交者 %s 发送失败消息", rollbackedEvent.Committer)
	err = s.channel.Send(rollbackedEvent.Committer, failedMsgBytes)
	if err != nil {
		log.Errorf("handleTransactionRollbacked: 发送失败消息失败", err)
	} else {
		log.Debugf("handleTransactionRollbacked: 成功发送失败消息")
	}
}

func (s *Synchronizer) Stop() {
	log.Debugf("Synchronizer.Stop: 正在停止同步器")
	s.setStatus(SynchronizerStatusStopping)

	// 发送取消信号
	if s.cancel != nil {
		log.Debugf("Synchronizer.Stop: 发送取消信号")
		s.cancel()
	}

	// 取消所有订阅
	if s.storageEngineEvents != nil {
		log.Debugf("Synchronizer.Stop: 取消存储引擎事件订阅")
		s.storageEngine.UnsubscribeCommitted(s.storageEngineEvents.CommittedCh)
		s.storageEngine.UnsubscribeCanceled(s.storageEngineEvents.CanceledCh)
		s.storageEngine.UnsubscribeRollbacked(s.storageEngineEvents.RollbackedCh)
	}
	s.storageEngineEvents = nil

	// 卸载消息处理器
	log.Debugf("Synchronizer.Stop: 卸载消息处理器")
	s.channel.SetMsgHandler(nil)

	// 关闭所有连接
	log.Debugf("Synchronizer.Stop: 关闭所有连接")
	s.channel.CloseAll()

	s.setStatus(SynchronizerStatusStopped)
	log.Debugf("Synchronizer.Stop: 同步器已停止")
}

func (s *Synchronizer) RemoveSubscription(clientId string, q query.Query) error {
	queryHash, err := query.StableStringify(q)
	if err != nil {
		return err
	}

	if clientMap, ok := s.activeSet[clientId]; ok {
		delete(clientMap, queryHash)
	}
	return nil
}
