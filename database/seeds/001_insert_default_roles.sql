-- 插入默认角色
INSERT INTO roles (name, display_name, description, is_system) VALUES
('admin', '管理员', '系统管理员，拥有所有权限', TRUE),
('user', '普通用户', '普通用户，拥有基本功能权限', TRUE),
('guest', '访客', '访客，只读权限', TRUE)
ON CONFLICT (name) DO NOTHING;
