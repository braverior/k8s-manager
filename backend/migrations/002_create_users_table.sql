-- Migration: 002_create_users_table.sql
-- 创建用户表

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
