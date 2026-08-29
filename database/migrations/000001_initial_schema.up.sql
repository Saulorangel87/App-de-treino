CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE athlete_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    birth_date DATE,
    sex TEXT CHECK (sex IN ('female', 'male', 'other', 'prefer_not_to_say')),
    height_cm NUMERIC(5,2),
    weight_kg NUMERIC(5,2),
    sport TEXT NOT NULL DEFAULT 'cycling',
    experience_level TEXT NOT NULL CHECK (experience_level IN ('beginner', 'intermediate', 'advanced')),
    activity_level TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_profile_id UUID NOT NULL REFERENCES athlete_profiles(id) ON DELETE CASCADE,
    goal_type TEXT NOT NULL,
    priority SMALLINT NOT NULL CHECK (priority IN (1, 2)),
    target_date DATE,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (athlete_profile_id, priority)
);

CREATE TABLE availability (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_profile_id UUID NOT NULL REFERENCES athlete_profiles(id) ON DELETE CASCADE,
    weekday SMALLINT NOT NULL CHECK (weekday BETWEEN 0 AND 6),
    available_minutes INTEGER NOT NULL CHECK (available_minutes >= 0),
    preferred_time TIME,
    location TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (athlete_profile_id, weekday)
);

CREATE TABLE recovery_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_profile_id UUID NOT NULL REFERENCES athlete_profiles(id) ON DELETE CASCADE,
    recorded_on DATE NOT NULL,
    sleep_minutes INTEGER CHECK (sleep_minutes >= 0),
    sleep_quality SMALLINT CHECK (sleep_quality BETWEEN 1 AND 5),
    stress_level SMALLINT CHECK (stress_level BETWEEN 1 AND 5),
    fatigue_level SMALLINT CHECK (fatigue_level BETWEEN 1 AND 5),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (athlete_profile_id, recorded_on)
);

CREATE TABLE injuries_or_limitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_profile_id UUID NOT NULL REFERENCES athlete_profiles(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    description TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    professional_clearance_recommended BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ
);

CREATE TABLE training_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_profile_id UUID NOT NULL REFERENCES athlete_profiles(id) ON DELETE CASCADE,
    starts_on DATE NOT NULL,
    ends_on DATE NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('draft', 'active', 'completed', 'cancelled')),
    prescription_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (ends_on >= starts_on)
);

CREATE TABLE workouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    training_plan_id UUID NOT NULL REFERENCES training_plans(id) ON DELETE CASCADE,
    scheduled_on DATE NOT NULL,
    name TEXT NOT NULL,
    objective TEXT NOT NULL,
    duration_minutes INTEGER NOT NULL CHECK (duration_minutes > 0),
    target_rpe NUMERIC(3,1) CHECK (target_rpe BETWEEN 0 AND 10),
    structure JSONB NOT NULL,
    explanation JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'planned' CHECK (status IN ('planned', 'completed', 'skipped', 'adapted')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE workout_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_id UUID NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    athlete_profile_id UUID NOT NULL REFERENCES athlete_profiles(id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_minutes INTEGER CHECK (duration_minutes >= 0),
    distance_km NUMERIC(8,2) CHECK (distance_km >= 0),
    average_power_watts INTEGER CHECK (average_power_watts >= 0),
    average_heart_rate INTEGER CHECK (average_heart_rate >= 0),
    actual_rpe NUMERIC(3,1) CHECK (actual_rpe BETWEEN 0 AND 10),
    metrics JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_session_id UUID NOT NULL UNIQUE REFERENCES workout_sessions(id) ON DELETE CASCADE,
    difficulty TEXT NOT NULL CHECK (difficulty IN ('very_easy', 'easy', 'moderate', 'hard', 'very_hard')),
    pain_reported BOOLEAN NOT NULL DEFAULT false,
    fatigue_after SMALLINT CHECK (fatigue_after BETWEEN 1 AND 5),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_recovery_profile_date ON recovery_data (athlete_profile_id, recorded_on DESC);
CREATE INDEX idx_workouts_plan_date ON workouts (training_plan_id, scheduled_on);
CREATE INDEX idx_sessions_profile_completed ON workout_sessions (athlete_profile_id, completed_at DESC);
