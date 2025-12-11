#!/bin/bash

# 数据库种子数据生成器
# Week 7 Day 1: 生成初始数据

set -e

SEEDS_DIR="database/seeds"

# 创建目录
mkdir -p $SEEDS_DIR

echo "🌱 开始生成种子数据脚本..."

# 001: 插入默认角色
cat > $SEEDS_DIR/001_insert_default_roles.sql << 'EOF'
-- 插入默认角色
INSERT INTO roles (name, display_name, description, is_system) VALUES
('admin', '管理员', '系统管理员，拥有所有权限', TRUE),
('user', '普通用户', '普通用户，拥有基本功能权限', TRUE),
('guest', '访客', '访客，只读权限', TRUE)
ON CONFLICT (name) DO NOTHING;
EOF

echo "✅ 创建默认角色种子数据"

# 002: 插入默认权限
cat > $SEEDS_DIR/002_insert_default_permissions.sql << 'EOF'
-- 插入默认权限
INSERT INTO permissions (name, resource, action, description) VALUES
-- 订阅相关权限
('read:subscribe', 'subscribe', 'read', '查看订阅'),
('write:subscribe', 'subscribe', 'write', '创建/修改订阅'),
('delete:subscribe', 'subscribe', 'delete', '删除订阅'),

-- 下载相关权限
('read:download', 'download', 'read', '查看下载'),
('write:download', 'download', 'write', '创建/修改下载'),
('delete:download', 'download', 'delete', '删除下载'),

-- 站点相关权限
('read:site', 'site', 'read', '查看站点'),
('write:site', 'site', 'write', '创建/修改站点'),
('delete:site', 'site', 'delete', '删除站点'),

-- 用户管理权限
('manage:user', 'user', 'manage', '管理用户'),

-- 系统管理权限
('manage:system', 'system', 'manage', '管理系统设置')
ON CONFLICT (name) DO NOTHING;
EOF

echo "✅ 创建默认权限种子数据"

# 003: 分配角色权限
cat > $SEEDS_DIR/003_insert_role_permissions.sql << 'EOF'
-- 分配角色权限

-- admin 拥有所有权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = 'admin'),
    id
FROM permissions
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- user 拥有基本权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = 'user'),
    id
FROM permissions
WHERE name IN (
    'read:subscribe', 'write:subscribe', 'delete:subscribe',
    'read:download', 'write:download', 'delete:download',
    'read:site', 'write:site', 'delete:site'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- guest 只有读权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = 'guest'),
    id
FROM permissions
WHERE action = 'read'
ON CONFLICT (role_id, permission_id) DO NOTHING;
EOF

echo "✅ 创建角色权限关联种子数据"

# 004: 创建默认管理员账户
cat > $SEEDS_DIR/004_insert_default_admin.sql << 'EOF'
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
EOF

echo "✅ 创建默认管理员种子数据"

echo ""
echo "🎉 所有种子数据脚本生成完成！"
echo "📁 种子数据位置: $SEEDS_DIR"
echo "📊 共生成 4 个文件"
echo ""
echo "⚠️  默认管理员账户:"
echo "   用户名: admin"
echo "   密码: admin123"
echo "   ⚠️  请在生产环境中立即修改密码！"
