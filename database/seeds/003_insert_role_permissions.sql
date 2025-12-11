-- 分配角色权限

-- admin 拥有所有权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = 'admin'),
    id
FROM permissions
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- user 拥有基本权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = 'user'),
    id
FROM permissions
WHERE name IN (
    'read:subscribe', 'write:subscribe', 'delete:subscribe',
    'read:download', 'write:download', 'delete:download',
    'read:site', 'write:site', 'delete:site'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- guest 只有读权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = 'guest'),
    id
FROM permissions
WHERE action = 'read'
ON CONFLICT (role_id, permission_id) DO NOTHING;
