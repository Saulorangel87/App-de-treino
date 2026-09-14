ALTER TABLE feedback
    ADD COLUMN equipment_used TEXT,
    ADD CONSTRAINT feedback_equipment_used_check
        CHECK (equipment_used IS NULL OR char_length(equipment_used) <= 120);
