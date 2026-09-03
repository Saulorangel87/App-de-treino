CREATE TABLE user_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category TEXT NOT NULL CHECK (category IN ('experience', 'bug', 'suggestion')),
    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    message TEXT NOT NULL CHECK (char_length(btrim(message)) BETWEEN 10 AND 2000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_feedback_created_at ON user_feedback (created_at DESC);
CREATE INDEX idx_user_feedback_user_created_at ON user_feedback (user_id, created_at DESC);
