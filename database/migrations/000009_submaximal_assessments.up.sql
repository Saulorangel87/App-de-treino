CREATE TABLE cycling_assessments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_profile_id UUID NOT NULL REFERENCES athlete_profiles(id) ON DELETE CASCADE,
    assessment_type TEXT NOT NULL DEFAULT 'submax_reference' CHECK (assessment_type = 'submax_reference'),
    completed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    duration_minutes INTEGER NOT NULL CHECK (duration_minutes BETWEEN 15 AND 30),
    target_rpe NUMERIC(3,1) NOT NULL DEFAULT 5.0 CHECK (target_rpe BETWEEN 1 AND 10),
    actual_rpe NUMERIC(3,1) NOT NULL CHECK (actual_rpe BETWEEN 1 AND 10),
    pain_reported BOOLEAN NOT NULL DEFAULT false,
    notes TEXT NOT NULL DEFAULT '' CHECK (char_length(notes) <= 1000),
    eligible_for_progression BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_cycling_assessments_profile_completed
    ON cycling_assessments (athlete_profile_id, completed_at DESC);
