# Model Routing Configuration

LiteLLM Go Gateway 支持基于模型名的自动路由，可以将不同的模型请求路由到对应的上游 provider，并支持自动 fallback。

## 默认路由规则

如果没有配置自定义路由规则，网关会使用以下默认规则：

- `gpt-*`, `text-*`, `davinci*` → OpenAI provider
- `claude-*` → Anthropic provider

**Fallback**: 如果模型不匹配任何规则，会使用第一个可用的 provider。

## 自定义路由配置

路由规则可以通过代码或配置文件进行自定义（未来版本将支持配置文件）。

### 路由规则格式

每条路由规则包含：
- **Pattern**: 模型名匹配模式（正则表达式）
- **Providers**: provider 名称列表（按优先级排序）

第一个 provider 是主 provider，其余是 fallback providers。

### 示例

```go
// 示例：通过代码配置路由规则
cfg := config.Config{
    Routing: config.RoutingConfig{
        Rules: []config.RoutingRule{
            {
                Pattern:   "^gpt-4.*",
                Providers: []string{"openai", "azure"},  // OpenAI 为主，Azure 为 fallback
            },
            {
                Pattern:   "^gpt-3.5.*",
                Providers: []string{"azure", "openai"},  // Azure 为主，OpenAI 为 fallback
            },
            {
                Pattern:   "^claude-.*",
                Providers: []string{"anthropic"},
            },
        },
    },
}
```

## Fallback 策略

当主 provider 失败时，网关会自动尝试 fallback providers：

1. **非流式请求** (`stream: false`):
   - 按顺序尝试所有配置的 providers
   - 每个 provider 都支持自动重试（仅网络错误）
   - 如果所有 providers 都失败，返回最后一个错误

2. **流式请求** (`stream: true`):
   - 如果主 provider 在建立流之前失败，尝试 fallback providers
   - 流建立后不支持 fallback（避免客户端收到重复数据）

## 日志示例

启用 fallback 后的日志示例：

```json
{
  "level": "INFO",
  "msg": "starting chat completion",
  "model": "gpt-4",
  "stream": false,
  "primary_provider": "openai",
  "fallback_count": 1,
  "request_id": "abc-123"
}

{
  "level": "WARN",
  "msg": "provider failed",
  "provider": "openai",
  "error": "connection timeout",
  "request_id": "abc-123"
}

{
  "level": "INFO",
  "msg": "trying fallback provider",
  "provider": "azure",
  "attempt": 2,
  "total_providers": 2,
  "request_id": "abc-123"
}

{
  "level": "INFO",
  "msg": "chat completion succeeded",
  "provider": "azure",
  "request_id": "abc-123"
}
```

## Provider 优先级

配置多个 provider 时，建议考虑：

1. **延迟**: 选择延迟较低的 provider 作为主 provider
2. **可靠性**: 将更稳定的 provider 放在前面
3. **成本**: 根据价格策略调整 provider 顺序
4. **配额**: 将配额充足的 provider 作为主 provider

## 未来功能

计划支持的高级路由功能：

- [ ] 基于权重的负载均衡
- [ ] 基于成本的智能路由
- [ ] 基于响应时间的动态调整
- [ ] Provider 健康检查和自动熔断
- [ ] 按租户/用户的路由策略
- [ ] 通过配置文件或环境变量配置路由规则
