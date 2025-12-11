-- 创建 auth_logs 表
CREATE TABLE IF NOT EXISTS auth_logs (
    id SERIAL PRIMARY KEY,
    user_id INT,
    action VARCHAR(50) NOT NULL,
    ip_address VARCHAR(50),
    user_agent TEXT,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_auth_logs_user_id ON auth_logs(user_id);
CREATE INDEX idx_auth_logs_action ON auth_logs(action);
CREATE INDEX idx_auth_logs_created_at ON auth_logs(created_at);

COMMENT ON TABLE auth_logs IS '认证日志表（审计）';
COMMENT ON COLUMN auth_logs.action IS '操作类型：login, logout, register, password_reset';
COMMENT ON COLUMN auth_logs.status IS '状态：success, failed';
