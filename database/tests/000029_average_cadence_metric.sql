BEGIN;

DO $$
DECLARE
    test_user_id UUID;
    test_profile_id UUID;
    test_plan_id UUID;
    test_workout_id UUID;
    test_session_id UUID;
    stored_cadence INTEGER;
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('average-cadence-test@example.invalid', 'test-only', 'Average Cadence Test')
    RETURNING id INTO test_user_id;

    INSERT INTO athlete_profiles (user_id, experience_level)
    VALUES (test_user_id, 'intermediate')
    RETURNING id INTO test_profile_id;

    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status)
    VALUES (test_profile_id, DATE '2099-08-01', DATE '2099-08-28', 'active')
    RETURNING id INTO test_plan_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, DATE '2099-08-01', 'Cadência média', 'Teste', 45,
        5, '{}'::jsonb, '{"rules":[]}'::jsonb, 'completed'
    ) RETURNING id INTO test_workout_id;

    INSERT INTO workout_sessions (
        workout_id, athlete_profile_id, started_at, completed_at,
        duration_minutes, actual_rpe, average_cadence_rpm, status
    ) VALUES (
        test_workout_id, test_profile_id, now() - interval '45 minutes', now(),
        45, 5, 88, 'completed'
    ) RETURNING id INTO test_session_id;

    SELECT average_cadence_rpm INTO stored_cadence
    FROM workout_sessions
    WHERE id = test_session_id;

    IF stored_cadence <> 88 THEN
        RAISE EXCEPTION 'average cadence was not stored: %', stored_cadence;
    END IF;

    BEGIN
        UPDATE workout_sessions
        SET average_cadence_rpm = 301
        WHERE id = test_session_id;
        RAISE EXCEPTION 'out-of-range average cadence should fail';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;
END;
$$;

ROLLBACK;
