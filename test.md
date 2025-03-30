client empty

subscript update message (client query) 我想订阅这个查询
client -> queries map 每个客户端订阅的查询
exec query => docs 结果集
version query message
version query response message
missing docs, out dated docs
sync message 发给客户端
update 自己的客户端数据库

client

client2 更新了服务端数据库

一个文档改了

- 知道这个更改会影响哪些查询 => Event reduce 算法
- 知道要向哪些客户端发送 sync message，sync message update 消息是一个更新向量
  还需知道客户端的版本才能计算

=== RUN TestBasicCURD
[INFO] 2025/03/30 20:18:29 dbPath: /var/folders/\_t/y3ftsbj95r969fqlhnhl7_b40000gn/T/TestBasicCURD3920669294/001
2025/03/30 20:18:29 [JOB 1] WAL file /var/folders/\_t/y3ftsbj95r969fqlhnhl7_b40000gn/T/TestBasicCURD3920669294/001/000002.log with log number 000002 stopped reading at offset: 4420; replayed 1 keys in 1 batches
[INFO] 2025/03/30 20:18:29 服务器启动中...
[DEBUG] 2025/03/30 20:18:29 NewSynchronizer: 正在创建同步器
[DEBUG] 2025/03/30 20:18:29 NewSynchronizer: 同步器创建成功
[DEBUG] 2025/03/30 20:18:29 Synchronizer.Start: 正在启动同步器
[DEBUG] 2025/03/30 20:18:29 Synchronizer.Start: 已订阅存储引擎事件
[DEBUG] 2025/03/30 20:18:29 Synchronizer.Start: 同步器启动完成
[INFO] 2025/03/30 20:18:29 服务器已启动
[DEBUG] 2025/03/30 20:18:29 Synchronizer: 事件处理协程已启动
[DEBUG] 2025/03/30 20:18:29 客户端通道创建成功
[DEBUG] 2025/03/30 20:18:29 SSE 客户端状态改变：disconnected -> connecting
[DEBUG] 2025/03/30 20:18:29 SSE 客户端状态改变：connecting -> connected
[DEBUG] 2025/03/30 20:18:29 SSE 接收启动成功，URL: http://localhost:8080/sse
[DEBUG] 2025/03/30 20:18:29 通道注册成功，当前活跃通道数量: 1
[INFO] 2025/03/30 20:18:29 客户端已启动
[WARN] 2025/03/30 20:18:29 collection users not found
[DEBUG] 2025/03/30 20:18:29 query1 之前结果：[]
[DEBUG] 2025/03/30 20:18:29 Synchronizer.msgHandler: 收到来自客户端 的消息，长度 67 字节
[DEBUG] 2025/03/30 20:18:29 msgHandler: 收到 SubscriptionUpdateMessageV1{Added: [FindManyQuery{Collection: users, Filter: ValueExpr{Value: true}, Sort: [], Skip: 0, Limit: 0}], Removed: []} 来自
[DEBUG] 2025/03/30 20:18:29 msgHandler: 客户端 订阅查询 FindManyQuery{Collection: users, Filter: ValueExpr{Value: true}, Sort: [], Skip: 0, Limit: 0}
[DEBUG] 2025/03/30 20:18:29 msgHandler: 向客户端 发送版本查询消息 VersionQueryMessageV1{Queries: {[users.user1]: {}, [users.user2]: {}, [users.user3]: {}}}
[DEBUG] 2025/03/30 20:18:29 客户端收到 VersionQueryMessageV1{Queries: {[users.user2]: {}, [users.user3]: {}, [users.user1]: {}}}
[DEBUG] 2025/03/30 20:18:29 Synchronizer.msgHandler: 收到来自客户端 的消息，长度 110 字节
[DEBUG] 2025/03/30 20:18:29 msgHandler: 收到 VersionQueryRespMessageV1{Responses: {[users.user3]: , [users.user1]: , [users.user2]: }} 来自
[DEBUG] 2025/03/30 20:18:29 msgHandler: 向客户端 发送同步消息 PostDocMessageV1{Upsert: [users.user3, users.user1, users.user2], Delete: []}
[DEBUG] 2025/03/30 20:18:29 客户端收到 PostDocMessageV1{Upsert: [users.user2, users.user3, users.user1], Delete: []}
[DEBUG] 2025/03/30 20:18:29 文档 users.user2 不存在，加入到本地数据库
[DEBUG] 2025/03/30 20:18:29 文档 users.user3 不存在，加入到本地数据库
[DEBUG] 2025/03/30 20:18:29 文档 users.user1 不存在，加入到本地数据库
[DEBUG] 2025/03/30 20:18:31 query1 之后结果：[0x1400000e198 0x1400000e1b0 0x1400000e1c8]
[INFO] 2025/03/30 20:18:34 开始关闭服务器...
[DEBUG] 2025/03/30 20:18:34 Synchronizer.Stop: 正在停止同步器
[DEBUG] 2025/03/30 20:18:34 Synchronizer.Stop: 发送取消信号
[DEBUG] 2025/03/30 20:18:34 Synchronizer: 收到取消信号，事件处理协程退出
[DEBUG] 2025/03/30 20:18:34 Synchronizer.Stop: 取消存储引擎事件订阅
[DEBUG] 2025/03/30 20:18:34 Synchronizer.Stop: 卸载消息处理器
[DEBUG] 2025/03/30 20:18:34 Synchronizer.Stop: 关闭所有连接
[DEBUG] 2025/03/30 20:18:34 Synchronizer.Stop: 同步器已停止
[INFO] 2025/03/30 20:18:34 服务器已完全关闭
[DEBUG] 2025/03/30 20:18:35 SSE 客户端状态改变：connected -> disconnected
[DEBUG] 2025/03/30 20:18:35 SSE 客户端状态改变：disconnected -> disconnected
[INFO] 2025/03/30 20:18:35 测试成功完成
--- PASS: TestBasicCURD (6.09s)
PASS
ok github.com/jlwxz-capstone-project/rapierdb-server-go/test/network 6.502s
