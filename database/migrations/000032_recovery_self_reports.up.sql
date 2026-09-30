-- Declaração do atleta de que está recuperado ("Estou recuperado"). É um registro
-- próprio, separado do check-in diário, para não sobrescrever o que o atleta
-- respondeu no dia. O motor trata a declaração como um check-in bom na
-- reavaliação da proteção; ela nunca remove uma proteção forte recente por dor.
-- Ver docs/motor-protecao-cenarios.md.
CREATE TABLE recovery_self_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_profile_id UUID NOT NULL REFERENCES athlete_profiles(id) ON DELETE CASCADE,
    reported_on DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (athlete_profile_id, reported_on)
);
