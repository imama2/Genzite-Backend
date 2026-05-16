INSERT INTO iam.roles (uuid, name, description, created_at, updated_at)
VALUES (gen_random_uuid(), 'admin', 'System administrator', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;
