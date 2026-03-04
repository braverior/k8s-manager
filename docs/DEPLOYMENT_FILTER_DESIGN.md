# Pod 页面 Deployment 标签筛选功能设计文档

## 1. 功能背景

### 当前问题

当前 Pod 页面已支持通过 URL 参数 `?deployment=xxx` 按 Deployment 筛选 Pod，但该功能仅在从 DeploymentsPage 点击"查看 Pods"时触发（`DeploymentsPage.tsx:323`）。用户在 Pod 页面内无法直观地了解当前 namespace 下有哪些 Deployment，也无法快速切换筛选条件。

### 需求描述

在 Pod 页面直接展示 Deployment 标签栏，用户点击标签即可快速按 Deployment 筛选 Pod，提升日常运维操作效率。

## 2. 功能设计

### 核心功能

- 在 Pod 页面搜索栏**上方**增加 Deployment 标签栏
- 标签栏展示当前 namespace 下所有 Deployment 名称
- 包含一个"全部"标签用于清除筛选，显示所有 Pod
- 每个 Deployment 标签显示对应的 Pod 数量（`pod_count` 字段）
- 点击标签通过已有的 `?deployment=xxx` URL 参数机制进行筛选
- 当前激活的标签高亮显示（使用 `default` variant），未激活标签使用 `outline` variant
- 移除原有的 "Filtered by Deployment: xxx" 提示条（标签栏已替代其功能）

### 交互逻辑

1. 页面加载时，同时请求 Pod 列表和 Deployment 列表
2. 若 URL 中无 `deployment` 参数，"全部"标签高亮
3. 点击某个 Deployment 标签 → 设置 `?deployment=xxx` → 触发 Pod 列表刷新
4. 点击"全部"标签 → 清除 `deployment` 参数 → 显示所有 Pod
5. 切换 namespace 时，自动重新加载 Deployment 列表

## 3. 技术方案

### 前端改动（`frontend/src/pages/PodsPage.tsx`）

1. **新增导入**：引入 `deploymentApi` 和 `Deployment` 类型
2. **新增状态**：`deployments: Deployment[]` 和 `deploymentsLoading: boolean`
3. **新增请求**：在 `useEffect` 中调用 `deploymentApi.list()` 获取 Deployment 列表
4. **新增标签栏 UI**：在搜索栏上方渲染 Deployment 标签
5. **移除旧提示条**：删除 "Filtered by Deployment" 的 Badge 提示

### 后端改动

**无需改动。** 后端已有完整支持：

- `GET /clusters/:cluster/namespaces/:namespace/deployments` — 获取 Deployment 列表
- `GET /clusters/:cluster/namespaces/:namespace/pods?deployment=xxx` — 按 Deployment 筛选 Pod
- `Deployment` 类型已包含 `pod_count` 字段

### 复用的现有资源

| 资源 | 来源 |
|------|------|
| `deploymentApi.list()` | `frontend/src/api/index.ts:224` |
| `Deployment` 类型（含 `pod_count`） | `frontend/src/types/index.ts:87` |
| `Badge` 组件 | `@/components/ui/badge` |
| URL 参数筛选机制 | `PodsPage.tsx:52-53`（已有 `useSearchParams`） |
| `clearDeploymentFilter()` | `PodsPage.tsx:98-101` |

## 4. UI 设计

```
┌──────────────────────────────────────────────────────────────────┐
│  🔲 Pods                                                        │
│  View and manage pods in {namespace}                  [12 pods] │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │ [全部(12)] [nginx(4)] [api-server(3)] [worker(5)]       │    │
│  │  ^^^^^^^^                                                │    │
│  │  高亮激活                                                │    │
│  └──────────────────────────────────────────────────────────┘    │
│                                                                  │
│  🔍 [Search Pods...                 ]  [▦][≡]  [↻]             │
│                                                                  │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐    │
│  │ Pod Card 1      │ │ Pod Card 2      │ │ Pod Card 3      │    │
│  │                 │ │                 │ │                 │    │
│  └─────────────────┘ └─────────────────┘ └─────────────────┘    │
└──────────────────────────────────────────────────────────────────┘
```

选中某个 Deployment 后：

```
│  ┌──────────────────────────────────────────────────────────┐    │
│  │ [全部(12)] [nginx(4)] [api-server(3)] [worker(5)]       │    │
│  │            ^^^^^^^^^^                                    │    │
│  │            高亮激活                                      │    │
│  └──────────────────────────────────────────────────────────┘    │
```

### 标签样式

- **激活状态**：`Badge variant="default"` — 实心背景，白色文字
- **非激活状态**：`Badge variant="outline"` — 边框样式，可点击
- **悬停效果**：`cursor-pointer` + `hover:bg-accent`

## 5. 涉及的文件变更

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `docs/DEPLOYMENT_FILTER_DESIGN.md` | 新建 | 本设计文档 |
| `frontend/src/pages/PodsPage.tsx` | 修改 | 增加 Deployment 标签栏 |

### 不涉及的变更

- 后端代码 — 无需修改
- `frontend/src/api/index.ts` — 无需修改，已有 `deploymentApi.list()`
- `frontend/src/types/index.ts` — 无需修改，已有 `Deployment` 接口
- 新增组件 — 标签栏直接内联在 PodsPage 中，无需单独抽组件
