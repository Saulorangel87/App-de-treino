BEGIN;

DO $$
DECLARE
    test_user_id UUID;
    test_profile_id UUID;
    test_plan_id UUID;
    test_workout_id UUID;
    test_session_id UUID;
    invalid_session_id UUID;
    stored_satisfaction SMALLINT;
    stored_terrain TEXT;
    stored_conditions TEXT;
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('feedback-context-test@example.invalid', 'test-only', 'Feedback Context Test')
    RETURNING id INTO test_user_id;

    INSERT INTO athlete_profiles (user_id, experience_level)
    VALUES (test_user_id, 'intermediate')
    RETURNING id INTO test_profile_id;

    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status)
    VALUES (test_profile_id, DATE '2099-05-01', DATE '2099-05-28', 'active')
    RETURNING id INTO test_plan_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, DATE '2099-05-01', 'Contexto estruturado', 'Teste', 45,
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
        fatigue_after, satisfaction, terrain, external_conditions
    ) VALUES (
        test_session_id, 'complete', 'moderate', false,
        3, 4, 'rolling', 'wind'
    );

    SELECT satisfaction, terrain, external_conditions
    INTO stored_satisfaction, stored_terrain, stored_conditions
    FROM feedback
    WHERE workout_session_id = test_session_id;

    IF stored_satisfaction <> 4 OR stored_terrain <> 'rolling' OR stored_conditions <> 'wind' THEN
        RAISE EXCEPTION 'structured feedback context was not stored: satisfaction %, terrain %, conditions %',
            stored_satisfaction, stored_terrain, stored_conditions;
    END IF;

    BEGIN
        INSERT INTO feedback (
            workout_session_id, completion_status, difficulty, pain_reported,
            fatigue_after, satisfaction, terrain, external_conditions
        ) VALUES (
            invalid_session_id, 'complete', 'moderate', false,
            3, 0, 'rolling', 'wind'
        );
        RAISE EXCEPTION 'satisfaction outside the range should fail';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;

    BEGIN
        INSERT INTO feedback (
            workout_session_id, completion_status, difficulty, pain_reported,
            fatigue_after, satisfaction, terrain, external_conditions
        ) VALUES (
            invalid_session_id, 'complete', 'moderate', false,
            3, 3, 'downhill', 'wind'
        );
        RAISE EXCEPTION 'unsupported terrain should fail';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;

    BEGIN
        INSERT INTO feedback (
            workout_session_id, completion_status, difficulty, pain_reported,
            fatigue_after, satisfaction, terrain, external_conditions
        ) VALUES (
            invalid_session_id, 'complete', 'moderate', false,
            3, 3, 'rolling', 'storm'
        );
        RAISE EXCEPTION 'unsupported external condition should fail';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;
END;
$$;

ROLLBACK;
