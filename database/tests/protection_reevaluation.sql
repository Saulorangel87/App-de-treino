-- Reavaliação da proteção (backend/internal/repository/planning_reevaluation.go).
-- As instruções abaixo repetem as do repositório: se elas mudarem, esta fixture
-- precisa mudar junto.
BEGIN;

DO $$
DECLARE
    user_a UUID; user_b UUID; user_c UUID;
    profile_a UUID; profile_b UUID; profile_c UUID;
    plan_a UUID; plan_a_draft UUID; plan_b UUID; plan_c UUID;
    w_planned UUID; w_adapted UUID; w_completed UUID; w_in_draft UUID; w_other_athlete UUID;
    src_pain_ineligible UUID; src_fatigue_ineligible UUID; src_old_pain UUID; src_fatigue_eligible UUID;
    session_id UUID;
    affected BIGINT;
    n INTEGER;
    pain_seen BOOLEAN;
    fatigue_seen INTEGER;
BEGIN
    INSERT INTO users (email, password_hash, display_name) VALUES ('reeval-a@example.invalid', 'test-only', 'A') RETURNING id INTO user_a;
    INSERT INTO users (email, password_hash, display_name) VALUES ('reeval-b@example.invalid', 'test-only', 'B') RETURNING id INTO user_b;
    INSERT INTO users (email, password_hash, display_name) VALUES ('reeval-c@example.invalid', 'test-only', 'C') RETURNING id INTO user_c;
    INSERT INTO athlete_profiles (user_id, experience_level) VALUES (user_a, 'advanced') RETURNING id INTO profile_a;
    INSERT INTO athlete_profiles (user_id, experience_level) VALUES (user_b, 'advanced') RETURNING id INTO profile_b;
    INSERT INTO athlete_profiles (user_id, experience_level) VALUES (user_c, 'advanced') RETURNING id INTO profile_c;

    -- Atleta A: plano ativo com treinos em vários estados e um rascunho.
    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status) VALUES (profile_a, DATE '2099-08-03', DATE '2099-08-30', 'active') RETURNING id INTO plan_a;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (plan_a, DATE '2099-08-05', 'Planejado', 'T', 60, 5, '{}'::jsonb, '{}'::jsonb, 'planned') RETURNING id INTO w_planned;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (plan_a, DATE '2099-08-06', 'Adaptado pelo check-in', 'T', 60, 5, '{}'::jsonb, '{}'::jsonb, 'adapted') RETURNING id INTO w_adapted;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (plan_a, DATE '2099-08-04', 'Concluído', 'T', 60, 5, '{}'::jsonb, '{}'::jsonb, 'completed') RETURNING id INTO w_completed;
    -- Um plano em rascunho não pode ser reescrito pela reavaliação.
    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status) VALUES (profile_a, DATE '2099-09-07', DATE '2099-10-04', 'draft') RETURNING id INTO plan_a_draft;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (plan_a_draft, DATE '2099-09-08', 'No rascunho', 'T', 60, 5, '{}'::jsonb, '{}'::jsonb, 'planned') RETURNING id INTO w_in_draft;
    -- Atleta B.
    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status) VALUES (profile_b, DATE '2099-08-03', DATE '2099-08-30', 'active') RETURNING id INTO plan_b;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (plan_b, DATE '2099-08-05', 'Do outro atleta', 'T', 60, 5, '{}'::jsonb, '{}'::jsonb, 'planned') RETURNING id INTO w_other_athlete;

    -- 1) O UPDATE do repositório só reescreve treino planejado do plano ativo do próprio atleta.
    UPDATE workouts w SET name = 'Reescrito', objective = 'T', duration_minutes = 30, target_rpe = 3.5,
        structure = '{}'::jsonb, explanation = '{"protection":{"level":"strong"}}'::jsonb
    FROM training_plans tp JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
    WHERE w.id = w_planned AND w.training_plan_id = tp.id AND ap.user_id = user_a AND tp.status = 'active' AND w.status = 'planned';
    GET DIAGNOSTICS affected = ROW_COUNT;
    IF affected <> 1 THEN RAISE EXCEPTION 'a planned workout of the own active plan must be rewritten, got %', affected; END IF;

    FOR n IN 1..4 LOOP
        UPDATE workouts w SET name = 'Indevido'
        FROM training_plans tp JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
        WHERE w.id = (ARRAY[w_adapted, w_completed, w_in_draft, w_other_athlete])[n]
          AND w.training_plan_id = tp.id AND ap.user_id = user_a AND tp.status = 'active' AND w.status = 'planned';
        GET DIAGNOSTICS affected = ROW_COUNT;
        IF affected <> 0 THEN RAISE EXCEPTION 'workout % must never be rewritten (adapted, completed, draft plan, other athlete), got %', n, affected; END IF;
    END LOOP;
    IF (SELECT count(*) FROM workouts WHERE name = 'Indevido') <> 0 THEN RAISE EXCEPTION 'an untouchable workout was rewritten'; END IF;

    -- 2) A leitura da reavaliação só devolve treinos planejados do plano ativo, dentro da janela.
    SELECT count(*) INTO n
    FROM workouts w
    JOIN training_plans tp ON tp.id = w.training_plan_id
    JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
    WHERE ap.user_id = user_a AND tp.status = 'active' AND w.status = 'planned'
      AND w.scheduled_on BETWEEN DATE '2099-08-01' AND DATE '2099-08-31';
    IF n <> 1 THEN RAISE EXCEPTION 'only the planned workout of the active plan should be listed, got %', n; END IF;
    SELECT count(*) INTO n
    FROM workouts w
    JOIN training_plans tp ON tp.id = w.training_plan_id
    JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
    WHERE ap.user_id = user_a AND tp.status = 'active' AND w.status = 'planned'
      AND w.scheduled_on BETWEEN DATE '2099-08-06' AND DATE '2099-08-31';
    IF n <> 0 THEN RAISE EXCEPTION 'a workout outside the window must not be listed, got %', n; END IF;

    -- 3) Sinais datados: dor sempre conta; fadiga respeita a integridade; a janela é de 14 dias.
    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status) VALUES (profile_c, DATE '2099-08-03', DATE '2099-08-30', 'active') RETURNING id INTO plan_c;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (plan_c, DATE '2099-08-03', 'Dor, dados incompletos', 'T', 60, 5, '{"x":1}'::jsonb, '{"data_integrity":{"eligible_for_history":false}}'::jsonb, 'completed') RETURNING id INTO src_pain_ineligible;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (plan_c, DATE '2099-08-04', 'Fadiga, dados incompletos', 'T', 60, 5, '{}'::jsonb, '{"data_integrity":{"eligible_for_history":false}}'::jsonb, 'completed') RETURNING id INTO src_fatigue_ineligible;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (plan_c, DATE '2099-08-05', 'Dor antiga', 'T', 60, 5, '{}'::jsonb, '{}'::jsonb, 'completed') RETURNING id INTO src_old_pain;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (plan_c, DATE '2099-08-06', 'Fadiga, dados completos', 'T', 60, 5, '{}'::jsonb, '{"data_integrity":{"eligible_for_history":true}}'::jsonb, 'completed') RETURNING id INTO src_fatigue_eligible;

    INSERT INTO workout_sessions (workout_id, athlete_profile_id, started_at, completed_at, duration_minutes, actual_rpe, status)
    VALUES (src_pain_ineligible, profile_c, now() - interval '2 days 1 hour', now() - interval '2 days', 60, 5, 'completed') RETURNING id INTO session_id;
    INSERT INTO feedback (workout_session_id, difficulty, pain_reported, fatigue_after) VALUES (session_id, 'moderate', true, 3);
    INSERT INTO workout_sessions (workout_id, athlete_profile_id, started_at, completed_at, duration_minutes, actual_rpe, status)
    VALUES (src_fatigue_ineligible, profile_c, now() - interval '3 days 1 hour', now() - interval '3 days', 60, 5, 'completed') RETURNING id INTO session_id;
    INSERT INTO feedback (workout_session_id, difficulty, pain_reported, fatigue_after) VALUES (session_id, 'moderate', false, 5);
    INSERT INTO workout_sessions (workout_id, athlete_profile_id, started_at, completed_at, duration_minutes, actual_rpe, status)
    VALUES (src_old_pain, profile_c, now() - interval '20 days 1 hour', now() - interval '20 days', 60, 5, 'completed') RETURNING id INTO session_id;
    INSERT INTO feedback (workout_session_id, difficulty, pain_reported, fatigue_after) VALUES (session_id, 'moderate', true, 3);
    INSERT INTO workout_sessions (workout_id, athlete_profile_id, started_at, completed_at, duration_minutes, actual_rpe, status)
    VALUES (src_fatigue_eligible, profile_c, now() - interval '1 day 1 hour', now() - interval '1 day', 60, 5, 'completed') RETURNING id INTO session_id;
    INSERT INTO feedback (workout_session_id, difficulty, pain_reported, fatigue_after) VALUES (session_id, 'moderate', false, 4);

    SELECT count(*) INTO n FROM (
        SELECT ws.completed_at, COALESCE(f.pain_reported, false) AS pain,
            CASE WHEN source_workout.explanation->'data_integrity' IS NULL
                OR source_workout.explanation->'data_integrity'->>'eligible_for_history' = 'true'
            THEN COALESCE(f.fatigue_after, 0) ELSE 0 END AS fatigue
        FROM workout_sessions ws
        JOIN workouts source_workout ON source_workout.id = ws.workout_id
        LEFT JOIN feedback f ON f.workout_session_id = ws.id
        WHERE ws.athlete_profile_id = profile_c AND ws.status = 'completed'
          AND ws.completed_at >= now() - make_interval(days => 14) AND ws.completed_at <= now()
    ) signals;
    IF n <> 3 THEN RAISE EXCEPTION 'the 20-day-old session must be outside the 14-day window, got % rows', n; END IF;

    SELECT COALESCE(f.pain_reported, false),
        CASE WHEN source_workout.explanation->'data_integrity' IS NULL
            OR source_workout.explanation->'data_integrity'->>'eligible_for_history' = 'true'
        THEN COALESCE(f.fatigue_after, 0) ELSE 0 END
    INTO pain_seen, fatigue_seen
    FROM workout_sessions ws JOIN workouts source_workout ON source_workout.id = ws.workout_id
    LEFT JOIN feedback f ON f.workout_session_id = ws.id WHERE ws.workout_id = src_pain_ineligible;
    IF NOT pain_seen THEN RAISE EXCEPTION 'a reported pain must count even when the integrity gate rejects the session'; END IF;

    SELECT CASE WHEN source_workout.explanation->'data_integrity' IS NULL
            OR source_workout.explanation->'data_integrity'->>'eligible_for_history' = 'true'
        THEN COALESCE(f.fatigue_after, 0) ELSE 0 END
    INTO fatigue_seen
    FROM workout_sessions ws JOIN workouts source_workout ON source_workout.id = ws.workout_id
    LEFT JOIN feedback f ON f.workout_session_id = ws.id WHERE ws.workout_id = src_fatigue_ineligible;
    IF fatigue_seen <> 0 THEN RAISE EXCEPTION 'fatigue from a session rejected by the integrity gate must not count, got %', fatigue_seen; END IF;

    SELECT CASE WHEN source_workout.explanation->'data_integrity' IS NULL
            OR source_workout.explanation->'data_integrity'->>'eligible_for_history' = 'true'
        THEN COALESCE(f.fatigue_after, 0) ELSE 0 END
    INTO fatigue_seen
    FROM workout_sessions ws JOIN workouts source_workout ON source_workout.id = ws.workout_id
    LEFT JOIN feedback f ON f.workout_session_id = ws.id WHERE ws.workout_id = src_fatigue_eligible;
    IF fatigue_seen <> 4 THEN RAISE EXCEPTION 'fatigue from an eligible session must count, got %', fatigue_seen; END IF;
END;
$$;

ROLLBACK;
