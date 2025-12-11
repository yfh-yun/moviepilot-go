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
