ALTER TABLE feedback
    DROP CONSTRAINT IF EXISTS feedback_equipment_used_check,
    DROP COLUMN IF EXISTS equipment_used;
