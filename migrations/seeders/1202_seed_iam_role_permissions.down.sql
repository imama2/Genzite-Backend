DELETE FROM iam.role_permissions rp
USING iam.roles r, iam.permissions p
WHERE rp.role_id = r.id AND rp.permission_id = p.id AND r.name = 'admin' AND p.name = 'migrations:*';
