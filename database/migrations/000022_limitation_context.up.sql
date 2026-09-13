ALTER TABLE injuries_or_limitations
    ADD COLUMN location TEXT NOT NULL DEFAULT '',
    ADD COLUMN pain_intensity SMALLINT CHECK (pain_intensity BETWEEN 1 AND 10),
    ADD COLUMN aggravating_movement TEXT NOT NULL DEFAULT '',
    ADD COLUMN started_on DATE;
