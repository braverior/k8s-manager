# K8s-Manager

一个现代化的 Kubernetes 多集群管理平台，提供直观的 Web 界面来管理和监控 Kubernetes 资源。
<img width="2998" height="1534" alt="image" src="https://github.com/user-attachments/assets/c068a936-0bed-4518-9645-4fd1b0c27639" />
Nodes:
<img width="2956" height="1420" alt="image" src="https://github.com/user-attachments/assets/0dd4183f-9b0c-420f-947e-15b6945ab504" />
Configmaps:
<img width="2978" height="1428" alt="image" src="https://github.com/user-attachments/assets/ffed82b1-becf-47d2-b6da-213f8d3f56c7" />
Deployments:
<img width="2962" height="1398" alt="image" src="https://github.com/user-attachments/assets/29a3941a-94c2-46b8-881c-cb9567dca9b3" />
<img width="2932" height="1528" alt="image" src="https://github.com/user-attachments/assets/a9614a7d-1eda-4a42-af60-2cdd87b08c39" />

HPAs：
<img width="2926" height="1412" alt="image" src="https://github.com/user-attachments/assets/ef38906d-e64c-4c32-9a6d-b5fb1f68d66e" />
Change History：
<img width="2962" height="1456" alt="image" src="https://github.com/user-attachments/assets/c47c4211-69b4-4c7a-b337-6b370c439cea" />



## 功能特性

- **多集群管理** - 支持同时管理多个 Kubernetes 集群（in-cluster 和 kubeconfig 模式）
- **资源管理** - Deployment、Pod、ConfigMap、Service、HPA 的 CRUD 操作
- **实时监控** - Dashboard 展示集群资源使用情况和节点状态
- **在线编辑** - 内置 Monaco Editor，支持 YAML 在线编辑
- **变更历史** - 记录资源变更历史，支持审计追溯
- **用户认证** - 支持飞书 OAuth 登录 + JWT Token 认证
- **权限控制** - 基于角色的访问控制（RBAC）

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

## 项目结构

```
k8s-manager/
├── backend/                          # 后端服务
│   ├── cmd/server/main.go           # 程序入口
│   ├── configs/config.yaml          # 配置模板
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handler/             # HTTP 处理器
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

## RBAC 权限说明

应用需要以下 Kubernetes 权限：

| 资源 | 权限 | 说明 |
|------|------|------|
| namespaces | get, list | 列出命名空间 |
| nodes | get, list | 查看节点信息 |
| pods | get, list, delete | Pod 管理 |
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

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| POST | `/api/v1/auth/feishu/login` | 飞书登录 |
| GET | `/api/v1/clusters` | 获取集群列表 |
| GET | `/api/v1/clusters/:cluster/namespaces` | 获取命名空间列表 |
| GET | `/api/v1/clusters/:cluster/deployments` | 获取 Deployment 列表 |
| GET | `/api/v1/clusters/:cluster/pods` | 获取 Pod 列表 |
| GET | `/api/v1/clusters/:cluster/configmaps` | 获取 ConfigMap 列表 |
| GET | `/api/v1/clusters/:cluster/services` | 获取 Service 列表 |
| GET | `/api/v1/clusters/:cluster/hpa` | 获取 HPA 列表 |
| GET | `/api/v1/clusters/:cluster/dashboard` | 获取 Dashboard 数据 |
| GET | `/api/v1/clusters/:cluster/namespaces/:namespace/histories/:id/diff-previous` | 与上一版本 Diff 对比 |

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
