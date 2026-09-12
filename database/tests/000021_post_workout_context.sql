BEGIN;

DO $$
DECLARE
    test_user_id UUID;
    test_profile_id UUID;
    test_plan_id UUID;
    test_workout_id UUID;
    test_session_id UUID;
    invalid_session_id UUID;
    stored_recovery SMALLINT;
    stored_confidence SMALLINT;
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('post-workout-context-test@example.invalid', 'test-only', 'Post Workout Context Test')
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
        test_plan_id, DATE '2099-04-01', 'Contexto pós-treino', 'Teste', 45,
        6, '{}'::jsonb, '{"rules":[]}'::jsonb, 'completed'
    ) RETURNING id INTO test_workout_id;

    INSERT INTO workout_sessions (
        workout_id, athlete_profile_id, started_at, completed_at,
        duration_minutes, actual_rpe, status
    ) VALUES (
        test_workout_id, test_profile_id, now() - interval '45 minutes', now(),
        45, 6, 'completed'
    ) RETURNING id INTO test_session_id;

    INSERT INTO workout_sessions (
        workout_id, athlete_profile_id, started_at, completed_at,
        duration_minutes, actual_rpe, status
    ) VALUES (
        test_workout_id, test_profile_id, now() - interval '40 minutes', now(),
        40, 6, 'completed'
    ) RETURNING id INTO invalid_session_id;

    INSERT INTO feedback (
        workout_session_id, completion_status, difficulty, pain_reported,
        fatigue_after, recovery_after, repeat_confidence
    ) VALUES (
        test_session_id, 'complete', 'moderate', false,
        3, 4, 5
    );

    SELECT recovery_after, repeat_confidence
    INTO stored_recovery, stored_confidence
    FROM feedback
    WHERE workout_session_id = test_session_id;

    IF stored_recovery <> 4 OR stored_confidence <> 5 THEN
        RAISE EXCEPTION 'post-workout context was not stored: recovery %, confidence %',
            stored_recovery, stored_confidence;
    END IF;

    BEGIN
        INSERT INTO feedback (
            workout_session_id, completion_status, difficulty, pain_reported,
            fatigue_after, recovery_after, repeat_confidence
        ) VALUES (
            invalid_session_id, 'complete', 'moderate', false,
            3, 0, 3
        );
        RAISE EXCEPTION 'recovery_after outside the range should fail';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;

    BEGIN
        INSERT INTO feedback (
            workout_session_id, completion_status, difficulty, pain_reported,
            fatigue_after, recovery_after, repeat_confidence
        ) VALUES (
            invalid_session_id, 'complete', 'moderate', false,
            3, 3, 6
        );
        RAISE EXCEPTION 'repeat_confidence outside the range should fail';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;
END;
$$;

ROLLBACK;
