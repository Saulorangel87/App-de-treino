-- Registro bruto de atividades importadas de arquivo (.fit/.gpx), independente
-- de qualquer treino planejado. Nunca é lido pelo motor de prescrição; só vira
-- execução registrada quando o atleta confirma o vínculo pelos endpoints já
-- existentes de conclusão/correção de treino (workout_sessions). Ver
-- docs/proxima-fase-dados-reais.md.
CREATE TABLE imported_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workout_id UUID REFERENCES workouts(id) ON DELETE SET NULL,
    source TEXT NOT NULL CHECK (source IN ('fit', 'gpx')),
    file_hash TEXT NOT NULL CHECK (char_length(file_hash) = 64),
    started_at TIMESTAMPTZ NOT NULL,
    moving_seconds INTEGER NOT NULL CHECK (moving_seconds >= 0),
    distance_km NUMERIC(8,2) CHECK (distance_km >= 0),
    elevation_gain_m INTEGER CHECK (elevation_gain_m >= 0),
    average_heart_rate INTEGER CHECK (average_heart_rate >= 0),
    max_heart_rate INTEGER CHECK (max_heart_rate >= 0),
    average_power_watts INTEGER CHECK (average_power_watts >= 0),
    normalized_power_watts INTEGER CHECK (normalized_power_watts >= 0),
    average_cadence_rpm INTEGER CHECK (average_cadence_rpm >= 0),
    imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, file_hash)
);

CREATE INDEX idx_imported_activities_user_started_at ON imported_activities (user_id, started_at DESC);
CREATE INDEX idx_imported_activities_workout_id ON imported_activities (workout_id) WHERE workout_id IS NOT NULL;
