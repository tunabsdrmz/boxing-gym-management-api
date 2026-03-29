-- Extended gym operations, fighter profile fields, auth tokens, admin flags.
-- Runs on fresh Docker Postgres init after 004_users.sql.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS locked BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS locked_reason TEXT;

ALTER TABLE fighters
    ADD COLUMN IF NOT EXISTS health_notes TEXT,
    ADD COLUMN IF NOT EXISTS contract_end DATE,
    ADD COLUMN IF NOT EXISTS emergency_contact_name TEXT,
    ADD COLUMN IF NOT EXISTS emergency_contact_phone TEXT,
    ADD COLUMN IF NOT EXISTS weight_class TEXT,
    ADD COLUMN IF NOT EXISTS fighter_status TEXT NOT NULL DEFAULT 'amateur'
        CHECK (fighter_status IN ('amateur', 'pro')),
    ADD COLUMN IF NOT EXISTS license_number TEXT;

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens (token_hash);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS schedule_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    resource_type TEXT NOT NULL DEFAULT 'general' CHECK (resource_type IN ('ring', 'mat', 'general')),
    trainer_id UUID REFERENCES trainers (id) ON DELETE SET NULL,
    notes TEXT,
    created_by UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_at > start_at)
);

CREATE INDEX IF NOT EXISTS idx_schedule_events_range ON schedule_events (resource_type, start_at, end_at);

CREATE TABLE IF NOT EXISTS attendance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    gym_date DATE NOT NULL,
    fighter_id UUID NOT NULL REFERENCES fighters (id) ON DELETE CASCADE,
    present BOOLEAN NOT NULL DEFAULT true,
    notes TEXT,
    recorded_by UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (gym_date, fighter_id)
);

CREATE INDEX IF NOT EXISTS idx_attendance_gym_date ON attendance_records (gym_date);

CREATE TABLE IF NOT EXISTS announcements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    expires_at TIMESTAMPTZ,
    created_by UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS fighter_assistant_trainers (
    fighter_id UUID NOT NULL REFERENCES fighters (id) ON DELETE CASCADE,
    trainer_id UUID NOT NULL REFERENCES trainers (id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'assistant' CHECK (role IN ('assistant', 'corner')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (fighter_id, trainer_id)
);

CREATE INDEX IF NOT EXISTS idx_fighter_assistant_trainer ON fighter_assistant_trainers (trainer_id);
