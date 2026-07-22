# Milestone Implementation Plan

## Milestone 1: 最小 OpenAI 代理 - 状态检查

### ✅ 已完成
- [x] Gin 路由与 Kratos 启动装配
- [x] `/healthz`、`/readyz` 端点
- [x] `/v1/chat/completions` 端点（普通和流式）
- [x] OpenAI provider 实现（Chat + ChatStream）
- [x] OpenAI 风格错误响应
- [x] 客户端断开取消（通过 context 传递）
- [x] 请求日志（结构化日志）
- [x] Request ID 中间件
- [x] 自动重试与 fallback 机制
- [x] SQLite 使用日志记录
- [x] **Prometheus metrics**
  - HTTP 请求总数、响应时间直方图、错误率
  - 按路径、方法、状态码分组
  - Provider 调用次数、成功/失败计数、响应时间
  - Token 使用统计（prompt/completion tokens）
- [x] **`/metrics` 端点**（无需认证，便于监控系统访问）
- [x] **显式的上游 timeout 配置**（各 provider 已有 30 秒超时）

### ❌ 缺失功能
无 - Milestone 1 已完成！

---

## Milestone 2: 模型抽象与 OpenAI-compatible - 实现大纲

### 概念澄清
当前架构已经实现了部分 Milestone 2 的功能：
- ✅ Provider 接口抽象（`biz.Provider`）
- ✅ 模型路由（`provider.Router` + `provider.Manager`）
- ✅ Retry、fallback 机制
- ✅ 动态路由配置（通过数据库）

需要补充的核心功能：
- **Deployment 概念**：逻辑模型名到物理部署的映射
- **YAML 配置**：模型/部署的声明式配置
- **GET /v1/models**：返回可用的逻辑模型列表
- **OpenAI-compatible provider**：支持任意 OpenAI 兼容端点
- **负载均衡**：Round-robin / weighted-random

---

## Milestone 2 实现大纲

### Phase 1: 核心抽象层设计 (2-3 天)

#### 1.1 Deployment 领域模型
```go
// internal/biz/deployment.go
type Deployment struct {
    Name           string   // 逻辑模型名（暴露给用户）
    ActualModel    string   // 真实模型名（发给上游）
    Providers      []string // 可用的 providers（按优先级）
    Strategy       string   // 负载均衡策略: "priority", "round-robin", "weighted"
    Weights        []int    // weighted 策略的权重
    MaxTokens      int      // 模型的最大 token 限制
    Description    string   // 描述
}

type DeploymentRepo interface {
    List() ([]Deployment, error)
    Get(name string) (*Deployment, error)
    Create(d *Deployment) error
    Update(name string, d *Deployment) error
    Delete(name string) error
}
```

**交付物：**
- `internal/biz/deployment.go` - 领域模型
- `internal/data/deployment.go` - SQLite 持久化
- 数据库表：`deployments`

---

#### 1.2 YAML 配置支持
```yaml
# config/deployments.yaml (可选，优先级低于数据库)
deployments:
  - name: gpt-4-turbo
    actual_model: gpt-4-turbo-2024-04-09
    providers: [openai, azure]
    strategy: priority
    max_tokens: 128000
    
  - name: claude-opus
    actual_model: claude-3-opus-20240229
    providers: [anthropic]
    strategy: priority
    max_tokens: 200000
    
  - name: gpt-3.5-distributed
    actual_model: gpt-3.5-turbo
    providers: [openai, azure, openai-compatible]
    strategy: round-robin
    max_tokens: 16385
```

**交付物：**
- `internal/config/deployment.go` - YAML 解析
- `config/deployments.example.yaml` - 示例配置
- 启动时加载逻辑：database > YAML > 默认规则

---

### Phase 2: Deployment 管理 (2 天)

#### 2.1 Deployment Service
```go
// internal/service/deployment.go
type DeploymentService struct {
    repo            biz.DeploymentRepo
    providerManager *provider.Manager
}

func (s *DeploymentService) ListDeployments() ([]biz.Deployment, error)
func (s *DeploymentService) GetDeployment(name string) (*biz.Deployment, error)
func (s *DeploymentService) CreateDeployment(req biz.DeploymentRequest) error
func (s *DeploymentService) UpdateDeployment(name string, req biz.DeploymentRequest) error
func (s *DeploymentService) DeleteDeployment(name string) error
func (s *DeploymentService) ResolveModel(logicalName string) (actualModel string, providers []biz.Provider, error)
```

**交付物：**
- `internal/service/deployment.go`
- 模型解析逻辑（逻辑名 → 物理名 + providers）

---

#### 2.2 Deployment HTTP API
```
GET    /admin/deployments          # 列出所有部署
GET    /admin/deployments/:name    # 获取部署详情
POST   /admin/deployments          # 创建部署
PUT    /admin/deployments/:name    # 更新部署
DELETE /admin/deployments/:name    # 删除部署
```

**交付物：**
- `internal/transport/httpapi/deployment_handler.go`
- 更新 `admin_handler.go` 注册路由

---

### Phase 3: GET /v1/models 实现 (1 天)

#### 3.1 Models 端点
```go
// 返回所有可用的逻辑模型
GET /v1/models
{
  "object": "list",
  "data": [
    {
      "id": "gpt-4-turbo",
      "object": "model",
      "created": 1686935002,
      "owned_by": "openai",
      "permission": [...],
      "root": "gpt-4-turbo",
      "parent": null
    },
    {
      "id": "claude-opus",
      "object": "model",
      "created": 1686935002,
      "owned_by": "anthropic"
    }
  ]
}
```

**实现要点：**
- 从 `DeploymentRepo` 读取所有部署
- 转换为 OpenAI `/v1/models` 格式
- 支持过滤（可选）

**交付物：**
- 更新 `internal/transport/httpapi/handler.go` 的 `models()` 方法
- 从静态数据改为动态读取 deployments

---

### Phase 4: 负载均衡策略 (2-3 天)

#### 4.1 Provider Selector 抽象
```go
// internal/provider/selector.go
type Selector interface {
    Select(providers []biz.Provider) biz.Provider
}

type PrioritySelector struct{}        // 总是选第一个
type RoundRobinSelector struct{}      // 轮询
type WeightedRandomSelector struct{}  // 按权重随机
```

**实现细节：**
- `RoundRobinSelector` 需要线程安全的计数器（使用 `sync/atomic`）
- `WeightedRandomSelector` 使用加权随机算法
- Provider 失败时自动切换到下一个

**交付物：**
- `internal/provider/selector.go` - 选择器接口和实现
- `internal/provider/selector_test.go` - 单元测试

---

#### 4.2 集成到 Router
```go
// internal/provider/router.go
type ModelRoute struct {
    Pattern   string
    Providers []biz.Provider
    Selector  Selector  // 新增
}

func (r *ModelRouter) SelectProvider(model string) biz.Provider {
    route := r.findRoute(model)
    return route.Selector.Select(route.Providers)
}
```

**交付物：**
- 更新 `internal/provider/router.go`
- 更新 `internal/provider/manager.go` 使用新的选择逻辑

---

### Phase 5: OpenAI-compatible Provider (2 天)

#### 5.1 通用 OpenAI-compatible Provider
```go
// internal/provider/openai_compatible/provider.go
type Provider struct {
    name    string  // 自定义名称，如 "custom-openai-1"
    apiKey  string
    baseURL string
    client  *http.Client
}

// 复用 OpenAI provider 的实现，只改 baseURL
```

**配置示例：**
```env
# .env
CUSTOM_OPENAI_1_API_KEY=sk-...
CUSTOM_OPENAI_1_BASE_URL=https://api.groq.com/openai/v1
CUSTOM_OPENAI_1_NAME=groq

CUSTOM_OPENAI_2_API_KEY=sk-...
CUSTOM_OPENAI_2_BASE_URL=https://api.together.xyz/v1
CUSTOM_OPENAI_2_NAME=together
```

**实现要点：**
- 支持多个 OpenAI-compatible 端点
- 通过环境变量或 YAML 配置
- 自动注册到 provider registry

**交付物：**
- `internal/provider/openai_compatible/provider.go`
- `internal/provider/openai_compatible/register.go`
- `internal/config/config.go` 添加配置结构

---

### Phase 6: 测试与文档 (1-2 天)

#### 6.1 集成测试
- 测试逻辑模型名映射
- 测试负载均衡策略
- 测试 fallback 机制
- 测试 `/v1/models` 端点

#### 6.2 文档更新
- `docs/DEPLOYMENT.md` - Deployment 概念和配置
- `docs/LOAD_BALANCING.md` - 负载均衡策略说明
- `docs/OPENAI_COMPATIBLE.md` - 如何添加 OpenAI 兼容端点
- 更新 `README.md`

---

## 实现顺序建议

### 第一周：补全 Milestone 1
1. **Day 1-2**: 添加 Prometheus metrics
   - 实现 metrics 中间件
   - 添加 `/metrics` 端点
   - 验证 metrics 采集

2. **Day 3**: 验收测试
   - 使用 OpenAI SDK 测试
   - 测试所有 Milestone 1 功能
   - 编写测试文档

### 第二周：Milestone 2 核心
3. **Day 4-5**: Phase 1 - Deployment 模型
   - 实现 `Deployment` 领域模型
   - 实现 SQLite 持久化
   - YAML 配置支持

4. **Day 6-7**: Phase 2 - Deployment 管理
   - Deployment Service
   - Deployment HTTP API
   - 集成到现有架构

5. **Day 8**: Phase 3 - `/v1/models` 端点
   - 实现动态模型列表
   - 集成测试

### 第三周：Milestone 2 高级
6. **Day 9-10**: Phase 4 - 负载均衡
   - 实现选择器接口
   - RoundRobin、Weighted 实现
   - 集成到路由器

7. **Day 11-12**: Phase 5 - OpenAI-compatible
   - 通用 provider 实现
   - 配置支持
   - 测试多端点

8. **Day 13-14**: Phase 6 - 测试与文档
   - 集成测试
   - 性能测试
   - 文档完善

---

## 验收标准

### Milestone 1 补全验收
- [ ] Prometheus metrics 正常采集
- [ ] `/metrics` 端点返回所有指标
- [ ] OpenAI SDK 替换 base_url 和 API Key 后正常工作
- [ ] 压力测试：1000 req/s 稳定运行

### Milestone 2 验收
- [ ] 可以通过 YAML 或 API 配置逻辑模型
- [ ] `/v1/models` 返回所有配置的模型
- [ ] 逻辑模型名 `gpt-4-turbo` 可以映射到多个物理 provider
- [ ] Round-robin 均匀分配请求
- [ ] Weighted 按权重分配请求
- [ ] Fallback 在主 provider 失败时工作
- [ ] 添加新 OpenAI-compatible 端点无需修改代码
- [ ] 客户端无需修改即可切换 deployment 配置

---

## 技术债务与优化
- [ ] 缓存 deployment 配置（避免每次请求查数据库）
- [ ] Provider 健康检查（定期 ping，自动摘除不健康节点）
- [ ] 配置热更新（无需重启即可生效）
- [ ] 更细粒度的 timeout 控制（connect / read / total）
- [ ] Circuit breaker 模式（熔断器）
