# 使用日志功能

litellm-go-gateway 自动将每次请求的 token 使用情况记录到 SQLite 数据库中。

## 配置

在 `.env` 文件中配置数据库路径：

```bash
# SQLite database path for usage logs. Default: ./data/usage.db
DATABASE_PATH=./data/usage.db
```

数据库文件和目录会在首次启动时自动创建。

## 记录的信息

每次聊天请求（无论成功或失败）都会记录以下信息：

- **request_id**: 请求的唯一标识符
- **provider**: 使用的 provider（openai/anthropic/azure）
- **model**: 请求的模型名称
- **prompt_tokens**: 输入 token 数量
- **completion_tokens**: 输出 token 数量
- **total_tokens**: 总 token 数量
- **success**: 请求是否成功（true/false）
- **error_code**: 失败时的错误码
- **duration**: 请求耗时（毫秒）
- **created_at**: 记录创建时间

## 查询使用记录

### 使用提供的查询工具

```bash
go run query_usage.go
```

这将显示最近 5 条使用记录。

### 使用 SQL 查询

如果系统安装了 `sqlite3` 命令行工具：

```bash
# 查看最近的记录
sqlite3 data/usage.db "SELECT * FROM usage_logs ORDER BY created_at DESC LIMIT 10;"

# 按 provider 统计 token 使用量
sqlite3 data/usage.db "SELECT provider, SUM(total_tokens) as total FROM usage_logs GROUP BY provider;"

# 查询某个时间段的使用情况
sqlite3 data/usage.db "SELECT * FROM usage_logs WHERE created_at >= datetime('now', '-1 day');"
```

## 数据库表结构

```sql
CREATE TABLE usage_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    success BOOLEAN NOT NULL DEFAULT 1,
    error_code TEXT,
    duration INTEGER NOT NULL,
    created_at DATETIME NOT NULL
);

-- 索引
CREATE INDEX idx_created_at ON usage_logs(created_at);
CREATE INDEX idx_provider ON usage_logs(provider);
CREATE INDEX idx_request_id ON usage_logs(request_id);
```

## 注意事项

1. **异步记录**: 日志记录是异步的，不会阻塞客户端响应
2. **失败时的 token 数**: 如果请求失败，token 数量可能为 0（因为上游未返回 usage 信息）
3. **数据库驱动**: 使用 `modernc.org/sqlite`（纯 Go 实现），无需 CGO
4. **性能**: SQLite 足以处理中小规模的使用记录。大规模生产环境建议使用 PostgreSQL 或 MySQL
