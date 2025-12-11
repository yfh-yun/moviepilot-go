-- 创建 download_tasks 表
CREATE TABLE IF NOT EXISTS download_tasks (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    subscription_id INT,
    downloader_type VARCHAR(20) NOT NULL,  -- qbittorrent, transmission
    hash VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    save_path VARCHAR(500),
    size BIGINT,
    status VARCHAR(20) DEFAULT 'downloading' NOT NULL,  -- downloading, completed, error, paused, seeding
    progress DECIMAL(5,2) DEFAULT 0 NOT NULL,
    download_speed BIGINT DEFAULT 0,
    upload_speed BIGINT DEFAULT 0,
    downloaded BIGINT DEFAULT 0,
    uploaded BIGINT DEFAULT 0,
    ratio DECIMAL(5,2) DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE SET NULL
);

-- 创建索引
CREATE INDEX idx_download_tasks_user_id ON download_tasks(user_id);
CREATE INDEX idx_download_tasks_subscription_id ON download_tasks(subscription_id);
CREATE INDEX idx_download_tasks_hash ON download_tasks(hash);
CREATE INDEX idx_download_tasks_status ON download_tasks(status);
CREATE INDEX idx_download_tasks_downloader_type ON download_tasks(downloader_type);
CREATE INDEX idx_download_tasks_created_at ON download_tasks(created_at);

-- 添加注释
COMMENT ON TABLE download_tasks IS '下载任务表';
COMMENT ON COLUMN download_tasks.id IS '任务ID';
COMMENT ON COLUMN download_tasks.user_id IS '用户ID';
COMMENT ON COLUMN download_tasks.subscription_id IS '订阅ID';
COMMENT ON COLUMN download_tasks.downloader_type IS '下载器类型';
COMMENT ON COLUMN download_tasks.hash IS '种子哈希';
COMMENT ON COLUMN download_tasks.name IS '任务名称';
COMMENT ON COLUMN download_tasks.save_path IS '保存路径';
COMMENT ON COLUMN download_tasks.size IS '文件大小（字节）';
COMMENT ON COLUMN download_tasks.status IS '状态';
COMMENT ON COLUMN download_tasks.progress IS '进度（百分比）';
COMMENT ON COLUMN download_tasks.download_speed IS '下载速度（字节/秒）';
COMMENT ON COLUMN download_tasks.upload_speed IS '上传速度（字节/秒）';
COMMENT ON COLUMN download_tasks.downloaded IS '已下载（字节）';
COMMENT ON COLUMN download_tasks.uploaded IS '已上传（字节）';
COMMENT ON COLUMN download_tasks.ratio IS '分享率';
COMMENT ON COLUMN download_tasks.error_message IS '错误信息';
COMMENT ON COLUMN download_tasks.started_at IS '开始时间';
COMMENT ON COLUMN download_tasks.completed_at IS '完成时间';
COMMENT ON COLUMN download_tasks.created_at IS '创建时间';
COMMENT ON COLUMN download_tasks.updated_at IS '更新时间';
