INSERT INTO iam.permissions (uuid, name, description, created_at, updated_at)
VALUES (gen_random_uuid(), 'migrations:*', 'Run migrations and seeders', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;
