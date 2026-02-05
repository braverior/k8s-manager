# K8s 管理工具 - 数据库设计文档

## 概述

数据库使用 MySQL，主要包含一张表：
- `resource_histories` - 存储 K8s 资源的版本历史

> **注意**：集群配置从配置文件加载，不存储在数据库中。

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
| cluster_name | VARCHAR(255) | 集群名称（对应配置文件中的集群名） |
| namespace | VARCHAR(255) | Kubernetes 命名空间 |
| resource_type | VARCHAR(50) | 资源类型：ConfigMap, Deployment, Service, HPA |
| resource_name | VARCHAR(255) | 资源名称 |
| version | INT UNSIGNED | 版本号（按资源递增） |
| content | LONGTEXT | 完整的资源 JSON 内容 |
| operation | VARCHAR(50) | 产生该版本的操作类型 |
| operator | VARCHAR(255) | 执行操作的用户/系统 |
| created_at | DATETIME | 记录创建时间 |

**索引说明：**
- `idx_cluster_namespace` - 集群+命名空间复合索引
- `idx_resource` - 特定资源查询复合索引
- `idx_version` - 版本查询复合索引
- `idx_created_at` - 时间查询索引

---

## 实体关系图

```
┌─────────────────────────────────────────────────────────────┐
│                    resource_histories                        │
├─────────────────────────────────────────────────────────────┤
│ PK │ id           │ BIGINT UNSIGNED AUTO_INCREMENT          │
│    │ cluster_name │ VARCHAR(255) NOT NULL                   │
│    │ namespace    │ VARCHAR(255) NOT NULL                   │
│    │ resource_type│ VARCHAR(50) NOT NULL                    │
│    │ resource_name│ VARCHAR(255) NOT NULL                   │
│    │ version      │ INT UNSIGNED NOT NULL                   │
│    │ content      │ LONGTEXT NOT NULL                       │
│    │ operation    │ VARCHAR(50) NOT NULL                    │
│    │ operator     │ VARCHAR(255)                            │
│    │ created_at   │ DATETIME DEFAULT CURRENT_TIMESTAMP      │
└─────────────────────────────────────────────────────────────┘
```

---

## 版本管理策略

### 版本号工作原理

1. 每个资源（由 cluster_name + namespace + resource_type + resource_name 唯一标识）维护独立的版本序列
2. 版本号从 1 开始，每次操作递增
3. 保留所有版本（无清理策略）
4. 回滚操作会使用历史内容创建新版本

### 版本递增逻辑

```go
// 获取特定资源的最新版本号
latestVersion, _ := historyRepo.GetLatestVersion(ctx, clusterName, namespace, resourceType, resourceName)

// 创建新的历史记录，版本号递增
newHistory := &entity.ResourceHistory{
    ClusterName:  clusterName,
    Namespace:    namespace,
    ResourceType: resourceType,
    ResourceName: resourceName,
    Version:      latestVersion + 1,  // 递增
    Content:      content,
    Operation:    operation,
}
```

### 操作类型说明

| 操作类型 | 说明 |
|----------|------|
| `create` | 资源被创建 |
| `update` | 资源被更新 |
| `delete` | 资源被删除 |
| `rollback_to_vN` | 资源被回滚到版本 N |

---

## 常用查询示例

### 获取某个资源的所有版本

```sql
SELECT * FROM resource_histories
WHERE cluster_name = 'production'
  AND namespace = 'default'
  AND resource_type = 'ConfigMap'
  AND resource_name = 'app-config'
ORDER BY version DESC;
```

### 获取最新版本号

```sql
SELECT MAX(version) as latest_version
FROM resource_histories
WHERE cluster_name = 'production'
  AND namespace = 'default'
  AND resource_type = 'ConfigMap'
  AND resource_name = 'app-config';
```

### 分页获取历史记录

```sql
SELECT * FROM resource_histories
WHERE cluster_name = 'production' AND namespace = 'default'
ORDER BY created_at DESC
LIMIT 20 OFFSET 0;
```

### 获取最近 24 小时内变更的资源

```sql
SELECT DISTINCT resource_type, resource_name, MAX(version) as latest_version
FROM resource_histories
WHERE cluster_name = 'production'
  AND namespace = 'default'
  AND created_at >= DATE_SUB(NOW(), INTERVAL 24 HOUR)
GROUP BY resource_type, resource_name;
```

---

## 数据库迁移

执行迁移脚本：

```bash
mysql -u root -p k8s_manager < migrations/001_create_resource_histories_table.sql
```

或者使用 GORM 自动迁移（应用默认启用）。
