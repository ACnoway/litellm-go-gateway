# 结构化日志功能

## 概述

项目已集成结构化日志功能，使用 Go 标准库的 `log/slog`，每个请求都会记录唯一的请求 ID，便于追踪和调试。

## 功能特性

1. **请求 ID 追踪**
   - 每个请求自动生成或沿用客户端提供的 `X-Request-ID`
   - 请求 ID 通过 context 传递到所有日志点
   - 响应头中返回 `X-Request-ID`，方便客户端关联日志

2. **结构化日志格式**
   - 支持 JSON 格式（生产环境推荐）
   - 支持文本格式（本地开发人类可读）
   - 通过 `LOG_FORMAT` 环境变量配置

3. **日志覆盖点**
   - HTTP 请求/响应（方法、路径、状态码、耗时、客户端 IP）
   - Service 层（聊天完成开始/成功/失败、重试逻辑）
   - 自动跳过健康检查端点的日志（避免噪音）

## 配置

在 `.env` 文件中添加：

```bash
# 日志格式: "json" 或 "text"
# 生产环境推荐 json，本地开发推荐 text
LOG_FORMAT=json
```

## 日志示例

### JSON 格式（生产环境）

```json
{
  "time": "2026-07-21T10:30:45.123Z",
  "level": "INFO",
  "msg": "request completed",
  "request_id": "a1b2c3d4e5f6g7h8",
  "method": "POST",
  "path": "/v1/chat/completions",
  "status": 200,
  "duration_ms": 1234,
  "client_ip": "192.168.1.100"
}

{
  "time": "2026-07-21T10:30:45.000Z",
  "level": "INFO",
  "msg": "starting chat completion",
  "request_id": "a1b2c3d4e5f6g7h8",
  "model": "gpt-4o",
  "stream": false,
  "provider": "openai"
}

{
  "time": "2026-07-21T10:30:46.234Z",
  "level": "INFO",
  "msg": "chat completion succeeded",
  "request_id": "a1b2c3d4e5f6g7h8",
  "provider": "openai"
}
```

### 文本格式（本地开发）

```
time=2026-07-21T10:30:45.123+08:00 level=INFO msg="request completed" request_id=a1b2c3d4e5f6g7h8 method=POST path=/v1/chat/completions status=200 duration_ms=1234 client_ip=192.168.1.100
time=2026-07-21T10:30:45.000+08:00 level=INFO msg="starting chat completion" request_id=a1b2c3d4e5f6g7h8 model=gpt-4o stream=false provider=openai
time=2026-07-21T10:30:46.234+08:00 level=INFO msg="chat completion succeeded" request_id=a1b2c3d4e5f6g7h8 provider=openai
```

## 重试日志示例

当发生网络错误需要重试时：

```json
{
  "time": "2026-07-21T10:31:00.000Z",
  "level": "WARN",
  "msg": "retryable error occurred, will retry",
  "request_id": "x1y2z3a4b5c6d7e8",
  "attempt": 1,
  "max_attempts": 3,
  "error": "dial tcp: connection refused",
  "next_delay_ms": 100
}

{
  "time": "2026-07-21T10:31:00.200Z",
  "level": "INFO",
  "msg": "retry succeeded",
  "request_id": "x1y2z3a4b5c6d7e8",
  "attempt": 2,
  "total_attempts": 3
}
```

## 自定义请求 ID

客户端可以提供自定义的请求 ID：

```bash
curl -H "X-Request-ID: my-custom-id-12345" \
     -H "Authorization: Bearer your-key" \
     http://localhost:8080/v1/models
```

服务器会沿用该 ID 并在响应头和日志中使用。

## 代码集成示例

在需要记录日志的地方，从 context 获取带请求 ID 的日志器：

```go
import "github.com/acnoway/litellm-go-gateway/internal/pkg/logger"

func YourFunction(ctx context.Context) {
    log := logger.FromContext(ctx)
    
    log.Info("operation started", "key", "value")
    log.Error("operation failed", "error", err)
    log.Warn("warning message", "detail", "info")
}
```

日志会自动包含请求 ID（如果 context 中存在）。

## 日志级别

当前配置为 `INFO` 级别，记录：
- `INFO`: 正常操作（请求完成、操作成功）
- `WARN`: 警告信息（重试、非致命错误）
- `ERROR`: 错误信息（操作失败、配置错误）

## 性能考虑

- 健康检查端点 (`/healthz`, `/readyz`) 不记录日志，避免高频探测产生日志噪音
- 请求 ID 使用 crypto/rand 生成，确保随机性和唯一性
- JSON 格式适合日志聚合系统（如 ELK、Loki）解析

## 架构说明

日志功能遵循项目的分层架构：

- **`internal/pkg/logger/`**: 日志包，提供 context-aware 的日志器
- **`internal/transport/httpapi/middleware.go`**: HTTP 中间件，记录请求/响应
- **`internal/service/chat.go`**: Service 层日志（业务操作）
- **`cmd/gateway/main.go`**: 应用启动时初始化日志器

请求 ID 通过 Gin 中间件注入到 context，随请求传递到所有层。
