-- Migration: 004_init_admin_user.sql
-- 初始化管理员用户（可选）
-- 注意：实际使用时请修改 id 为真实的飞书 open_id

-- 示例：插入一个初始管理员用户
-- INSERT INTO users (id, name, email, role, status, created_at) VALUES
-- ('ou_xxxxxxxxxx', '管理员', 'admin@example.com', 'admin', 'active', NOW());

-- 示例：为管理员添加所有集群的权限（管理员实际上不需要，因为代码里管理员自动拥有所有权限）
-- INSERT INTO user_permissions (user_id, cluster, namespaces, created_at) VALUES
-- ('ou_xxxxxxxxxx', 'local', '["*"]', NOW());
