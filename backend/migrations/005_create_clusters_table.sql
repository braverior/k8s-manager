-- Migration: 005_create_clusters_table.sql
-- 创建集群信息表（动态集群管理）
-- 注意：此表由 GORM AutoMigrate 自动创建，此文件仅作为结构参考

CREATE TABLE IF NOT EXISTS clusters (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL COMMENT '集群名称（唯一）',
    description VARCHAR(1024) DEFAULT '' COMMENT '集群描述',
    cluster_type VARCHAR(50) NOT NULL DEFAULT 'kubeconfig' COMMENT '集群类型：in-cluster, kubeconfig',
    kubeconfig_data TEXT COMMENT 'AES-256-GCM 加密后的 kubeconfig（base64 编码）',
    api_server VARCHAR(1024) DEFAULT '' COMMENT 'API Server 地址（缓存显示用）',
    status VARCHAR(50) NOT NULL DEFAULT 'disconnected' COMMENT '连接状态：connected, disconnected',
    source VARCHAR(50) NOT NULL DEFAULT 'database' COMMENT '数据来源：config（配置文件导入）, database（页面添加）',
    created_by VARCHAR(255) DEFAULT '' COMMENT '创建者',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    UNIQUE KEY uk_name (name),
    INDEX idx_source (source),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='集群信息表';
