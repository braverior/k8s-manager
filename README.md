# K8s-Manager

一个现代化的 Kubernetes 多集群管理平台，提供直观的 Web 界面来管理和监控 Kubernetes 资源。

## 界面预览

**Dashboard：**
<img width="2998" height="1534" alt="Dashboard" src="https://github.com/user-attachments/assets/c068a936-0bed-4518-9645-4fd1b0c27639" />

**Nodes：**
<img width="2956" height="1420" alt="Nodes" src="https://github.com/user-attachments/assets/0dd4183f-9b0c-420f-947e-15b6945ab504" />

**ConfigMaps：**
<img width="2978" height="1428" alt="ConfigMaps" src="https://github.com/user-attachments/assets/ffed82b1-becf-47d2-b6da-213f8d3f56c7" />

**Deployments：**
<img width="2962" height="1398" alt="Deployments" src="https://github.com/user-attachments/assets/29a3941a-94c2-46b8-881c-cb9567dca9b3" />
<img width="2932" height="1528" alt="Deployment Detail" src="https://github.com/user-attachments/assets/a9614a7d-1eda-4a42-af60-2cdd87b08c39" />

**HPAs：**
<img width="2926" height="1412" alt="HPAs" src="https://github.com/user-attachments/assets/ef38906d-e64c-4c32-9a6d-b5fb1f68d66e" />

**Change History：**
<img width="2962" height="1456" alt="Change History" src="https://github.com/user-attachments/assets/c47c4211-69b4-4c7a-b337-6b370c439cea" />

## 功能特性

- **多集群管理** - 支持同时管理多个 Kubernetes 集群（in-cluster 和 kubeconfig 模式）
- **资源管理** - Deployment、Pod、ConfigMap、Service、HPA 的 CRUD 操作
- **实时监控** - Dashboard 展示集群资源使用情况和节点状态
- **在线编辑** - 内置 Monaco Editor，支持 YAML 在线编辑
- **变更历史** - 记录资源变更历史，支持版本对比（Diff）和一键回滚
- **Pod 运维** - 支持查看 Pod 日志、事件，WebSocket 终端（exec）直接进入容器
- **Deployment 关联** - 支持基于 Deployment 筛选关联的 Pod
- **用户认证** - 支持飞书 OAuth 登录 + JWT Token 认证
- **权限控制** - 基于角色的访问控制（RBAC），支持集群和命名空间级别的精细化权限管理

## 技术栈

### 后端
| 技术 | 版本 | 说明 |
|------|------|------|
| Go | 1.21 | 主要开发语言 |
| Gin | 1.9 | Web 框架 |
| GORM | 1.25 | ORM 框架 |
| client-go | 0.29 | Kubernetes 客户端 |
| Zap | 1.26 | 日志库 |
| Viper | 1.18 | 配置管理 |

### 前端
| 技术 | 版本 | 说明 |
|------|------|------|
| React | 18 | UI 框架 |
| TypeScript | 5.3 | 类型安全 |
| Vite | 5.0 | 构建工具 |
| Tailwind CSS | 3.4 | 样式框架 |
| shadcn/ui | - | 组件库 |
| Monaco Editor | 4.6 | 代码编辑器 |
| xterm.js | 6.0 | Web 终端（Pod exec） |

## 项目结构

```
k8s-manager/
├── backend/                          # 后端服务
│   ├── cmd/server/main.go           # 程序入口
│   ├── configs/config.yaml          # 配置模板
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handler/             # HTTP 处理器
│   │   │   ├── middleware/          # 中间件（认证、CORS、日志等）
│   │   │   └── router/              # 路由定义
│   │   ├── service/                 # 业务逻辑层
│   │   ├── repository/              # 数据访问层
│   │   ├── k8s/                     # K8s 客户端管理
│   │   ├── model/                   # 数据模型
│   │   ├── config/                  # 配置加载
│   │   └── pkg/logger/              # 日志工具
│   ├── migrations/                  # 数据库迁移
│   ├── deploy/                      # K8s 部署配置
│   ├── Dockerfile                   # 后端镜像构建
│   └── Makefile                     # 构建脚本
│
├── frontend/                         # 前端应用
│   ├── src/
│   │   ├── main.tsx                 # 入口文件
│   │   ├── App.tsx                  # 根组件
│   │   ├── api/                     # API 客户端
│   │   ├── components/              # 公共组件
│   │   ├── pages/                   # 页面组件
│   │   ├── hooks/                   # 自定义 Hooks
│   │   ├── lib/                     # 工具函数
│   │   └── types/                   # TypeScript 类型
│   ├── Dockerfile                   # 前端镜像构建
│   ├── nginx.conf                   # Nginx 配置
│   └── vite.config.ts               # Vite 配置
│
└── docs/                             # 文档
    ├── API_DESIGN.md                # API 接口文档
    └── DEPLOYMENT_FILTER_DESIGN.md  # Deployment 筛选设计文档
```

## 架构图

```
                                    ┌─────────────────────────────────────┐
                                    │          Kubernetes Cluster         │
                                    │  ┌─────────────────────────────────┐│
┌──────────┐    ┌──────────────┐   │  │         k8s-manager namespace   ││
│  Browser │───▶│   Ingress/   │───┼─▶│  ┌─────────┐    ┌─────────────┐ ││
│          │    │   NodePort   │   │  │  │Frontend │    │   Backend   │ ││
└──────────┘    └──────────────┘   │  │  │ (Nginx) │───▶│  (Go API)   │ ││
                                    │  │  └─────────┘    └──────┬──────┘ ││
                                    │  └───────────────────────│────────┘│
                                    │                          │         │
                                    │              ┌───────────▼───────┐ │
                                    │              │  K8s API Server   │ │
                                    │              └───────────────────┘ │
                                    └─────────────────────────────────────┘
                                                       │
                                               ┌───────▼───────┐
                                               │     MySQL     │
                                               │  (外部数据库)  │
                                               └───────────────┘
```

## 快速开始

### 前置要求

- Docker 20.10+
- Kubernetes 1.25+
- MySQL 8.0+
- kubectl 已配置

### 本地开发

```bash
# 1. 启动后端
cd backend
cp configs/config.yaml configs/config.local.yaml  # 复制并修改配置
go run cmd/server/main.go --config=configs/config.local.yaml

# 2. 启动前端
cd frontend
npm install
npm run dev
```

访问 http://localhost:3000

## 部署指南

### 方式一：Docker 镜像构建

#### 构建后端镜像

```bash
cd backend
docker build -t k8s-manager-backend:latest .
```

**Dockerfile 说明（多阶段构建）：**
```dockerfile
# 构建阶段 - 编译 Go 二进制
FROM golang:1.21-alpine AS builder
# 运行阶段 - 最小化镜像
FROM alpine:3.19
```

#### 构建前端镜像

```bash
cd frontend
docker build -t k8s-manager-frontend:latest .
```

**Dockerfile 说明：**
```dockerfile
# 构建阶段 - 编译前端资源
FROM node:20-alpine AS builder
# 运行阶段 - Nginx 托管静态文件
FROM nginx:alpine
```

### 方式二：Kubernetes 部署

项目提供了完整的 Kubernetes 部署配置，位于 `backend/deploy/` 目录。

#### 部署文件说明

| 文件 | 说明 |
|------|------|
| `namespace.yaml` | 创建 k8s-manager 命名空间 |
| `serviceaccount.yaml` | ServiceAccount 配置 |
| `rbac.yaml` | ClusterRole 和 ClusterRoleBinding |
| `configmap.yaml` | 应用配置（需要修改） |
| `deployment.yaml` | Deployment 配置 |
| `service.yaml` | Service 配置（NodePort: 30001） |
| `kustomization.yaml` | Kustomize 配置 |

#### 部署步骤

```bash
# 1. 修改配置
vim backend/deploy/configmap.yaml
# 更新以下配置：
# - database: 数据库连接信息
# - feishu: 飞书应用配置
# - jwt.secret: JWT 密钥（生产环境必须修改）

# 2. 构建并推送镜像
docker build -t your-registry/k8s-manager-backend:v1.0.0 ./backend
docker build -t your-registry/k8s-manager-frontend:v1.0.0 ./frontend
docker push your-registry/k8s-manager-backend:v1.0.0
docker push your-registry/k8s-manager-frontend:v1.0.0

# 3. 更新 deployment.yaml 中的镜像地址
vim backend/deploy/deployment.yaml

# 4. 使用 Kustomize 部署
cd backend/deploy
kubectl apply -k .

# 或逐个部署
kubectl apply -f namespace.yaml
kubectl apply -f serviceaccount.yaml
kubectl apply -f rbac.yaml
kubectl apply -f configmap.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml

# 5. 验证部署
kubectl get pods -n k8s-manager
kubectl get svc -n k8s-manager
kubectl logs -n k8s-manager -l app=k8s-manager
```

#### 访问应用

```bash
# NodePort 方式
http://<node-ip>:30001

# 或配置 Ingress
```

### 方式三：单 Pod 部署（All-in-One）

如果需要在单个 Pod 中同时运行前后端，可以使用以下合并的 Dockerfile：

```dockerfile
# ============================================
# All-in-One Dockerfile
# 前后端合并到单个镜像
# ============================================

# Stage 1: 构建前端
FROM node:20-alpine AS frontend-builder
WORKDIR /frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: 构建后端
FROM golang:1.21-alpine AS backend-builder
WORKDIR /backend
ENV GOPROXY=https://goproxy.cn,direct
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o k8s-manager ./cmd/server

# Stage 3: 最终运行镜像
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata nginx supervisor

WORKDIR /app

# 复制后端二进制
COPY --from=backend-builder /backend/k8s-manager .

# 复制前端静态文件
COPY --from=frontend-builder /frontend/dist /usr/share/nginx/html
COPY frontend/nginx.conf /etc/nginx/http.d/default.conf

# 创建 supervisord 配置
RUN mkdir -p /etc/supervisor.d
COPY <<EOF /etc/supervisor.d/app.ini
[supervisord]
nodaemon=true

[program:nginx]
command=nginx -g "daemon off;"
autostart=true
autorestart=true

[program:backend]
command=/app/k8s-manager --config=/etc/k8s-manager/config.yaml
autostart=true
autorestart=true
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0
EOF

# 创建非 root 用户
RUN adduser -D -g '' appuser && \
    chown -R appuser:appuser /app /var/lib/nginx /var/log/nginx

EXPOSE 80 8080

CMD ["supervisord", "-c", "/etc/supervisord.conf"]
```

**构建和运行：**
```bash
# 在项目根目录构建
docker build -f Dockerfile.all-in-one -t k8s-manager:latest .

# 运行
docker run -d \
  -p 80:80 \
  -p 8080:8080 \
  -v /path/to/config.yaml:/etc/k8s-manager/config.yaml \
  k8s-manager:latest
```

### 方式四：Sidecar 模式（单 Pod 双容器）

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-manager
  namespace: k8s-manager
spec:
  replicas: 1
  selector:
    matchLabels:
      app: k8s-manager
  template:
    metadata:
      labels:
        app: k8s-manager
    spec:
      serviceAccountName: k8s-manager
      containers:
        # 前端容器
        - name: frontend
          image: k8s-manager-frontend:latest
          ports:
            - name: http
              containerPort: 80
          resources:
            requests:
              cpu: 50m
              memory: 64Mi
            limits:
              cpu: 200m
              memory: 128Mi

        # 后端容器
        - name: backend
          image: k8s-manager-backend:latest
          ports:
            - name: api
              containerPort: 8080
          args:
            - --config=/etc/k8s-manager/config.yaml
          volumeMounts:
            - name: config
              mountPath: /etc/k8s-manager
              readOnly: true
          livenessProbe:
            httpGet:
              path: /health
              port: api
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /health
              port: api
            initialDelaySeconds: 5
            periodSeconds: 5
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi

      volumes:
        - name: config
          configMap:
            name: k8s-manager-config
```

## 配置说明

### 配置文件示例 (`config.yaml`)

```yaml
# 服务器配置
server:
  port: 8080
  mode: release  # debug, release, test

# 数据库配置
database:
  host: mysql.example.com
  port: 3306
  username: root
  password: your-password
  database: k8s_manager
  charset: utf8mb4
  max_idle_conns: 10
  max_open_conns: 100

# 日志配置
log:
  level: info     # debug, info, warn, error
  format: json    # json, console
  output: stdout  # stdout, file
  file_path: ""

# 集群配置
clusters:
  # 当前集群（部署在 K8s 中时使用 ServiceAccount）
  - name: local
    description: 当前集群
    type: in-cluster

  # 外部集群示例
  # - name: production
  #   description: 生产环境集群
  #   type: kubeconfig
  #   kubeconfig_path: /etc/k8s-manager/kubeconfigs/prod.yaml

# 飞书 OAuth 配置
feishu:
  app_id: "your-app-id"
  app_secret: "your-app-secret"
  redirect_uri: "https://your-domain.com/auth/feishu/callback"
  authorize_url: "https://passport.feishu.cn/suite/passport/oauth/authorize"

# JWT 配置
jwt:
  secret: "your-jwt-secret-please-change-in-production"
  expire_time: 86400  # 24小时
  issuer: "k8s-manager"
```

### 环境变量覆盖

支持通过环境变量覆盖配置：

```bash
export DATABASE_HOST=mysql.example.com
export DATABASE_PASSWORD=secret
export JWT_SECRET=your-super-secret-key
```

## 数据库初始化

应用使用 MySQL 存储用户信息、权限配置和资源变更历史。需要按顺序执行以下迁移脚本：

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS k8s_manager DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 执行迁移（按顺序）
mysql -u root -p k8s_manager < backend/migrations/001_create_resource_histories_table.sql
mysql -u root -p k8s_manager < backend/migrations/002_create_users_table.sql
mysql -u root -p k8s_manager < backend/migrations/003_create_user_permissions_table.sql
mysql -u root -p k8s_manager < backend/migrations/004_init_admin_user.sql
```

## RBAC 权限说明

应用需要以下 Kubernetes 权限：

| 资源 | 权限 | 说明 |
|------|------|------|
| namespaces | get, list | 列出命名空间 |
| nodes | get, list | 查看节点信息 |
| pods | get, list, delete | Pod 管理 |
| pods/log | get | 查看 Pod 日志 |
| pods/exec | create, get | Pod 终端（exec） |
| events | get, list | 查看 Pod 事件 |
| pods (metrics) | get, list | Pod 监控指标 |
| nodes (metrics) | get, list | 节点监控指标 |
| configmaps | get, list, create, update, patch, delete | ConfigMap 完整管理 |
| services | get, list, create, update, patch, delete | Service 完整管理 |
| deployments | get, list, create, update, patch, delete | Deployment 完整管理 |
| horizontalpodautoscalers | get, list, create, update, patch, delete | HPA 完整管理 |

## 健康检查

后端提供健康检查端点：

```bash
# 健康检查
curl http://localhost:8080/health
```

## API 端点

详细的 API 文档请参考 [API 接口文档](docs/API_DESIGN.md)。

### 认证
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/auth/feishu/config` | 获取飞书 OAuth 配置 |
| POST | `/api/v1/auth/feishu/login` | 飞书 OAuth 登录 |
| GET | `/api/v1/auth/me` | 获取当前用户信息 |
| POST | `/api/v1/auth/logout` | 退出登录 |

### 集群与节点
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/api/v1/clusters` | 获取集群列表 |
| GET | `/api/v1/clusters/:cluster` | 获取集群详情 |
| POST | `/api/v1/clusters/:cluster/test-connection` | 测试集群连接 |
| GET | `/api/v1/clusters/:cluster/namespaces` | 获取命名空间列表 |
| GET | `/api/v1/clusters/:cluster/dashboard` | 获取 Dashboard 数据 |
| GET | `/api/v1/clusters/:cluster/nodes` | 获取节点列表 |
| GET | `/api/v1/clusters/:cluster/nodes/:name` | 获取节点详情 |

### 资源管理（以下路径前缀：`/api/v1/clusters/:cluster/namespaces/:namespace`）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/configmaps` | 列表 / 创建 ConfigMap |
| GET/PUT/DELETE | `/configmaps/:name` | 获取 / 更新 / 删除 ConfigMap |
| GET/POST | `/deployments` | 列表 / 创建 Deployment |
| GET/PUT/DELETE | `/deployments/:name` | 获取 / 更新 / 删除 Deployment |
| GET | `/deployments/:name/pods` | 获取 Deployment 关联的 Pod |
| GET/POST | `/services` | 列表 / 创建 Service |
| GET/PUT/DELETE | `/services/:name` | 获取 / 更新 / 删除 Service |
| GET/POST | `/hpas` | 列表 / 创建 HPA |
| GET/PUT/DELETE | `/hpas/:name` | 获取 / 更新 / 删除 HPA |

### Pod 管理（以下路径前缀：`/api/v1/clusters/:cluster/namespaces/:namespace`）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/pods` | 获取 Pod 列表（支持 `?deployment=` 筛选） |
| GET | `/pods/:name` | 获取 Pod 详情 |
| DELETE | `/pods/:name` | 删除 Pod（触发重启） |
| GET | `/pods/:name/containers` | 获取 Pod 容器列表 |
| GET | `/pods/:name/logs` | 获取 Pod 日志 |
| GET | `/pods/:name/events` | 获取 Pod 事件 |
| GET | `/pods/:name/exec` | WebSocket 终端（进入容器） |

### 变更历史（以下路径前缀：`/api/v1/clusters/:cluster/namespaces/:namespace`）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/histories` | 获取变更历史列表 |
| GET | `/histories/:id` | 获取历史记录详情 |
| GET | `/histories/diff` | 两个版本之间的 Diff 对比 |
| GET | `/histories/:id/diff-previous` | 与上一版本 Diff 对比 |
| POST | `/histories/:id/rollback` | 回滚到指定版本 |

### 用户管理（仅管理员）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/users` | 获取用户列表 |
| GET | `/api/v1/admin/users/:user_id` | 获取用户详情 |
| PUT | `/api/v1/admin/users/:user_id/role` | 修改用户角色 |
| PUT | `/api/v1/admin/users/:user_id/status` | 修改用户状态 |
| GET | `/api/v1/admin/users/:user_id/permissions` | 获取用户权限 |
| PUT | `/api/v1/admin/users/:user_id/permissions` | 设置用户权限 |
| POST | `/api/v1/admin/permissions/batch` | 批量设置权限 |

## 飞书应用配置

本项目使用飞书 OAuth 2.0 实现用户登录认证。部署前需要在[飞书开放平台](https://open.feishu.cn/)创建应用并完成配置。

### 创建飞书应用

1. 登录[飞书开放平台](https://open.feishu.cn/)，进入「开发者后台」
2. 点击「创建企业自建应用」
3. 填写应用名称（如 `K8s-Manager`）和描述
4. 记录 `App ID` 和 `App Secret`，填入配置文件

### 开通应用能力

在应用的「添加应用能力」中，开通以下能力：

| 能力 | 说明 |
|------|------|
| **网页应用** | 必须开通，用于 OAuth 网页登录 |

### 配置安全设置

在应用的「安全设置」中：

- **重定向 URL**：添加 OAuth 回调地址，与配置文件中的 `redirect_uri` 保持一致
  - 本地开发：`http://localhost:3000/auth/feishu/callback`
  - 生产环境：`https://your-domain.com/auth/feishu/callback`

### 申请权限范围

在应用的「权限管理」中，申请以下 OAuth 权限范围（Scope）：

| 权限范围 | 权限名称 | 是否必须 | 说明 |
|----------|----------|----------|------|
| `contact:user.base:readonly` | 获取用户基本信息 | **必须** | 获取用户 open_id、名称、头像 |
| `contact:user.email:readonly` | 获取用户邮箱 | 推荐 | 获取用户邮箱地址 |
| `contact:user.phone:readonly` | 获取用户手机号 | 可选 | 获取用户手机号 |
| `contact:user.employee_id:readonly` | 获取用户工号 | 可选 | 获取用户工号信息 |

### 发布应用

1. 在「版本管理与发布」中创建版本
2. 设置可用范围（全部员工 或 指定部门/人员）
3. 提交审核并发布

> **注意**：应用未发布前，仅应用创建者可以登录测试。

### 登录流程说明

```
┌──────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────┐
│  Browser │     │   Frontend   │     │   Backend    │     │  飞书 API │
└────┬─────┘     └──────┬───────┘     └──────┬───────┘     └────┬─────┘
     │  1. 点击登录      │                    │                  │
     │──────────────────▶│                    │                  │
     │                   │  2. 获取飞书配置    │                  │
     │                   │───────────────────▶│                  │
     │                   │  3. 返回 AppID 等   │                  │
     │                   │◀───────────────────│                  │
     │  4. 重定向到飞书授权页                   │                  │
     │◀──────────────────│                    │                  │
     │  5. 用户在飞书授权                       │                  │
     │─────────────────────────────────────────────────────────▶│
     │  6. 飞书回调携带 code                    │                  │
     │◀─────────────────────────────────────────────────────────│
     │  7. 发送 code 到后端                    │                  │
     │──────────────────▶│───────────────────▶│                  │
     │                   │                    │  8. code 换 token │
     │                   │                    │─────────────────▶│
     │                   │                    │  9. 获取用户信息   │
     │                   │                    │─────────────────▶│
     │                   │                    │  10. 返回用户信息  │
     │                   │                    │◀─────────────────│
     │                   │  11. 创建/更新用户，签发 JWT            │
     │                   │◀───────────────────│                  │
     │  12. 返回 JWT Token + 用户信息          │                  │
     │◀──────────────────│                    │                  │
```

**流程说明：**
1. 前端获取飞书 OAuth 配置（`App ID`、`redirect_uri`、`authorize_url`）
2. 构建授权 URL 并重定向用户到飞书登录页
3. 用户在飞书完成授权后，飞书回调 `redirect_uri` 并携带 `authorization_code`
4. 前端将 `code` 发送到后端 `/api/v1/auth/feishu/login`
5. 后端用 `code` 向飞书换取 `user_access_token`
6. 后端用 `token` 获取用户信息（open_id、姓名、头像、邮箱等）
7. 后端在数据库中创建或更新用户记录，签发 JWT Token
8. 前端存储 JWT Token，后续请求通过 `Authorization: Bearer <token>` 进行认证

### 用户信息存储

通过飞书 OAuth 获取的用户信息将存储到数据库：

| 字段 | 来源 | 说明 |
|------|------|------|
| `id` | `open_id` | 飞书用户唯一标识，作为主键 |
| `name` | `name` | 用户姓名 |
| `email` | `email` | 用户邮箱（需申请权限） |
| `mobile` | `mobile` | 手机号（需申请权限） |
| `avatar_url` | `picture` | 用户头像 URL |
| `employee_id` | `employee_no` | 工号（需申请权限） |
| `role` | - | 系统角色：`admin` / `user`，默认 `user` |
| `status` | - | 用户状态：`active` / `disabled`，默认 `active` |

### 配置示例

```yaml
feishu:
  app_id: "cli_xxxxxxxxxx"          # 飞书应用 App ID
  app_secret: "xxxxxxxxxxxxxxxx"     # 飞书应用 App Secret
  redirect_uri: "https://your-domain.com/auth/feishu/callback"  # OAuth 回调地址
  authorize_url: "https://passport.feishu.cn/suite/passport/oauth/authorize"  # 飞书授权地址（通常无需修改）
```

支持通过环境变量覆盖：

```bash
export FEISHU_APP_ID=cli_xxxxxxxxxx
export FEISHU_APP_SECRET=xxxxxxxxxxxxxxxx
export FEISHU_REDIRECT_URI=https://your-domain.com/auth/feishu/callback
```

## 常见问题

### Q: Pod 无法启动，提示权限不足？
A: 检查 ServiceAccount 和 RBAC 配置是否正确部署：
```bash
kubectl get clusterrole k8s-manager
kubectl get clusterrolebinding k8s-manager
```

### Q: 无法连接数据库？
A: 确保：
1. 数据库地址从 Pod 内可访问
2. 数据库用户有正确权限
3. 数据库已创建

### Q: 飞书登录失败？
A: 检查：
1. 飞书应用配置正确
2. redirect_uri 与飞书后台配置一致
3. 应用已发布

## License

MIT License
