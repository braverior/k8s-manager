-- Migration: 003_create_user_permissions_table.sql
-- 创建用户权限表

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
