#!/bin/bash

# 数据库迁移脚本生成器
# Week 7 Day 1: 生成所有数据库迁移文件

set -e

MIGRATIONS_DIR="database/migrations"
SEEDS_DIR="database/seeds"

# 创建目录
mkdir -p $MIGRATIONS_DIR
mkdir -p $SEEDS_DIR

echo "🚀 开始生成数据库迁移脚本..."

# 已经创建了 000001_create_users_table，跳过

# 000002: roles 表
cat > $MIGRATIONS_DIR/000002_create_roles_table.up.sql << 'EOF'
-- 创建 roles 表
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_roles_name ON roles(name);

COMMENT ON TABLE roles IS '角色表';
COMMENT ON COLUMN roles.is_system IS '是否系统角色（系统角色不可删除）';
EOF

cat > $MIGRATIONS_DIR/000002_create_roles_table.down.sql << 'EOF'
DROP TABLE IF EXISTS roles CASCADE;
EOF

echo "✅ 创建 roles 表迁移"

# 000003: permissions 表
cat > $MIGRATIONS_DIR/000003_create_permissions_table.up.sql << 'EOF'
-- 创建 permissions 表
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_permissions_name ON permissions(name);
CREATE INDEX idx_permissions_resource ON permissions(resource);

COMMENT ON TABLE permissions IS '权限表';
COMMENT ON COLUMN permissions.name IS '权限名称（格式：action:resource）';
EOF

cat > $MIGRATIONS_DIR/000003_create_permissions_table.down.sql << 'EOF'
DROP TABLE IF EXISTS permissions CASCADE;
EOF

echo "✅ 创建 permissions 表迁移"

# 000004: user_roles 表
cat > $MIGRATIONS_DIR/000004_create_user_roles_table.up.sql << 'EOF'
-- 创建 user_roles 表
CREATE TABLE IF NOT EXISTS user_roles (
    user_id INT NOT NULL,
    role_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

COMMENT ON TABLE user_roles IS '用户角色关联表';
EOF

cat > $MIGRATIONS_DIR/000004_create_user_roles_table.down.sql << 'EOF'
DROP TABLE IF EXISTS user_roles CASCADE;
EOF

echo "✅ 创建 user_roles 表迁移"

# 000005: role_permissions 表
cat > $MIGRATIONS_DIR/000005_create_role_permissions_table.up.sql << 'EOF'
-- 创建 role_permissions 表
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INT NOT NULL,
    permission_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

COMMENT ON TABLE role_permissions IS '角色权限关联表';
EOF

cat > $MIGRATIONS_DIR/000005_create_role_permissions_table.down.sql << 'EOF'
DROP TABLE IF EXISTS role_permissions CASCADE;
EOF

echo "✅ 创建 role_permissions 表迁移"

# 000006: auth_logs 表
cat > $MIGRATIONS_DIR/000006_create_auth_logs_table.up.sql << 'EOF'
-- 创建 auth_logs 表
CREATE TABLE IF NOT EXISTS auth_logs (
    id SERIAL PRIMARY KEY,
    user_id INT,
    action VARCHAR(50) NOT NULL,
    ip_address VARCHAR(50),
    user_agent TEXT,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_auth_logs_user_id ON auth_logs(user_id);
CREATE INDEX idx_auth_logs_action ON auth_logs(action);
CREATE INDEX idx_auth_logs_created_at ON auth_logs(created_at);

COMMENT ON TABLE auth_logs IS '认证日志表（审计）';
COMMENT ON COLUMN auth_logs.action IS '操作类型：login, logout, register, password_reset';
COMMENT ON COLUMN auth_logs.status IS '状态：success, failed';
EOF

cat > $MIGRATIONS_DIR/000006_create_auth_logs_table.down.sql << 'EOF'
DROP TABLE IF EXISTS auth_logs CASCADE;
EOF

echo "✅ 创建 auth_logs 表迁移"

# 000007: sites 表
cat > $MIGRATIONS_DIR/000007_create_sites_table.up.sql << 'EOF'
-- 创建 sites 表
CREATE TABLE IF NOT EXISTS sites (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    url VARCHAR(500) NOT NULL,
    type VARCHAR(20) NOT NULL,
    priority INT DEFAULT 5,
    enabled BOOLEAN DEFAULT TRUE NOT NULL,
    
    -- 认证信息
    cookie TEXT,
    user_agent VARCHAR(500),
    proxy VARCHAR(200),
    
    -- 签到配置
    checkin_enabled BOOLEAN DEFAULT FALSE NOT NULL,
    checkin_cron VARCHAR(50) DEFAULT '0 8 * * *',
    checkin_url VARCHAR(500),
    
    -- 流量统计
    upload BIGINT DEFAULT 0,
    download BIGINT DEFAULT 0,
    ratio DECIMAL(10, 2) DEFAULT 0,
    
    -- 状态信息
    status VARCHAR(20) DEFAULT 'active' NOT NULL,
    last_checkin TIMESTAMP,
    last_sync TIMESTAMP,
    error_message TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_sites_user_id ON sites(user_id);
CREATE INDEX idx_sites_type ON sites(type);
CREATE INDEX idx_sites_status ON sites(status);
CREATE INDEX idx_sites_enabled ON sites(enabled);

COMMENT ON TABLE sites IS '站点表';
COMMENT ON COLUMN sites.type IS '站点类型：pt, public, rss';
COMMENT ON COLUMN sites.status IS '状态：active, error, disabled';
EOF

cat > $MIGRATIONS_DIR/000007_create_sites_table.down.sql << 'EOF'
DROP TABLE IF EXISTS sites CASCADE;
EOF

echo "✅ 创建 sites 表迁移"

# 000008: site_cookies 表
cat > $MIGRATIONS_DIR/000008_create_site_cookies_table.up.sql << 'EOF'
-- 创建 site_cookies 表
CREATE TABLE IF NOT EXISTS site_cookies (
    id SERIAL PRIMARY KEY,
    site_id INT NOT NULL,
    cookie TEXT NOT NULL,
    is_valid BOOLEAN DEFAULT TRUE NOT NULL,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
);

CREATE INDEX idx_site_cookies_site_id ON site_cookies(site_id);
CREATE INDEX idx_site_cookies_is_valid ON site_cookies(is_valid);

COMMENT ON TABLE site_cookies IS 'Cookie历史表';
EOF

cat > $MIGRATIONS_DIR/000008_create_site_cookies_table.down.sql << 'EOF'
DROP TABLE IF EXISTS site_cookies CASCADE;
EOF

echo "✅ 创建 site_cookies 表迁移"

# 000009: checkin_logs 表
cat > $MIGRATIONS_DIR/000009_create_checkin_logs_table.up.sql << 'EOF'
-- 创建 checkin_logs 表
CREATE TABLE IF NOT EXISTS checkin_logs (
    id SERIAL PRIMARY KEY,
    site_id INT NOT NULL,
    success BOOLEAN NOT NULL,
    message TEXT,
    bonus INT DEFAULT 0,
    continue_days INT DEFAULT 0,
    error_message TEXT,
    checkin_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
);

CREATE INDEX idx_checkin_logs_site_id ON checkin_logs(site_id);
CREATE INDEX idx_checkin_logs_checkin_time ON checkin_logs(checkin_time);
CREATE INDEX idx_checkin_logs_success ON checkin_logs(success);

COMMENT ON TABLE checkin_logs IS '签到日志表';
EOF

cat > $MIGRATIONS_DIR/000009_create_checkin_logs_table.down.sql << 'EOF'
DROP TABLE IF EXISTS checkin_logs CASCADE;
EOF

echo "✅ 创建 checkin_logs 表迁移"

# 000010: site_stats 表
cat > $MIGRATIONS_DIR/000010_create_site_stats_table.up.sql << 'EOF'
-- 创建 site_stats 表
CREATE TABLE IF NOT EXISTS site_stats (
    id SERIAL PRIMARY KEY,
    site_id INT NOT NULL,
    date DATE NOT NULL,
    upload_delta BIGINT DEFAULT 0,
    download_delta BIGINT DEFAULT 0,
    upload_total BIGINT DEFAULT 0,
    download_total BIGINT DEFAULT 0,
    ratio DECIMAL(10, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
    UNIQUE (site_id, date)
);

CREATE INDEX idx_site_stats_site_id ON site_stats(site_id);
CREATE INDEX idx_site_stats_date ON site_stats(date);

COMMENT ON TABLE site_stats IS '流量统计表';
EOF

cat > $MIGRATIONS_DIR/000010_create_site_stats_table.down.sql << 'EOF'
DROP TABLE IF EXISTS site_stats CASCADE;
EOF

echo "✅ 创建 site_stats 表迁移"

# 000011: sync_logs 表
cat > $MIGRATIONS_DIR/000011_create_sync_logs_table.up.sql << 'EOF'
-- 创建 sync_logs 表
CREATE TABLE IF NOT EXISTS sync_logs (
    id SERIAL PRIMARY KEY,
    site_id INT NOT NULL,
    success BOOLEAN NOT NULL,
    duration_ms INT,
    error_message TEXT,
    synced_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
);

CREATE INDEX idx_sync_logs_site_id ON sync_logs(site_id);
CREATE INDEX idx_sync_logs_synced_at ON sync_logs(synced_at);

COMMENT ON TABLE sync_logs IS '同步日志表';
EOF

cat > $MIGRATIONS_DIR/000011_create_sync_logs_table.down.sql << 'EOF'
DROP TABLE IF EXISTS sync_logs CASCADE;
EOF

echo "✅ 创建 sync_logs 表迁移"

echo ""
echo "🎉 所有迁移脚本生成完成！"
echo "📁 迁移文件位置: $MIGRATIONS_DIR"
echo "📊 共生成 22 个文件（11 up + 11 down）"
