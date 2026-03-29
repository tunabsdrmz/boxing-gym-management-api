-- Application users for JWT auth and RBAC (emails stored lowercased by the app).
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'viewer' CHECK (role IN ('admin', 'staff', 'viewer')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- Dev admin: admin@gym.local / admin123 — change in production.
INSERT INTO users (id, email, password_hash, role) VALUES
    (
        'c0000001-0000-4000-8000-000000000001',
        'admin@gym.local',
        '$2a$10$vfiRvvLVUzYd8mb4j9/gz.1eohAVjqyFmZGiCp5t.xTQJycXk6Zae',
        'admin'
    )
ON CONFLICT (email) DO NOTHING;
