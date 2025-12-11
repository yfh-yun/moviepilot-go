-- 插入默认权限
INSERT INTO permissions (name, resource, action, description) VALUES
-- 订阅相关权限
('read:subscribe', 'subscribe', 'read', '查看订阅'),
('write:subscribe', 'subscribe', 'write', '创建/修改订阅'),
('delete:subscribe', 'subscribe', 'delete', '删除订阅'),

-- 下载相关权限
('read:download', 'download', 'read', '查看下载'),
('write:download', 'download', 'write', '创建/修改下载'),
('delete:download', 'download', 'delete', '删除下载'),

-- 站点相关权限
('read:site', 'site', 'read', '查看站点'),
('write:site', 'site', 'write', '创建/修改站点'),
('delete:site', 'site', 'delete', '删除站点'),

-- 用户管理权限
('manage:user', 'user', 'manage', '管理用户'),

-- 系统管理权限
('manage:system', 'system', 'manage', '管理系统设置')
ON CONFLICT (name) DO NOTHING;
