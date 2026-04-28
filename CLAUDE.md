# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 常用命令

### 后端（cwd = `backend/`）
```bash
# 本地运行（默认读取 configs/config.yaml）
go run cmd/server/main.go --config=configs/config.yaml

# 构建 / 清理 / 测试 / 依赖
make build        # 产物：bin/server
make test         # go test -v ./...
make tidy         # go mod tidy
make lint         # golangci-lint run（未装会报错）
make fmt

# 跑单个测试
go test -v -run TestFooBar ./internal/service/...
```

默认本地端口为 `configs/config.yaml` 中的 `server.port`（仓库里默认 **30001**）。Docker 镜像内端口是 **8000**（见 Dockerfile），前端 Nginx 监听 8080 反代 `/api` 到 `127.0.0.1:8000`。

### 前端（cwd = `frontend/`）
```bash
npm install
npm run dev       # Vite dev server @ :3000，代理 /api + ws → http://127.0.0.1:30001
npm run build     # 先 tsc 类型检查，再 vite build
npm run lint      # eslint --max-warnings 0，有告警会失败

# 只做类型检查不产物
npx tsc --noEmit
```

前端与后端的联调端口硬编码在 `frontend/vite.config.ts` 的 proxy 里（目标 `127.0.0.1:30001`）——改后端端口时两边都要动。

### 整体构建
- `docker build -t k8s-manager .`：多阶段，Node 构前端 → Go 构后端 → alpine + nginx 打一个 All-in-One 镜像。
- `build.sh`：CI/发布脚本（打 tag、推多区域镜像仓库），日常开发不用。

## 架构要点（跨文件才能看清的那部分）

### 后端分层
```
cmd/server/main.go
  └── 手工装配所有依赖：Repo → ClientManager → Service → Handler → Router
internal/api/router/router.go
  └── 单文件集中路由定义，所有路径都在这里（包括中间件挂载）
internal/api/handler/*.go       — HTTP 层，只做参数解析和响应
internal/service/*.go           — 业务逻辑，含历史记录、权限前校验、字段聚合
internal/k8s/                   — K8s 相关
  ├── client_manager.go         — 多集群客户端池（clientset + metrics client + rest.Config）
  └── operator/*.go             — 按资源类型封装 client-go，返回原始 k8s 对象或轻量 struct
internal/repository/            — GORM 访问 MySQL
internal/model/
  ├── dto/                      — 对外 JSON 结构
  └── entity/entity.go          — GORM 实体，启动时 AutoMigrate
internal/pkg/
  ├── crypto/                   — AES-256-GCM，加密 kubeconfig
  ├── errors/                   — 统一错误码 + apperrors.Wrap
  └── logger/                   — Zap 封装
```

关键约定：
- **Handler 只调 Service，Service 可调 operator + repository**。operator 层不感知 DTO，也不碰数据库。
- Service 的 `toResponse` 聚合 `k8s.Pod/Deployment/…` + metrics + spec 中的字段，产出前端友好的 DTO。新增资源属性时改两处：operator（拿原始数据）+ service toResponse（拼字段）。
- 有写操作的资源（ConfigMap/Deployment/Service/HPA）在 Apply/Delete 时由 Service 显式调 `historyRepo.Create` 记一笔。Pod 不走历史（Pod 是运行时产物）。

### 多集群与 kubeconfig
- `ClientManager` 是进程内 `map[clusterName]*Clientset` + `map[clusterName]metricsv.Interface`。启动时 `ClusterManageService.LoadAllClusters` 从 MySQL 把所有 `source=database` 的集群解密 kubeconfig 后建连接。
- **kubeconfig 在数据库里用 AES-256-GCM 加密**（`encryption.key` → `internal/pkg/crypto`）。这个 key 一旦换掉，库里已有集群全部无法解密——改 key 要配套做数据迁移或清库重传。
- 运行时增删集群走 `/api/v1/admin/clusters`，Service 负责同步更新 DB 和 ClientManager。
- 旧版本遗留的 `source=config` 集群在启动时由 `MigrateConfigClusters` 自动迁移为 `source=database`。

### 认证与权限
- 登录走飞书 OAuth（`/api/v1/auth/feishu/login`）→ 签发 JWT。
- 中间件组合：`Auth`（解 JWT+取 user）→ `AdminRequired`（角色=admin）/ `ResourcePermission`（查 `user_permissions` 表，命中 cluster + namespace 才放行）。
- `user_permissions.namespaces` 是 JSON 数组，`["*"]` = 所有 namespace。
- **WebSocket exec 路由（`/pods/:name/exec`）不走 Auth 中间件**，单独在 router.go 外层注册，认证通过 URL 参数 `?token=` 在 handler 内做——因为浏览器 WebSocket 不支持自定义 Header。改这条路由时别误加 Auth 中间件。

### 数据库
- MySQL + GORM，启动时 `AutoMigrate` 建/改表。`backend/migrations/*.sql` 仅是参考文档，不在代码路径里执行。新增字段直接在 `entity.go` 加 tag 即可。
- 变更历史表 `resource_histories` 按 `(cluster, namespace, resource_type, resource_name, version)` 维护版本递增，用于 Diff 和 Rollback。

### 前端结构
```
src/
├── api/           — axios 客户端，按资源分模块（podApi / deploymentApi / ...）
├── pages/         — 顶层页面与主业务逻辑
├── components/    — 含 ui/（shadcn/ui）+ 自定义通用组件（PodTerminalDialog / DeleteConfirmDialog ...）
├── hooks/         — use-cluster（全局集群/命名空间选择）、use-toast
├── lib/           — utils（分类/格式化）
└── types/index.ts — 所有 DTO 的 TS 映射，必须与后端 dto 包同步
```
- 集群和命名空间的选中态在 `use-cluster` 里全局共享，页面通过它读 `selectedCluster` / `selectedNamespace`。
- Monaco Editor 与 xterm.js 体积较大，已在 `vite.config.ts` 中单独 chunk + 预打包。

## 扩展模式（新增资源类型）

从底往上大致要改：
1. `backend/internal/k8s/operator/<resource>.go`：用 client-go 暴露 List/Get/Apply/Delete。
2. `backend/internal/model/dto/<resource>.go`（或加到 `resource.go`）：请求/响应结构。
3. `backend/internal/service/<resource>_service.go`：业务编排 + 历史记录（写操作必须调 `historyRepo.Create`）。
4. `backend/internal/api/handler/<resource>_handler.go`：参数解析。
5. `backend/cmd/server/main.go`：实例化并注入。
6. `backend/internal/api/router/router.go`：挂到 `clusterResources` 组下，自动继承 `Auth + ResourcePermission` 中间件。
7. 前端：`types/index.ts` + `api/` + 新 page 组件 + `App.tsx` 路由。

## 容易踩的坑

- **metrics-server 未装的集群**：`ClientManager.GetMetricsClient` 会返回 nil，`PodService.List` 里必须容错（已有实现），新增依赖 metrics 的字段也要做 nil 兜底。
- **Pod 的 limits/requests 在 `pod.Spec.Containers`，而状态和 metrics 在 `pod.Status.ContainerStatuses` / metrics API**：按 `Name` 匹配合并，三份来源任一可能有缺失。
- **GORM AutoMigrate 不删字段也不改类型**：破坏性变更要手工 DDL 或清表重建，不能只改 entity。
- **配置加密 key 和 JWT secret 强烈建议通过环境变量注入**（`ENCRYPTION_KEY` / `JWT_SECRET`），不要把生产 key 提交到 `configs/config.yaml`（当前仓库里的是开发占位值）。
