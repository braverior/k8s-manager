# K8s 管理工具 - 数据库设计文档

## 概述

数据库使用 MySQL，包含以下数据表：
- `resource_histories` - 存储 K8s 资源的版本历史
- `users` - 用户信息（飞书 OAuth 登录后自动创建）
- `user_permissions` - 用户权限（集群 + 命名空间级别）
- `clusters` - 集群信息（kubeconfig 加密存储）

> **注意**：所有数据表由 GORM AutoMigrate 在应用启动时自动创建，`migrations/` 目录下的 SQL 文件仅作为结构参考。

## 数据表

### resource_histories 表

存储所有被管理 Kubernetes 资源的版本历史。

```sql
CREATE TABLE IF NOT EXISTS resource_histories (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    cluster_name VARCHAR(255) NOT NULL COMMENT '集群名称',
    namespace VARCHAR(255) NOT NULL COMMENT 'Kubernetes 命名空间',
    resource_type VARCHAR(50) NOT NULL COMMENT '资源类型：ConfigMap, Deployment, Service, HPA',
    resource_name VARCHAR(255) NOT NULL COMMENT '资源名称',
    version INT UNSIGNED NOT NULL COMMENT '版本号，按资源自增',
    content LONGTEXT NOT NULL COMMENT '资源 YAML/JSON 内容',
    operation VARCHAR(50) NOT NULL COMMENT '操作类型：create, update, delete, rollback_to_vN',
    operator VARCHAR(255) COMMENT '操作者标识',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_cluster_namespace (cluster_name, namespace),
    INDEX idx_resource (cluster_name, namespace, resource_type, resource_name),
    INDEX idx_version (cluster_name, namespace, resource_type, resource_name, version),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='资源版本历史表';
```

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED | 主键，自增 |
| cluster_name | VARCHAR(255) | 集群名称 |
| namespace | VARCHAR(255) | Kubernetes 命名空间 |
| resource_type | VARCHAR(50) | 资源类型：ConfigMap, Deployment, Service, HPA |
| resource_name | VARCHAR(255) | 资源名称 |
| version | INT UNSIGNED | 版本号（按资源递增） |
| content | LONGTEXT | 完整的资源 JSON 内容 |
| operation | VARCHAR(50) | 产生该版本的操作类型 |
| operator | VARCHAR(255) | 执行操作的用户/系统 |
| created_at | DATETIME | 记录创建时间 |

### users 表

存储飞书 OAuth 登录的用户信息。

```sql
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY COMMENT '飞书 open_id',
    name VARCHAR(255) NOT NULL COMMENT '用户姓名',
    email VARCHAR(255) COMMENT '用户邮箱',
    mobile VARCHAR(50) COMMENT '手机号',
    avatar_url VARCHAR(1024) COMMENT '头像 URL',
    employee_id VARCHAR(255) COMMENT '企业员工 ID',
    department_id VARCHAR(255) COMMENT '部门 ID',
    department_name VARCHAR(255) COMMENT '部门名称',
    department_path VARCHAR(1024) COMMENT '部门完整路径',
    role VARCHAR(50) NOT NULL DEFAULT 'user' COMMENT '角色：admin/user',
    status VARCHAR(50) NOT NULL DEFAULT 'active' COMMENT '状态：active/disabled',
    last_login_at DATETIME NULL COMMENT '最后登录时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    INDEX idx_email (email),
    INDEX idx_employee_id (employee_id),
    INDEX idx_department_id (department_id),
    INDEX idx_role (role),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
```

### user_permissions 表

存储用户对集群和命名空间的访问权限。

```sql
CREATE TABLE IF NOT EXISTS user_permissions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL COMMENT '用户 ID（关联 users.id）',
    cluster VARCHAR(255) NOT NULL COMMENT '集群名称',
    namespaces TEXT NOT NULL COMMENT '命名空间列表，JSON 数组格式，["*"] 表示所有',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    INDEX idx_user_id (user_id),
    INDEX idx_cluster (cluster),
    UNIQUE KEY uk_user_cluster (user_id, cluster),
    CONSTRAINT fk_user_permissions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户权限表';
```

### clusters 表

存储集群信息。kubeconfig 使用 AES-256-GCM 加密存储。

```sql
CREATE TABLE IF NOT EXISTS clusters (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL COMMENT '集群名称（唯一）',
    description VARCHAR(1024) DEFAULT '' COMMENT '集群描述',
    cluster_type VARCHAR(50) NOT NULL DEFAULT 'kubeconfig' COMMENT '集群类型',
    kubeconfig_data TEXT COMMENT 'AES-256-GCM 加密后的 kubeconfig（base64 编码）',
    api_server VARCHAR(1024) DEFAULT '' COMMENT 'API Server 地址（缓存显示用）',
    status VARCHAR(50) NOT NULL DEFAULT 'disconnected' COMMENT '连接状态：connected, disconnected',
    source VARCHAR(50) NOT NULL DEFAULT 'database' COMMENT '数据来源',
    created_by VARCHAR(255) DEFAULT '' COMMENT '创建者',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    UNIQUE KEY uk_name (name),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='集群信息表';
```

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED | 主键，自增 |
| name | VARCHAR(255) | 集群名称，唯一索引 |
| description | VARCHAR(1024) | 集群描述 |
| cluster_type | VARCHAR(50) | 集群类型（kubeconfig） |
| kubeconfig_data | TEXT | AES-256-GCM 加密后的 kubeconfig 内容（base64 编码） |
| api_server | VARCHAR(1024) | Kubernetes API Server 地址，从 kubeconfig 解析缓存 |
| status | VARCHAR(50) | 连接状态：`connected`、`disconnected` |
| source | VARCHAR(50) | 数据来源（`database`） |
| created_by | VARCHAR(255) | 创建者用户名 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 最后更新时间 |

**安全说明：**

- kubeconfig 使用 AES-256-GCM 加密后存储，加密密钥通过环境变量 `ENCRYPTION_KEY` 配置
- API 响应中不返回 kubeconfig 原始内容，仅返回 `has_kubeconfig: true/false`
- 删除集群时会同步清理 `user_permissions` 表中关联的权限记录

---

## 实体关系图

```
┌───────────────────┐     ┌───────────────────────┐     ┌───────────────────┐
│      users        │     │   user_permissions    │     │     clusters      │
├───────────────────┤     ├───────────────────────┤     ├───────────────────┤
│ PK │ id (open_id) │◄────│ FK │ user_id          │     │ PK │ id           │
│    │ name         │     │    │ cluster ─ ─ ─ ─ ─│─ ─ ▶│ UK │ name         │
│    │ email        │     │    │ namespaces (JSON) │     │    │ description  │
│    │ role         │     │    │ created_at        │     │    │ cluster_type │
│    │ status       │     │    │ updated_at        │     │    │ kubeconfig   │
│    │ ...          │     └───────────────────────┘     │    │  _data (加密) │
└───────────────────┘                                   │    │ api_server   │
                                                        │    │ status       │
┌───────────────────────────────────────────┐           │    │ created_by   │
│            resource_histories             │           │    │ ...          │
├───────────────────────────────────────────┤           └───────────────────┘
│ PK │ id                                  │
│    │ cluster_name                        │
│    │ namespace                           │
│    │ resource_type                       │
│    │ resource_name                       │
│    │ version                             │
│    │ content                             │
│    │ operation                           │
│    │ operator                            │
│    │ created_at                          │
└───────────────────────────────────────────┘
```

> `user_permissions.cluster` 与 `clusters.name` 为逻辑关联（字符串引用），无物理外键。删除集群时由应用层清理孤立权限记录。

---

## 版本管理策略

### 版本号工作原理

1. 每个资源（由 cluster_name + namespace + resource_type + resource_name 唯一标识）维护独立的版本序列
2. 版本号从 1 开始，每次操作递增
3. 保留所有版本（无清理策略）
4. 回滚操作会使用历史内容创建新版本

### 操作类型说明

| 操作类型 | 说明 |
|----------|------|
| `create` | 资源被创建 |
| `update` | 资源被更新 |
| `delete` | 资源被删除 |
| `rollback_to_vN` | 资源被回滚到版本 N |

---

## 数据库迁移

应用使用 GORM AutoMigrate 在启动时自动创建/更新表结构，`migrations/` 目录下的 SQL 文件仅作参考：

```
migrations/
├── 001_create_resource_histories_table.sql
├── 002_create_users_table.sql
├── 003_create_user_permissions_table.sql
├── 004_init_admin_user.sql
└── 005_create_clusters_table.sql
```
