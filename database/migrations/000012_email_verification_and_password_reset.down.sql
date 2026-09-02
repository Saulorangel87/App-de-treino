DROP TABLE IF EXISTS auth_email_tokens;

ALTER TABLE users
    DROP COLUMN IF EXISTS email_verified_at;
