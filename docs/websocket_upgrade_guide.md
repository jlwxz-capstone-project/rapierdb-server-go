# RapierDB SSE 升级为 WebSocket 指南

## 概述

本指南详细说明如何将RapierDB项目从当前的HTTP+SSE架构升级为WebSocket架构，以获得更好的性能和双向通信能力。

## 升级优势

### WebSocket vs SSE 对比

| 特性 | SSE (Server-Sent Events) | WebSocket |
|------|--------------------------|-----------|
| 通信方向 | 单向 (服务器→客户端) | 双向 |
| 协议开销 | HTTP头部开销较大 | 低开销帧格式 |
| 二进制数据 | 需要Base64编码 | 原生支持 |
| 连接管理 | 自动重连 | 手动管理 |
| 延迟 | 相对较高 | 更低 |
| 浏览器支持 | 良好 | 优秀 |

### 性能提升预期

- **延迟降低**: 预计减少20-40%的往返时间
- **吞吐量提升**: 预计提升30-50%的消息处理能力
- **资源使用**: 减少网络开销和CPU使用率

## 升级步骤

### 步骤1: 安装依赖

```bash
# 添加WebSocket库依赖
go get github.com/gorilla/websocket

# 更新go.mod
go mod tidy
```

### 步骤2: 服务器端升级

#### 原有SSE服务器配置：
```go
// 原有 HTTP+SSE 网络配置
networkOpts := &network_server.HttpNetworkOptions{
    BaseUrl:         "localhost:8080",
    ReceiveEndpoint: "/api",
    SendEndpoint:    "/sse",
}
network := network_server.NewHttpNetworkWithContext(networkOpts, ctx)
```

#### 升级为WebSocket服务器：
```go
// 新的 WebSocket 网络配置
networkOpts := &network_server.WebSocketNetworkOptions{
    BaseUrl:       "localhost:8080",
    WebSocketPath: "/ws",
    AllowOrigin:   "*", // 配置CORS
}
network := network_server.NewWebSocketNetworkWithContext(networkOpts, ctx)
```

### 步骤3: 客户端升级

#### 原有SSE客户端配置：
```go
// 原有 HTTP+SSE 客户端配置
clientOpts := &network_client.HttpNetworkOptions{
    BackendUrl:      "http://localhost:8080",
    ReceiveEndpoint: "/sse",
    SendEndpoint:    "/api",
    Headers: map[string]string{
        "X-Client-ID": clientId,
    },
}
client := network_client.NewHttpNetworkWithContext(clientOpts, ctx)
```

#### 升级为WebSocket客户端：
```go
// 新的 WebSocket 客户端配置
clientOpts := &network_client.WebSocketNetworkOptions{
    ServerUrl:     "http://localhost:8080",
    WebSocketPath: "/ws",
    Headers: map[string][]string{
        "X-Client-ID": {clientId},
    },
}
client := network_client.NewWebSocketNetworkWithContext(clientOpts, ctx)
```

### 步骤4: 同步器升级

#### 在同步器中使用WebSocket：
```go
// 创建WebSocket网络提供者
network := network_server.NewWebSocketNetworkWithContext(&network_server.WebSocketNetworkOptions{
    BaseUrl:       "localhost:8080",
    WebSocketPath: "/ws",
    AllowOrigin:   "*",
}, ctx)

// 使用WebSocket网络创建同步器
synchronizer := synchronizer2.NewSynchronizerWithContext(ctx, &synchronizer2.SynchronizerParams{
    DbConnector: dbConnector,
    Network:     network,  // 使用WebSocket网络
    DbUrl:       dbUrl,
})
```

## 配置选项

### WebSocket服务器配置

```go
type WebSocketNetworkOptions struct {
    BaseUrl          string                    // 服务器监听地址
    WebSocketPath    string                    // WebSocket端点路径
    ShutdownTimeout  time.Duration             // 优雅关闭超时
    Authenticator    auth.Authenticator        // 认证器
    AllowOrigin      string                    // CORS允许的源
    CheckOrigin      func(*http.Request) bool  // 自定义源检查
    ReadBufferSize   int                       // 读缓冲区大小
    WriteBufferSize  int                       // 写缓冲区大小
    HandshakeTimeout time.Duration             // 握手超时
}
```

### WebSocket客户端配置

```go
type WebSocketNetworkOptions struct {
    ServerUrl       string                // 服务器URL
    WebSocketPath   string                // WebSocket端点路径
    Headers         map[string][]string   // HTTP头部
    HandshakeTimeout time.Duration        // 握手超时
    ReadBufferSize  int                   // 读缓冲区大小
    WriteBufferSize int                   // 写缓冲区大小
    ReconnectDelay  time.Duration         // 重连延迟
}
```

## 测试和验证

### 运行WebSocket测试

```bash
# 基础WebSocket功能测试
go test -run TestWebSocketServerNetwork ./test/network/

# WebSocket客户端-服务器测试
go test -run TestWebSocketClientServer ./test/network/

# 性能对比测试
go test -run TestWebSocketVsSSEComparison ./test/synchronizer/

# 完整同步器测试
go test -run TestWebSocketSynchronizer ./test/synchronizer/
```

### 性能测试

```bash
# 运行性能基准测试
go test -bench=. -run TestWebSocketPerformance ./test/network/

# 长时间运行测试
go test -timeout 300s -run TestWebSocketSynchronizer ./test/synchronizer/
```

## 迁移策略

### 渐进式迁移

1. **并行部署**: 同时运行SSE和WebSocket服务器，在不同端口
2. **客户端切换**: 逐步将客户端切换到WebSocket端点
3. **监控对比**: 比较两种实现的性能指标
4. **完全切换**: 确认稳定后停用SSE服务器

### 兼容性考虑

- **向后兼容**: 保留SSE实现作为fallback选项
- **特性标识**: 使用配置标识启用/禁用WebSocket
- **错误处理**: 实现WebSocket连接失败时的SSE降级

## 故障排除

### 常见问题

1. **连接失败**
   - 检查防火墙设置
   - 验证WebSocket端点路径
   - 确认CORS配置

2. **性能问题**
   - 调整缓冲区大小
   - 检查Ping/Pong间隔
   - 监控内存使用

3. **认证问题**
   - 验证认证头部传递
   - 检查认证器配置
   - 确认客户端ID格式

### 调试技巧

```go
// 启用WebSocket调试日志
log.SetLevel(log.DebugLevel)

// 监控连接状态
statusCh := network.SubscribeStatusChange()
go func() {
    for status := range statusCh {
        log.Debugf("Network status changed: %v", status)
    }
}()

// 监控连接关闭事件
connClosedCh := network.SubscribeConnectionClosed()
go func() {
    for event := range connClosedCh {
        log.Debugf("Connection closed: %s", event.ClientId)
    }
}()
```

## 示例代码

### 完整的WebSocket服务器示例

```go
package main

import (
    "context"
    "log"
    
    network_server "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/server"
)

func main() {
    ctx := context.Background()
    
    // 创建WebSocket服务器
    server := network_server.NewWebSocketNetworkWithContext(&network_server.WebSocketNetworkOptions{
        BaseUrl:       "localhost:8080",
        WebSocketPath: "/ws",
        AllowOrigin:   "*",
    }, ctx)
    
    // 设置消息处理器
    server.SetMsgHandler(func(clientId string, msg []byte) {
        log.Printf("Received from %s: %s", clientId, string(msg))
        // 回显消息
        server.Send(clientId, []byte("Echo: "+string(msg)))
    })
    
    // 启动服务器
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
    
    log.Println("WebSocket server started on ws://localhost:8080/ws")
    select {} // 保持运行
}
```

### 完整的WebSocket客户端示例

```go
package main

import (
    "context"
    "log"
    "time"
    
    network_client "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/client"
)

func main() {
    ctx := context.Background()
    
    // 创建WebSocket客户端
    client := network_client.NewWebSocketNetworkWithContext(&network_client.WebSocketNetworkOptions{
        ServerUrl:     "http://localhost:8080",
        WebSocketPath: "/ws",
        Headers: map[string][]string{
            "X-Client-ID": {"test-client"},
        },
    }, ctx)
    
    // 设置消息处理器
    client.SetMsgHandler(func(msg []byte) {
        log.Printf("Received: %s", string(msg))
    })
    
    // 连接服务器
    if err := client.Connect(); err != nil {
        log.Fatal(err)
    }
    
    // 发送测试消息
    client.Send([]byte("Hello WebSocket!"))
    
    // 保持连接
    time.Sleep(5 * time.Second)
    client.Close()
}
```

## 总结

WebSocket升级将为RapierDB带来显著的性能提升和更好的用户体验。通过本指南的步骤，您可以安全地将现有的SSE架构迁移到WebSocket，同时保持系统的稳定性和兼容性。

建议在生产环境中采用渐进式迁移策略，充分测试后再完全切换到WebSocket架构。 