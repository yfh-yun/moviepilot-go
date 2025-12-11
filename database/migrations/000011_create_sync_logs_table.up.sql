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
