BEGIN;

DO $$
DECLARE
    test_user_id UUID;
    test_profile_id UUID;
    test_plan_id UUID;
    source_workout_id UUID;
    next_workout_id UUID;
    untouched_workout_id UUID;
    test_session_id UUID;
    adapted_duration INTEGER;
    adapted_rpe NUMERIC(3,1);
    adapted_reason TEXT;
    untouched_duration INTEGER;
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('adaptation-trigger-test@example.invalid', 'test-only', 'Adaptation Test')
    RETURNING id INTO test_user_id;

    INSERT INTO athlete_profiles (user_id, experience_level)
    VALUES (test_user_id, 'intermediate')
    RETURNING id INTO test_profile_id;

    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status)
    VALUES (test_profile_id, DATE '2099-01-05', DATE '2099-02-01', 'active')
    RETURNING id INTO test_plan_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, DATE '2099-01-05', 'Fonte', 'Teste', 50,
        5, '{}'::jsonb, '{"rules":[]}'::jsonb, 'completed'
    ) RETURNING id INTO source_workout_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, DATE '2099-01-07', 'Próximo', 'Teste', 60,
        6, '{}'::jsonb, '{"rules":[]}'::jsonb, 'planned'
    ) RETURNING id INTO next_workout_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, DATE '2099-01-09', 'Posterior', 'Teste', 70,
        6, '{}'::jsonb, '{"rules":[]}'::jsonb, 'planned'
    ) RETURNING id INTO untouched_workout_id;

    INSERT INTO workout_sessions (
        workout_id, athlete_profile_id, started_at, completed_at,
        duration_minutes, actual_rpe, status
    ) VALUES (
        source_workout_id, test_profile_id, now() - interval '50 minutes', now(),
        50, 7, 'completed'
    ) RETURNING id INTO test_session_id;

    INSERT INTO feedback (
        workout_session_id, difficulty, pain_reported, fatigue_after, notes
    ) VALUES (
        test_session_id, 'hard', false, 4, 'Teste transacional.'
    );

    SELECT duration_minutes, target_rpe, explanation#>>'{adaptation,reason}'
    INTO adapted_duration, adapted_rpe, adapted_reason
    FROM workouts WHERE id = next_workout_id;

    SELECT duration_minutes INTO untouched_duration
    FROM workouts WHERE id = untouched_workout_id;

    IF adapted_duration <> 54 OR adapted_rpe <> 5 OR adapted_reason IS NULL THEN
        RAISE EXCEPTION 'unexpected adapted workout: duration %, rpe %, reason %',
            adapted_duration, adapted_rpe, adapted_reason;
    END IF;
    IF untouched_duration <> 70 THEN
        RAISE EXCEPTION 'only one workout should be adapted, got %', untouched_duration;
    END IF;
END;
$$;

ROLLBACK;
