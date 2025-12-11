-- 创建 download_history 表
CREATE TABLE IF NOT EXISTS download_history (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    task_id INT,
    hash VARCHAR(64) NOT NULL,
    name VARCHAR(200) NOT NULL,
    size BIGINT,
    downloader_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,  -- completed, failed, deleted
    download_time INT,  -- 下载耗时（秒）
    average_speed BIGINT,  -- 平均速度（字节/秒）
    final_ratio DECIMAL(5,2),  -- 最终分享率
    save_path VARCHAR(500),
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (task_id) REFERENCES download_tasks(id) ON DELETE SET NULL
);

-- 创建索引
CREATE INDEX idx_download_history_user_id ON download_history(user_id);
CREATE INDEX idx_download_history_task_id ON download_history(task_id);
CREATE INDEX idx_download_history_hash ON download_history(hash);
CREATE INDEX idx_download_history_status ON download_history(status);
CREATE INDEX idx_download_history_created_at ON download_history(created_at);

-- 添加注释
COMMENT ON TABLE download_history IS '下载历史表';
COMMENT ON COLUMN download_history.id IS '历史ID';
COMMENT ON COLUMN download_history.user_id IS '用户ID';
COMMENT ON COLUMN download_history.task_id IS '任务ID';
COMMENT ON COLUMN download_history.hash IS '种子哈希';
COMMENT ON COLUMN download_history.name IS '任务名称';
COMMENT ON COLUMN download_history.size IS '文件大小（字节）';
COMMENT ON COLUMN download_history.downloader_type IS '下载器类型';
COMMENT ON COLUMN download_history.status IS '状态';
COMMENT ON COLUMN download_history.download_time IS '下载耗时（秒）';
COMMENT ON COLUMN download_history.average_speed IS '平均速度（字节/秒）';
COMMENT ON COLUMN download_history.final_ratio IS '最终分享率';
COMMENT ON COLUMN download_history.save_path IS '保存路径';
COMMENT ON COLUMN download_history.error_message IS '错误信息';
COMMENT ON COLUMN download_history.started_at IS '开始时间';
COMMENT ON COLUMN download_history.completed_at IS '完成时间';
COMMENT ON COLUMN download_history.created_at IS '创建时间';
