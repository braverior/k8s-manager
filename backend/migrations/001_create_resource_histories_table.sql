-- Migration: 001_create_resource_histories_table.sql
-- 创建资源版本历史表

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
