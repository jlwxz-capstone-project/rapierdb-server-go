场景 1：客户端数据库为空，并创建一个新查询 q

我的设计：

- 客户端发送一个 `SubscriptionUpdateMessage` 通知服务端自己希望订阅这个新查询 q；
- 服务端会维护每个客户端订阅的所有查询（可以理解为 `Map<ClientId, Set<Query>>`），因此在收到客户端发的 `SubscriptionUpdateMessage` 后，将新查询 q 加入对应客户端的查询集合；
- 服务端执行查询 q，得到结果集 res（res 是一个文档集合）；
- 服务端使用 `canView` 检查是否结果集中所有文档这个客户端都能看到
- 服务端发送一个 `VersionQueryMessage` 查询客户端结果集 res 中所有文档的版本；
- 客户端收到 `VersionQueryMessage` 后，回复一条 `VersionQueryRespMessage` 告诉服务端它希望查询的文档的版本。由于此时客户端数据库为空，这个消息应该表示 “你问的文档我都没有”。
- 服务端收到 `VersionQueryRespMessage` 后，回复一条 `SyncMessage`：
  - 对所有客户端没有的文档，将文档的快照加入 `SyncMessage`
  - 对所有客户端过时的文档，将新旧文档的差量加入 `SyncMessage`
  - 注：由于客户端数据库为空，因此这里应该加入的全部是快照

存在的问题：

- 对结果集中每个文档都调用 `canView` 可能存在性能问题
- 如果一个结果集中有部分文档客户端有权限访问，但部分文档没有权限，应该怎么做？
- 如何保证 `VersionQueryRespMessage` 和 `VersionQueryMessage` 的一一对应关系？
  - 假设 `VersionQueryMessage` 只请求了 d1, d2 两个文档的版本，但客户端返回的 `VersionQueryRespMessage` 给出了 d1, d2, d3 三个文档的版本，则根据上面的逻辑这会造成越权（拿到 d3）。
  - 或者在生成 `SyncMessage` 时再次使用 `canView` 检验？

---

场景 2：客户端提交一个事务

我的设计：

- 如果客户端
  - 使用乐观更新策略：立即更新客户端数据库，并将这一事务放入待确认队列。每收到 `AckTransactionMessage`，检查被确认的事务是否在等待队列中，如果是，则将事务从待确认队列中移除。
  - 使用悲观更新策略：暂不更新客户端数据库，并将这一事务放入待确认队列。每收到 `AckTransactionMessage`，检查被确认的事务是否在等待队列中，如果是，则将事务从待确认队列中移除，并更新客户端数据库。
- 发送一个 `PostTransactionMessage` 向服务端提交这一事务；
- 服务端收到 `PostTransactionMessage` 后，尝试向存储引擎提交事务；
- 服务端同步器监听存储引擎的三类事件：
  - `TransactionCommitted`：表示一个事务提交成功
    - 向事务提交者发送 `AckTransactionMessage`
    - 对每个客户端 `ClientId`：
      - 初始化一个 `SyncMessage`；
      - 遍历这个客户端订阅集 `Set<Query>` 中的每个查询 q，如果 q 的结果集因为提交的事务而改变，则对比新旧结果集：
        - 对改变的文档，将新旧文档的差量加入 `SyncMessage`
        - 对新增的文档，将新文档的快照加入 `SyncMessage`
        - 对删除的文档，将这个文档的 key 加入删除集
      - 如果 `SyncMessage` 不为空，则将其发送给客户端；
  - `TransactionCanceled`：表示一个事务被取消，向事务提交者发送 `TransactionFailedMessage`，携带失败原因；
  - `TransactionRollbacked`：表示一个事务被回滚，向事务发送着发送 `TransactionFailedMessage`，携带失败原因；

存在的问题：

- 如何确定一个事务会不会影响一个查询？
  - 使用 [[Event-Reduce]] 算法。
- “将新旧结果集的差量加入 `SyncMessage`” 隐含假设：客户端的状态和旧结果集一致。如果客户端的状态比旧结果集更旧，则仅根据收到的 `SyncMessage` 无法更新客户端文档至最新版本。这种情况怎么办？
  - 方案一、客户端发送 `VersionQueryRespMessage` 消息告诉服务端自己的消息版本。但是这会导致 `VersionQueryRespMessage` 和 `VersionQueryMessage` 不再是一对一的，而可以单独用来通知服务端客户端文档的版本，可能会造成更多的问题。所以也许单独为这种场景设计一个新的消息类型（比如 `VersionGapMessage`）？
  - 方案二、在服务端维护每个客户端每个文档的版本（不好）
- 如果事务提交非常频繁，考虑使用节流算法

---

场景 3：客户端取消订阅一个查询

我的设计：

- 客户端发送一条 `SubscriptionUpdateMessage` 消息，通知服务端自己希望取消订阅一个查询；
- 服务端收到客户端发的 `SubscriptionUpdateMessage` 后，检查自己维护的这个客户端的订阅集，如果查询在订阅集中，则将之移除；

---

场景 4：客户端不为空，订阅了若干查询，现连接到服务端

我的设计：

（和场景 1 相似） #later

Running tool: /opt/homebrew/bin/go test -timeout 30s -run ^TestClientCodeStart$ github.com/jlwxz-capstone-project/rapierdb-server-go/test/network

=== RUN TestClientCodeStart
[INFO] 2025/04/06 14:21:52 dbPath: /var/folders/\_t/y3ftsbj95r969fqlhnhl7_b40000gn/T/TestClientCodeStart3726425168/001
2025/04/06 14:21:52 [JOB 1] WAL file /var/folders/\_t/y3ftsbj95r969fqlhnhl7_b40000gn/T/TestClientCodeStart3726425168/001/000002.log with log number 000002 stopped reading at offset: 4420; replayed 1 keys in 1 batches
[INFO] 2025/04/06 14:21:52 服务器启动中...
[DEBUG] 2025/04/06 14:21:52 NewSynchronizer: 正在创建同步器
[DEBUG] 2025/04/06 14:21:52 NewSynchronizer: 同步器创建成功
[DEBUG] 2025/04/06 14:21:52 Synchronizer.Start: 正在启动同步器
[DEBUG] 2025/04/06 14:21:52 Synchronizer.Start: 已订阅存储引擎事件
[DEBUG] 2025/04/06 14:21:52 Synchronizer.Start: 同步器启动完成
[INFO] 2025/04/06 14:21:52 服务器已启动
[DEBUG] 2025/04/06 14:21:52 客户端通道创建成功
[DEBUG] 2025/04/06 14:21:52 SSE 客户端状态改变：disconnected -> connecting
[DEBUG] 2025/04/06 14:21:52 Synchronizer: 事件处理协程已启动
[DEBUG] 2025/04/06 14:21:52 SSE 客户端状态改变：connecting -> connected
[DEBUG] 2025/04/06 14:21:52 SSE 接收启动成功，URL: http://localhost:8080/sse
[DEBUG] 2025/04/06 14:21:52 通道注册成功，当前活跃通道数量: 1
[INFO] 2025/04/06 14:21:52 客户端已启动
[WARN] 2025/04/06 14:21:52 collection users not found
[DEBUG] 2025/04/06 14:21:52 query1 之前结果：[]
[DEBUG] 2025/04/06 14:21:52 Synchronizer.msgHandler: 收到来自客户端 user1 的消息，长度 67 字节
[DEBUG] 2025/04/06 14:21:52 msgHandler: 收到 SubscriptionUpdateMessageV1{Added: [FindManyQuery{Collection: users, Filter: ValueExpr{Value: true}, Sort: [], Skip: 0, Limit: 0}], Removed: []} 来自 user1
[DEBUG] 2025/04/06 14:21:52 msgHandler: 客户端 user1 订阅查询 FindManyQuery{Collection: users, Filter: ValueExpr{Value: true}, Sort: [], Skip: 0, Limit: 0}
[DEBUG] 2025/04/06 14:21:52 msgHandler: 向客户端 user1 发送版本查询消息 VersionQueryMessageV1{Queries: {[users.user2]: {}, [users.user3]: {}, [users.user1]: {}}}
[DEBUG] 2025/04/06 14:21:52 客户端收到 VersionQueryMessageV1{Queries: {[users.user2]: {}, [users.user3]: {}, [users.user1]: {}}}
[DEBUG] 2025/04/06 14:21:52 Synchronizer.msgHandler: 收到来自客户端 user1 的消息，长度 110 字节
[DEBUG] 2025/04/06 14:21:52 msgHandler: 收到 VersionQueryRespMessageV1{Responses: {[users.user1]: , [users.user2]: , [users.user3]: }} 来自 user1
[DEBUG] 2025/04/06 14:21:52 msgHandler: 向客户端 user1 发送同步消息 PostDocMessageV1{Upsert: [users.user1, users.user2, users.user3], Delete: []}
[DEBUG] 2025/04/06 14:21:52 客户端收到 PostDocMessageV1{Upsert: [users.user1, users.user2, users.user3], Delete: []}
[DEBUG] 2025/04/06 14:21:52 文档 users.user1 不存在，加入到本地数据库
[DEBUG] 2025/04/06 14:21:52 文档 users.user2 不存在，加入到本地数据库
[DEBUG] 2025/04/06 14:21:52 文档 users.user3 不存在，加入到本地数据库
[DEBUG] 2025/04/06 14:21:54 query1 之后结果：[0x1400011c2b8 0x1400011c2d0 0x1400011c2e8]
[DEBUG] 2025/04/06 14:21:54 客户端提交事务 123e4567-e89b-12d3-a456-426614174001
[DEBUG] 2025/04/06 14:21:54 客户端将事务 123e4567-e89b-12d3-a456-426614174001 加入到待确认队列
[DEBUG] 2025/04/06 14:21:54 Synchronizer.msgHandler: 收到来自客户端 user1 的消息，长度 402 字节
[DEBUG] 2025/04/06 14:21:54 msgHandler: 收到 PostTransactionMessageV1{Transaction: &{123e4567-e89b-12d3-a456-426614174001 testdb user1 [0x14000484240]}} 来自 user1
[DEBUG] 2025/04/06 14:21:54 msgHandler: 正在提交事务 123e4567-e89b-12d3-a456-426614174001 到存储引擎
[DEBUG] 2025/04/06 14:21:54 msgHandler: 事务 123e4567-e89b-12d3-a456-426614174001 提交成功
[DEBUG] 2025/04/06 14:21:54 Synchronizer: 收到事务提交事件
[DEBUG] 2025/04/06 14:21:54 handleTransactionCommitted: 开始处理事务提交事件
[DEBUG] 2025/04/06 14:21:54 handleTransactionCommitted: 处理事务 123e4567-e89b-12d3-a456-426614174001 提交事件，提交者 user1
[DEBUG] 2025/04/06 14:21:54 handleTransactionCommitted: 向提交者 user1 发送确认消息
[DEBUG] 2025/04/06 14:21:54 handleTransactionCommitted: 当前连接的客户端数量: 1
[DEBUG] 2025/04/06 14:21:54 handleTransactionCommitted: 跳过提交者 user1
[DEBUG] 2025/04/06 14:21:54 handleTransactionCommitted: 事务 123e4567-e89b-12d3-a456-426614174001 处理完成
[DEBUG] 2025/04/06 14:21:54 客户端收到 AckTransactionMessageV1{TxID: 123e4567-e89b-12d3-a456-426614174001}
[DEBUG] 2025/04/06 14:21:54 客户端收到事务确认消息，应用事务 123e4567-e89b-12d3-a456-426614174001
[DEBUG] 2025/04/06 14:21:56 事务提交之后结果：[0x1400044a090 0x1400044a0a8 0x1400044a0c0 0x1400044a0d8]
[INFO] 2025/04/06 14:21:57 开始关闭服务器...
[DEBUG] 2025/04/06 14:21:57 Synchronizer.Stop: 正在停止同步器
[DEBUG] 2025/04/06 14:21:57 Synchronizer.Stop: 发送取消信号
[DEBUG] 2025/04/06 14:21:57 Synchronizer.Stop: 取消存储引擎事件订阅
[DEBUG] 2025/04/06 14:21:57 Synchronizer.Stop: 卸载消息处理器
[DEBUG] 2025/04/06 14:21:57 Synchronizer.Stop: 关闭所有连接
[DEBUG] 2025/04/06 14:21:57 Synchronizer: 收到取消信号，事件处理协程退出
[DEBUG] 2025/04/06 14:21:57 Synchronizer.Stop: 同步器已停止
[INFO] 2025/04/06 14:21:57 服务器已完全关闭
[DEBUG] 2025/04/06 14:21:57 SSE 客户端状态改变：connected -> disconnected
[INFO] 2025/04/06 14:21:58 测试成功完成
--- PASS: TestClientCodeStart (6.10s)
PASS
ok github.com/jlwxz-capstone-project/rapierdb-server-go/test/network 6.507s

authenitcation <--

authorization <--
