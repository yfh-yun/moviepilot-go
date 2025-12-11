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
