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
