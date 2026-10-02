-- Avaliação com números (000037): dados opcionais do pedal de referência, para
-- mostrar eficiência aeróbica e deriva de frequência cardíaca e comparar uma
-- avaliação com a anterior. A regra "apto a progredir" não muda.
ALTER TABLE cycling_assessments
    ADD COLUMN average_heart_rate INTEGER CHECK (average_heart_rate BETWEEN 30 AND 250),
    ADD COLUMN average_power_w INTEGER CHECK (average_power_w BETWEEN 0 AND 2000),
    ADD COLUMN distance_km NUMERIC(6,2) CHECK (distance_km BETWEEN 0 AND 500),
    ADD COLUMN heart_rate_first_half INTEGER CHECK (heart_rate_first_half BETWEEN 30 AND 250),
    ADD COLUMN heart_rate_second_half INTEGER CHECK (heart_rate_second_half BETWEEN 30 AND 250),
    ADD CONSTRAINT cycling_assessments_halves_together
        CHECK ((heart_rate_first_half IS NULL) = (heart_rate_second_half IS NULL));
