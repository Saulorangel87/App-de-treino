ALTER TABLE feedback
    ADD COLUMN recovery_after SMALLINT,
    ADD COLUMN repeat_confidence SMALLINT,
    ADD CONSTRAINT feedback_recovery_after_check
        CHECK (recovery_after IS NULL OR recovery_after BETWEEN 1 AND 5),
    ADD CONSTRAINT feedback_repeat_confidence_check
        CHECK (repeat_confidence IS NULL OR repeat_confidence BETWEEN 1 AND 5);
