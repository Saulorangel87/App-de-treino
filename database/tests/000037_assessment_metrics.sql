-- Avaliação com números (000037): colunas opcionais, limites e as duas metades
-- da frequência cardíaca sempre juntas.
BEGIN;

DO $$
DECLARE
    test_user_id UUID;
    test_profile_id UUID;
    old_row UUID;
    stored_hr INTEGER;
    rejected BOOLEAN;
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('assessment-metrics-test@example.invalid', 'test-only', 'Assessment Metrics Test')
    RETURNING id INTO test_user_id;
    INSERT INTO athlete_profiles (user_id, experience_level) VALUES (test_user_id, 'intermediate')
    RETURNING id INTO test_profile_id;

    -- Uma avaliação sem números (como as anteriores à migração) continua válida.
    INSERT INTO cycling_assessments (athlete_profile_id, duration_minutes, actual_rpe, eligible_for_progression)
    VALUES (test_profile_id, 20, 4.5, true) RETURNING id INTO old_row;
    SELECT average_heart_rate INTO stored_hr FROM cycling_assessments WHERE id = old_row;
    IF stored_hr IS NOT NULL THEN
        RAISE EXCEPTION 'avaliação sem números deveria ter frequência cardíaca nula, veio %', stored_hr;
    END IF;

    -- Todos os números preenchidos.
    INSERT INTO cycling_assessments (athlete_profile_id, duration_minutes, actual_rpe, average_heart_rate,
        average_power_w, distance_km, heart_rate_first_half, heart_rate_second_half)
    VALUES (test_profile_id, 20, 4.5, 138, 150, 9.5, 135, 141);

    -- As duas metades andam juntas.
    rejected := false;
    BEGIN
        INSERT INTO cycling_assessments (athlete_profile_id, duration_minutes, actual_rpe, heart_rate_first_half)
        VALUES (test_profile_id, 20, 4.5, 135);
    EXCEPTION WHEN check_violation THEN rejected := true;
    END;
    IF NOT rejected THEN RAISE EXCEPTION 'uma metade sozinha deveria ser recusada'; END IF;

    -- Limites.
    rejected := false;
    BEGIN
        INSERT INTO cycling_assessments (athlete_profile_id, duration_minutes, actual_rpe, average_heart_rate)
        VALUES (test_profile_id, 20, 4.5, 20);
    EXCEPTION WHEN check_violation THEN rejected := true;
    END;
    IF NOT rejected THEN RAISE EXCEPTION 'frequência cardíaca de 20 bpm deveria ser recusada'; END IF;

    rejected := false;
    BEGIN
        INSERT INTO cycling_assessments (athlete_profile_id, duration_minutes, actual_rpe, average_power_w)
        VALUES (test_profile_id, 20, 4.5, 2500);
    EXCEPTION WHEN check_violation THEN rejected := true;
    END;
    IF NOT rejected THEN RAISE EXCEPTION 'potência de 2500 W deveria ser recusada'; END IF;
END $$;

ROLLBACK;
