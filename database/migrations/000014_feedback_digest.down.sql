DROP INDEX IF EXISTS idx_user_feedback_digest_pending;

ALTER TABLE user_feedback
    DROP COLUMN IF EXISTS digest_sent_at;
