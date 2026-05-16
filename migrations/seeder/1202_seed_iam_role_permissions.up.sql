INSERT INTO iam.role_permissions (role_id, permission_id, uuid, created_at, updated_at)
SELECT r.id, p.id, gen_random_uuid(), NOW(), NOW()
FROM iam.roles r
JOIN iam.permissions p ON p.name = 'migrations:*'
WHERE r.name = 'admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;
