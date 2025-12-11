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
