-- 创建 subscriptions 表
CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,  -- movie, tv
    tmdb_id INT,
    imdb_id VARCHAR(20),
    season INT,
    quality VARCHAR(50),
    resolution VARCHAR(20),  -- 1080p, 4K, etc
    source VARCHAR(50),  -- BluRay, WEB-DL, etc
    codec VARCHAR(20),  -- H264, H265, etc
    filter_rules JSONB,
    enabled BOOLEAN DEFAULT true NOT NULL,
    auto_download BOOLEAN DEFAULT true NOT NULL,
    notify_on_match BOOLEAN DEFAULT false NOT NULL,
    last_refresh_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_type ON subscriptions(type);
CREATE INDEX idx_subscriptions_tmdb_id ON subscriptions(tmdb_id);
CREATE INDEX idx_subscriptions_enabled ON subscriptions(enabled);
CREATE INDEX idx_subscriptions_deleted_at ON subscriptions(deleted_at);

-- 添加注释
COMMENT ON TABLE subscriptions IS '订阅表';
COMMENT ON COLUMN subscriptions.id IS '订阅ID';
COMMENT ON COLUMN subscriptions.user_id IS '用户ID';
COMMENT ON COLUMN subscriptions.name IS '订阅名称';
COMMENT ON COLUMN subscriptions.type IS '类型: movie-电影, tv-剧集';
COMMENT ON COLUMN subscriptions.tmdb_id IS 'TMDB ID';
COMMENT ON COLUMN subscriptions.imdb_id IS 'IMDB ID';
COMMENT ON COLUMN subscriptions.season IS '季数（仅剧集）';
COMMENT ON COLUMN subscriptions.quality IS '质量要求';
COMMENT ON COLUMN subscriptions.resolution IS '分辨率';
COMMENT ON COLUMN subscriptions.source IS '来源';
COMMENT ON COLUMN subscriptions.codec IS '编码';
COMMENT ON COLUMN subscriptions.filter_rules IS '过滤规则（JSON）';
COMMENT ON COLUMN subscriptions.enabled IS '是否启用';
COMMENT ON COLUMN subscriptions.auto_download IS '是否自动下载';
COMMENT ON COLUMN subscriptions.notify_on_match IS '匹配时是否通知';
COMMENT ON COLUMN subscriptions.last_refresh_at IS '最后刷新时间';
COMMENT ON COLUMN subscriptions.created_at IS '创建时间';
COMMENT ON COLUMN subscriptions.updated_at IS '更新时间';
COMMENT ON COLUMN subscriptions.deleted_at IS '删除时间（软删除）';
