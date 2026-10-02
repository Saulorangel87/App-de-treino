-- Modo tarefa (000036): origem da duração por sessão e progressão que exige pelo
-- menos 80% do tempo planejado.
BEGIN;

DO $$
DECLARE
    test_user_id UUID;
    test_profile_id UUID;
    test_plan_id UUID;
    src_a UUID;
    src_a2 UUID;
    src_b UUID;
    src_c UUID;
    next_one UUID;
    next_two UUID;
    default_source TEXT;
    duration_one INTEGER;
    duration_two INTEGER;
    adapted_one BOOLEAN;
    adapted_two BOOLEAN;
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('task-mode-test@example.invalid', 'test-only', 'Task Mode Test')
    RETURNING id INTO test_user_id;
    INSERT INTO athlete_profiles (user_id, experience_level) VALUES (test_user_id, 'intermediate')
    RETURNING id INTO test_profile_id;
    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status)
    VALUES (test_profile_id, CURRENT_DATE - 10, CURRENT_DATE + 10, 'active') RETURNING id INTO test_plan_id;

    -- Quatro treinos já feitos (planejados com 90 min e esforço 6) e dois por vir.
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (test_plan_id, CURRENT_DATE - 6, 'A', 'Teste', 90, 6, '{}', '{"rules":[]}', 'completed') RETURNING id INTO src_a;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (test_plan_id, CURRENT_DATE - 5, 'A2', 'Teste', 90, 6, '{}', '{"rules":[]}', 'completed') RETURNING id INTO src_a2;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (test_plan_id, CURRENT_DATE - 4, 'B', 'Teste', 90, 6, '{}', '{"rules":[]}', 'completed') RETURNING id INTO src_b;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (test_plan_id, CURRENT_DATE - 3, 'C', 'Teste', 90, 6, '{}', '{"rules":[]}', 'completed') RETURNING id INTO src_c;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (test_plan_id, CURRENT_DATE + 1, 'N1', 'Teste', 60, 5, '{}', '{"rules":[]}', 'planned') RETURNING id INTO next_one;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (test_plan_id, CURRENT_DATE + 2, 'N2', 'Teste', 60, 5, '{}', '{"rules":[]}', 'planned') RETURNING id INTO next_two;

    -- A origem da duração começa como cronômetro e só aceita valores conhecidos.
    INSERT INTO workout_sessions (workout_id, athlete_profile_id, started_at, completed_at, duration_minutes, actual_rpe, status)
    VALUES (src_a, test_profile_id, now() - interval '2 hours', now() - interval '30 minutes', 30, 4, 'completed');
    SELECT duration_source INTO default_source FROM workout_sessions WHERE workout_id = src_a;
    IF default_source <> 'timer' THEN
        RAISE EXCEPTION 'duration_source must default to timer, got %', default_source;
    END IF;
    BEGIN
        INSERT INTO workout_sessions (workout_id, athlete_profile_id, started_at, completed_at, duration_minutes, actual_rpe, status, duration_source)
        VALUES (src_a2, test_profile_id, now() - interval '2 hours', now(), 71, 4, 'completed', 'adivinhada');
        RAISE EXCEPTION 'an unknown duration_source was accepted';
    EXCEPTION WHEN check_violation THEN NULL;
    END;
    DELETE FROM workout_sessions WHERE workout_id = src_a;

    -- A: 30 de 90 minutos (33%), fácil e sem fadiga: não progride.
    INSERT INTO workout_sessions (workout_id, athlete_profile_id, started_at, completed_at, duration_minutes, actual_rpe, status, duration_source)
    VALUES (src_a, test_profile_id, now() - interval '2 hours', now() - interval '30 minutes', 30, 4, 'completed', 'reported');
    INSERT INTO feedback (workout_session_id, completion_status, difficulty, pain_reported, fatigue_after)
    SELECT id, 'complete', 'easy', false, 2 FROM workout_sessions WHERE workout_id = src_a;
    -- A2: 71 de 90 (78,9%), logo abaixo do limite: não progride.
    INSERT INTO workout_sessions (workout_id, athlete_profile_id, started_at, completed_at, duration_minutes, actual_rpe, status, duration_source)
    VALUES (src_a2, test_profile_id, now() - interval '3 hours', now() - interval '1 hour', 71, 4, 'completed', 'reported');
    INSERT INTO feedback (workout_session_id, completion_status, difficulty, pain_reported, fatigue_after)
    SELECT id, 'complete', 'easy', false, 2 FROM workout_sessions WHERE workout_id = src_a2;

    SELECT duration_minutes, explanation ? 'adaptation' INTO duration_one, adapted_one FROM workouts WHERE id = next_one;
    IF duration_one <> 60 OR adapted_one THEN
        RAISE EXCEPTION 'a short easy session must not progress the next workout (got % min, adapted=%)', duration_one, adapted_one;
    END IF;

    -- B: 72 de 90 (exatamente 80%): progride o próximo treino planejado em 5%.
    INSERT INTO workout_sessions (workout_id, athlete_profile_id, started_at, completed_at, duration_minutes, actual_rpe, status, duration_source)
    VALUES (src_b, test_profile_id, now() - interval '4 hours', now() - interval '2 hours', 72, 4, 'completed', 'timer');
    INSERT INTO feedback (workout_session_id, completion_status, difficulty, pain_reported, fatigue_after)
    SELECT id, 'complete', 'easy', false, 2 FROM workout_sessions WHERE workout_id = src_b;
    SELECT duration_minutes, explanation #>> '{adaptation,kind}' = 'progression' INTO duration_one, adapted_one FROM workouts WHERE id = next_one;
    IF duration_one <> 63 OR NOT adapted_one THEN
        RAISE EXCEPTION '80%% of the planned time must progress the next workout to 63 min (got % min, progression=%)', duration_one, adapted_one;
    END IF;

    -- C: 90 de 90 (100%): progride o treino seguinte ainda planejado.
    INSERT INTO workout_sessions (workout_id, athlete_profile_id, started_at, completed_at, duration_minutes, actual_rpe, status)
    VALUES (src_c, test_profile_id, now() - interval '5 hours', now() - interval '3 hours', 90, 4, 'completed');
    INSERT INTO feedback (workout_session_id, completion_status, difficulty, pain_reported, fatigue_after)
    SELECT id, 'complete', 'easy', false, 2 FROM workout_sessions WHERE workout_id = src_c;
    SELECT duration_minutes, explanation #>> '{adaptation,kind}' = 'progression' INTO duration_two, adapted_two FROM workouts WHERE id = next_two;
    IF duration_two <> 63 OR NOT adapted_two THEN
        RAISE EXCEPTION 'a full session must progress the next planned workout (got % min, progression=%)', duration_two, adapted_two;
    END IF;
END;
$$;

ROLLBACK;
