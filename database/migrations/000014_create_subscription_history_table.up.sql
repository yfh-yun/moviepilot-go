-- 创建 subscription_history 表
CREATE TABLE IF NOT EXISTS subscription_history (
    id SERIAL PRIMARY KEY,
    subscription_id INT NOT NULL,
    action VARCHAR(50) NOT NULL,
    items_found INT DEFAULT 0,
    items_matched INT DEFAULT 0,
    items_downloaded INT DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE CASCADE
);

cd /workspaces/moviepilot/moviepilot-go-project/moviepilot-go && go build ./tests/business/tmdb_service_test.go 2>&1
CREATE INDEX idx_subscription_history_subscription_id ON subscription_history(subscription_id);
CREATE INDEX idx_subscription_history_created_at ON subscription_history(created_at);

-- 添加注释
COMMENT ON TABLE subscription_history IS '订阅历史表';
COMMENT ON COLUMN subscription_history.id IS '历史ID';
COMMENT ON COLUMN subscription_history.subscription_id IS '订阅ID';
COMMENT ON COLUMN subscription_history.action IS '操作类型';
COMMENT ON COLUMN subscription_history.items_found IS '发现项数';
COMMENT ON COLUMN subscription_history.items_matched IS '匹配项数';
COMMENT ON COLUMN subscription_history.items_downloaded IS '下载项数';
COMMENT ON COLUMN subscription_history.error_message IS '错误信息';
COMMENT ON COLUMN subscription_history.created_at IS '创建时间';
