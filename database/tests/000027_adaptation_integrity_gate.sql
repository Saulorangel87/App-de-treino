BEGIN;

DO $$
DECLARE
    function_source TEXT;
BEGIN
    SELECT pg_get_functiondef(p.oid)
    INTO function_source
    FROM pg_proc p
    JOIN pg_namespace n ON n.oid = p.pronamespace
    WHERE n.nspname = current_schema()
      AND p.proname = 'adapt_training_plan_after_feedback'
    LIMIT 1;

    IF position('NEW.completion_status <> ''complete''' IN function_source) = 0
       OR position('actual_duration IS NULL' IN function_source) = 0
       OR position('w_recent.scheduled_on >= source_date - 14' IN function_source) = 0 THEN
        RAISE EXCEPTION 'adaptation integrity guard is not installed';
    END IF;
END;
$$;

ROLLBACK;

BEGIN;

DO $$
DECLARE
    test_user_id UUID;
    test_profile_id UUID;
    test_plan_id UUID;
    prior_workout_id UUID;
    source_workout_id UUID;
    next_workout_id UUID;
    prior_session_id UUID;
    source_session_id UUID;
    next_duration INTEGER;
    next_rpe NUMERIC(3,1);
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('adaptation-protection-test@example.invalid', 'test-only', 'Protection Gate Test')
    RETURNING id INTO test_user_id;

    INSERT INTO athlete_profiles (user_id, experience_level)
    VALUES (test_user_id, 'intermediate')
    RETURNING id INTO test_profile_id;

    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status)
    VALUES (test_profile_id, DATE '2099-03-01', DATE '2099-03-28', 'active')
    RETURNING id INTO test_plan_id;

    -- The protective signal is recorded before the source workout exists, so
    -- it cannot adapt the later candidate by itself.
    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, DATE '2099-03-01', 'Proteção anterior', 'Teste', 40,
        5, '{}'::jsonb, '{"rules":[]}'::jsonb, 'completed'
    ) RETURNING id INTO prior_workout_id;

    INSERT INTO workout_sessions (
        workout_id, athlete_profile_id, started_at, completed_at,
        duration_minutes, actual_rpe, status
    ) VALUES (
        prior_workout_id, test_profile_id, now() - interval '40 minutes', now(),
        40, 7, 'completed'
    ) RETURNING id INTO prior_session_id;

    INSERT INTO feedback (
        workout_session_id, difficulty, pain_reported, fatigue_after, notes
    ) VALUES (
        prior_session_id, 'hard', false, 4, 'Sinal protetivo anterior.'
    );

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, DATE '2099-03-05', 'Fonte atual', 'Teste', 40,
        5, '{}'::jsonb, '{"rules":[]}'::jsonb, 'completed'
    ) RETURNING id INTO source_workout_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, DATE '2099-03-07', 'Próximo protegido', 'Teste', 60,
        6, '{}'::jsonb, '{"rules":[]}'::jsonb, 'planned'
    ) RETURNING id INTO next_workout_id;

    INSERT INTO workout_sessions (
        workout_id, athlete_profile_id, started_at, completed_at,
        duration_minutes, actual_rpe, status
    ) VALUES (
        source_workout_id, test_profile_id, now() - interval '30 minutes', now(),
        30, 3, 'completed'
    ) RETURNING id INTO source_session_id;

    INSERT INTO feedback (
        workout_session_id, difficulty, pain_reported, fatigue_after, notes
    ) VALUES (
        source_session_id, 'very_easy', false, 2, 'Resposta fácil depois de sinal protetivo.'
    );

    SELECT duration_minutes, target_rpe
    INTO next_duration, next_rpe
    FROM workouts WHERE id = next_workout_id;

    IF next_duration <> 60 OR next_rpe <> 6 THEN
        RAISE EXCEPTION 'recent protective signal must block progression: duration %, rpe %',
            next_duration, next_rpe;
    END IF;
END;
$$;

ROLLBACK;

BEGIN;

DO $$
DECLARE
    test_user_id UUID;
    test_profile_id UUID;
    test_plan_id UUID;
    source_workout_id UUID;
    next_workout_id UUID;
    source_session_id UUID;
    next_duration INTEGER;
    next_rpe NUMERIC(3,1);
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('adaptation-missing-duration-test@example.invalid', 'test-only', 'Duration Gate Test')
    RETURNING id INTO test_user_id;

    INSERT INTO athlete_profiles (user_id, experience_level)
    VALUES (test_user_id, 'intermediate')
    RETURNING id INTO test_profile_id;

    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status)
    VALUES (test_profile_id, DATE '2099-04-01', DATE '2099-04-28', 'active')
    RETURNING id INTO test_plan_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, DATE '2099-04-01', 'Fonte sem duração', 'Teste', 40,
        5, '{}'::jsonb, '{"rules":[]}'::jsonb, 'completed'
    ) RETURNING id INTO source_workout_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, DATE '2099-04-03', 'Próximo sem progressão', 'Teste', 60,
        6, '{}'::jsonb, '{"rules":[]}'::jsonb, 'planned'
    ) RETURNING id INTO next_workout_id;

    INSERT INTO workout_sessions (
        workout_id, athlete_profile_id, started_at, completed_at,
        duration_minutes, actual_rpe, status
    ) VALUES (
        source_workout_id, test_profile_id, now() - interval '20 minutes', now(),
        NULL, 3, 'completed'
    ) RETURNING id INTO source_session_id;

    INSERT INTO feedback (
        workout_session_id, difficulty, pain_reported, fatigue_after, notes
    ) VALUES (
        source_session_id, 'very_easy', false, 2, 'Duração ausente.'
    );

    SELECT duration_minutes, target_rpe
    INTO next_duration, next_rpe
    FROM workouts WHERE id = next_workout_id;

    IF next_duration <> 60 OR next_rpe <> 6 THEN
        RAISE EXCEPTION 'missing duration must block adaptation: duration %, rpe %',
            next_duration, next_rpe;
    END IF;
END;
$$;

ROLLBACK;
