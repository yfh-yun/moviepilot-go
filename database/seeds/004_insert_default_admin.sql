-- 创建默认管理员账户
-- 用户名: admin
-- 密码: admin123
-- 密码哈希使用 bcrypt (cost=10)

INSERT INTO users (username, email, password_hash, nickname, status) VALUES
('admin', 'admin@moviepilot.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Administrator', 'active')
ON CONFLICT (username) DO NOTHING;

-- 分配管理员角色
INSERT INTO user_roles (user_id, role_id)
SELECT 
    (SELECT id FROM users WHERE username = 'admin'),
    (SELECT id FROM roles WHERE name = 'admin')
WHERE NOT EXISTS (
    SELECT 1 FROM user_roles 
    WHERE user_id = (SELECT id FROM users WHERE username = 'admin')
    AND role_id = (SELECT id FROM roles WHERE name = 'admin')
);
