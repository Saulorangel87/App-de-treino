ALTER TABLE feedback
    DROP CONSTRAINT IF EXISTS feedback_external_conditions_check,
    DROP CONSTRAINT IF EXISTS feedback_terrain_check,
    DROP CONSTRAINT IF EXISTS feedback_satisfaction_check,
    DROP COLUMN IF EXISTS external_conditions,
    DROP COLUMN IF EXISTS terrain,
    DROP COLUMN IF EXISTS satisfaction;
