BEGIN;

DO $$
DECLARE
    test_user_id UUID;
    test_profile_id UUID;
    test_plan_id UUID;
    source_workout_id UUID;
    next_workout_id UUID;
    test_session_id UUID;
    invalid_session_id UUID;
    next_duration INTEGER;
    next_rpe NUMERIC(3,1);
    adaptation_reason TEXT;
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('completion-context-test@example.invalid', 'test-only', 'Completion Context Test')
    RETURNING id INTO test_user_id;

    INSERT INTO athlete_profiles (user_id, experience_level)
    VALUES (test_user_id, 'intermediate')
    RETURNING id INTO test_profile_id;

    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status)
    VALUES (test_profile_id, DATE '2099-02-02', DATE '2099-03-01', 'active')
    RETURNING id INTO test_plan_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, DATE '2099-02-02', 'Fonte parcial', 'Teste', 50,
        6, '{}'::jsonb, '{"rules":[]}'::jsonb, 'completed'
    ) RETURNING id INTO source_workout_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, DATE '2099-02-04', 'Próximo treino', 'Teste', 60,
        6, '{}'::jsonb, '{"rules":[]}'::jsonb, 'planned'
    ) RETURNING id INTO next_workout_id;

    INSERT INTO workout_sessions (
        workout_id, athlete_profile_id, started_at, completed_at,
        duration_minutes, actual_rpe, status
    ) VALUES (
        source_workout_id, test_profile_id, now() - interval '50 minutes', now(),
        50, 4, 'completed'
    ) RETURNING id INTO test_session_id;

    INSERT INTO workout_sessions (
        workout_id, athlete_profile_id, started_at, completed_at,
        duration_minutes, actual_rpe, status
    ) VALUES (
        source_workout_id, test_profile_id, now() - interval '40 minutes', now(),
        40, 4, 'completed'
    ) RETURNING id INTO invalid_session_id;

    INSERT INTO feedback (
        workout_session_id, completion_status, partial_reason,
        difficulty, pain_reported, fatigue_after, notes
    ) VALUES (
        test_session_id, 'partial', 'time_available_changed',
        'easy', false, 2, 'Sessão interrompida por falta de tempo.'
    );

    SELECT duration_minutes, target_rpe, explanation#>>'{adaptation,reason}'
    INTO next_duration, next_rpe, adaptation_reason
    FROM workouts WHERE id = next_workout_id;

    IF next_duration <> 60 OR next_rpe <> 6 OR adaptation_reason IS NOT NULL THEN
        RAISE EXCEPTION 'partial completion must not progress workout: duration %, rpe %, reason %',
            next_duration, next_rpe, adaptation_reason;
    END IF;

    BEGIN
        INSERT INTO feedback (
            workout_session_id, completion_status, partial_reason,
            difficulty, pain_reported, fatigue_after
        ) VALUES (
            invalid_session_id, 'partial', NULL,
            'easy', false, 2
        );
        RAISE EXCEPTION 'partial completion without reason should fail';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;
END;
$$;

ROLLBACK;
