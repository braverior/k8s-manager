# K8s-Manager

一个现代化的 Kubernetes 多集群管理平台，提供直观的 Web 界面来管理和监控 Kubernetes 资源。

采用 All-in-One 部署模式，前后端打包到单个 Docker 容器中，通过页面上传 kubeconfig 即可动态管理多个 Kubernetes 集群，无需在被管理集群上部署任何组件。

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

- **多集群管理** - 通过页面上传 kubeconfig 动态添加/删除集群，kubeconfig AES-256-GCM 加密存储
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
| Go | 1.23 | 主要开发语言 |
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

## 架构图

```
┌──────────┐         ┌──────────────────────────────┐
│  Browser │────────▶│     K8s-Manager Container    │
└──────────┘         │                              │
                     │  ┌────────┐    ┌───────────┐ │
                     │  │ Nginx  │───▶│  Backend  │ │
                     │  │ (:80)  │    │  (Go API) │ │
                     │  └────────┘    └─────┬─────┘ │
                     └──────────────────────│───────┘
                                            │
                     ┌──────────────────────┼──────────────────┐
                     │                      │                  │
              ┌──────▼──────┐  ┌────────────▼─────┐  ┌───────▼───────┐
              │    MySQL    │  │  K8s Cluster A   │  │ K8s Cluster B │
              │  (数据存储)  │  │  (kubeconfig)    │  │ (kubeconfig)  │
              └─────────────┘  └──────────────────┘  └───────────────┘
```

> 通过管理页面上传 kubeconfig 即可动态添加远程集群，无需在被管理集群上部署任何组件。

## 项目结构

```
k8s-manager/
├── Dockerfile                        # All-in-One 镜像构建
├── docker-compose.yml                # Docker Compose 部署配置
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
│   │   └── pkg/
│   │       ├── logger/              # 日志工具
│   │       ├── errors/              # 错误码定义
│   │       └── crypto/              # AES 加密工具
│   ├── migrations/                  # 数据库迁移（参考用）
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
│   ├── nginx.conf                   # Nginx 配置（含 API 反向代理）
│   └── vite.config.ts               # Vite 配置
│
└── docs/                             # 文档
    ├── API_DESIGN.md                # API 接口文档
    └── DEPLOYMENT_FILTER_DESIGN.md  # Deployment 筛选设计文档
```

## 快速开始

### 前置要求

- Docker 20.10+（生产部署）
- MySQL 8.0+
- Node.js 20+（本地开发）
- Go 1.23+（本地开发）

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

### Docker 部署（推荐）

```bash
# 1. 构建镜像
docker build -t k8s-manager:latest .

# 2. 准备配置文件
cp backend/configs/config.yaml config.yaml
# 编辑 config.yaml，修改数据库连接、飞书配置等

# 3. 运行容器
docker run -d \
  --name k8s-manager \
  -p 80:80 \
  -v $(pwd)/config.yaml:/app/configs/config.yaml:ro \
  -e DATABASE_PASSWORD=your-password \
  -e JWT_SECRET=your-jwt-secret \
  -e ENCRYPTION_KEY=your-encryption-key \
  -e FEISHU_APP_ID=cli_xxxxxxxxxx \
  -e FEISHU_APP_SECRET=xxxxxxxxxxxxxxxx \
  -e FEISHU_REDIRECT_URI=https://your-domain.com/auth/feishu/callback \
  k8s-manager:latest
```

或使用 Docker Compose：

```bash
# 编辑 docker-compose.yml 中的环境变量
docker compose up -d
```

访问 http://localhost

### Kubernetes 部署

项目提供了 Kubernetes 部署配置，位于 `deploy/` 目录。

```bash
# 1. 构建并推送镜像
docker build -t your-registry/k8s-manager:latest .
docker push your-registry/k8s-manager:latest

# 2. 修改配置
#    - deploy/configmap.yaml: 数据库连接、飞书配置、JWT 密钥等
#    - deploy/deployment.yaml: 镜像地址
#    - deploy/service.yaml: NodePort 端口（默认 30000）

# 3. 一键部署
kubectl apply -k deploy/

# 或逐个部署
kubectl apply -f deploy/configmap.yaml
kubectl apply -f deploy/deployment.yaml
kubectl apply -f deploy/service.yaml

# 4. 验证
kubectl get pods -l app=k8s-manager
kubectl logs -l app=k8s-manager
```

访问 `http://<node-ip>:30000`

#### 管理本机集群

如果 K8s-Manager 部署在本机集群中，需要通过管理页面上传 kubeconfig 来管理该集群。注意 Pod 内无法访问 `127.0.0.1`（指向 Pod 自身），需要将 kubeconfig 中的 API Server 地址替换：

| 集群类型 | 原地址 | 替换为 |
|---------|--------|--------|
| Docker Desktop | `https://127.0.0.1:6443` | `https://kubernetes.docker.internal:6443` |
| minikube | `https://127.0.0.1:xxxxx` | `https://host.minikube.internal:xxxxx` |

## 配置说明

### 配置文件示例 (`config.yaml`)

```yaml
# 服务器配置
server:
  port: 8000
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

# 加密配置（用于加密存储 kubeconfig）
encryption:
  key: "your-encryption-key-please-change-in-production"
```

### 环境变量覆盖

支持通过环境变量覆盖配置（推荐用于敏感信息）：

| 环境变量 | 说明 |
|---------|------|
| `DATABASE_PASSWORD` | 数据库密码 |
| `JWT_SECRET` | JWT 签名密钥 |
| `ENCRYPTION_KEY` | AES 加密密钥（用于加密 kubeconfig） |
| `FEISHU_APP_ID` | 飞书应用 App ID |
| `FEISHU_APP_SECRET` | 飞书应用 App Secret |
| `FEISHU_REDIRECT_URI` | 飞书 OAuth 回调地址 |
| `CONFIG_PATH` | 配置文件路径（默认 `/app/configs/config.yaml`） |

## 数据库初始化

应用使用 MySQL 存储用户信息、权限配置、集群信息和资源变更历史。**数据表由 GORM AutoMigrate 在应用启动时自动创建，无需手动执行迁移脚本。**

你只需提前创建好数据库即可：

```bash
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS k8s_manager DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

> `backend/migrations/` 目录下的 SQL 文件仅作为表结构参考文档，不需要手动执行。

### 历史数据迁移

如果从旧版本（支持 in-cluster 模式）升级，应用启动时会自动将遗留的 `source=config` 集群记录转为 `source=database`。历史数据（resource_histories、user_permissions）通过集群名字关联，数据不会丢失。升级后需要通过管理页面为这些集群上传 kubeconfig 才能恢复连接。

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
| POST | `/deployments/:name/restart` | 重启 Deployment（滚动重启） |
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

### 集群管理（仅管理员）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/clusters` | 获取集群管理列表 |
| GET | `/api/v1/admin/clusters/:name` | 获取集群管理详情 |
| POST | `/api/v1/admin/clusters` | 添加集群（上传 kubeconfig） |
| PUT | `/api/v1/admin/clusters/:name` | 更新集群（描述/kubeconfig） |
| DELETE | `/api/v1/admin/clusters/:name` | 删除集群 |
| POST | `/api/v1/admin/clusters/test-connection` | 测试新 kubeconfig 连接 |

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

## 常见问题

### Q: 无法连接数据库？
A: 确保：
1. 数据库地址从容器内可访问（如使用 Docker，注意网络配置）
2. 数据库用户有正确权限
3. 数据库已创建

### Q: 飞书登录失败？
A: 检查：
1. 飞书应用配置正确
2. redirect_uri 与飞书后台配置一致
3. 应用已发布

### Q: 从旧版本升级后集群显示断开？
A: 旧版本使用 in-cluster 模式的集群在升级后需要通过管理页面上传 kubeconfig。历史数据会自动保留（通过集群名称关联）。

## License

MIT License
