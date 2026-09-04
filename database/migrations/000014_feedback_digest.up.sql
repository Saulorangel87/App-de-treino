ALTER TABLE user_feedback
    ADD COLUMN digest_sent_at TIMESTAMPTZ;

CREATE INDEX idx_user_feedback_digest_pending
    ON user_feedback (created_at ASC)
    WHERE digest_sent_at IS NULL;
