-- 创建 permissions 表
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_permissions_name ON permissions(name);
CREATE INDEX idx_permissions_resource ON permissions(resource);

COMMENT ON TABLE permissions IS '权限表';
COMMENT ON COLUMN permissions.name IS '权限名称（格式：action:resource）';
