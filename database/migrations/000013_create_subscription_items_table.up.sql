-- 创建 subscription_items 表
CREATE TABLE IF NOT EXISTS subscription_items (
    id SERIAL PRIMARY KEY,
    subscription_id INT NOT NULL,
    title VARCHAR(200) NOT NULL,
    torrent_url VARCHAR(500),
    magnet_link TEXT,
    size BIGINT,
    seeders INT,
    leechers INT,
    publish_date TIMESTAMP,
    matched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    downloaded BOOLEAN DEFAULT false NOT NULL,
    download_task_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX idx_subscription_items_subscription_id ON subscription_items(subscription_id);
CREATE INDEX idx_subscription_items_downloaded ON subscription_items(downloaded);
CREATE INDEX idx_subscription_items_publish_date ON subscription_items(publish_date);

-- 添加注释
COMMENT ON TABLE subscription_items IS '订阅项表';
COMMENT ON COLUMN subscription_items.id IS '订阅项ID';
COMMENT ON COLUMN subscription_items.subscription_id IS '订阅ID';
COMMENT ON COLUMN subscription_items.title IS '标';
COMMENT ON COLUMN subscription_items.torrent_url IS '种子URL';
COMMENT ON COLUMN subscription_items.magnet_link IS '磁力链接';
COMMENT ON COLUMN subscription_items.size IS '文件大小（字节）';
COMMENT ON COLUMN subscription_items.seeders IS '做种数';
COMMENT ON COLUMN subscription_items.leechers IS '下载数';
COMMENT ON COLUMN subscription_items.publish_date IS '发布时'''';
COMMENT ON COLUMN subscription_items.matched_at IS '匹配时间';
COMMENT ON COLUMN subscription_items.downloaded IS '是否已下载';
COMMENT ON COLUMN subscription_items.download_task_id IS '下载任务ID';
COMMENT ON COLUMN subscription_items.created_at IS '创建时间';
