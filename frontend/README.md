# Frontend for LiteLLM Gateway

基于 Vue 3 的管理后台前端。

## 技术栈

- Vue 3 + TypeScript
- Vite (构建工具)
- Pinia (状态管理)
- Vue Router (路由)
- Axios (HTTP 客户端)
- TailwindCSS (样式)

## 开发环境

```bash
# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev

# 构建生产版本
pnpm build

# 预览生产构建
pnpm preview
```

## 配置说明

### 环境变量

创建 `.env` 文件：

```bash
# Gateway API Key (必需)
VITE_GATEWAY_API_KEY=sk-your-api-key-here

# Gateway 地址 (可选，默认使用代理)
VITE_GATEWAY_URL=http://localhost:8080/api
```

### Vite 配置

默认配置会将 `/api` 请求代理到 `http://localhost:8080`。如果后端运行在其他地址，请修改 `vite.config.ts`。

## 项目结构

```
frontend/
├── src/
│   ├── api/              # API 客户端和类型定义
│   ├── stores/           # Pinia 状态管理
│   ├── views/            # 页面组件
│   ├── components/       # 共享组件
│   ├── router/           # 路由配置
│   ├── styles/           # 全局样式
│   ├── main.ts           # 应用入口
│   └── App.vue           # 根组件
├── index.html
├── package.json
├── tsconfig.json
└── vite.config.ts
```

## 功能特性

### 已实现

- [x] 用户认证（API Key）
- [x] Dashboard 仪表盘
- [x] Models 管理（只读）
- [x] Providers 管理（只读）
- [x] Deployments CRUD
  - 列表展示
  - 创建
  - 编辑
  - 删除
  - 启用/禁用切换

### 待实现

- [ ] Routing Rules 管理
- [ ] Prometheus Metrics 展示
- [ ] Usage Logs 查看
- [ ] 实时刷新

## 开发指南

### 添加新页面

1. 在 `src/views/` 创建组件
2. 在 `src/router/index.ts` 添加路由
3. 在 `src/components/layout/Sidebar.vue` 添加导航链接

### API 调用

使用 `src/api/client.ts` 中封装的 API 函数：

```typescript
import { deploymentsApi } from '@/api/client'

// 获取列表
const response = await deploymentsApi.list()

// 创建
await deploymentsApi.create(data)

// 更新
await deploymentsApi.update(id, data)

// 删除
await deploymentsApi.delete(id)
```
