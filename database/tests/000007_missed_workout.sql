BEGIN;

DO $$
DECLARE
    test_user_id UUID;
    test_profile_id UUID;
    test_plan_id UUID;
    missed_workout_id UUID;
    adapted_workout_id UUID;
    future_workout_id UUID;
    completed_workout_id UUID;
    resulting_status TEXT;
    adapted_status TEXT;
    future_status TEXT;
    completed_status TEXT;
    session_count INTEGER;
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('missed-workout-test@example.invalid', 'test-only', 'Missed Workout Test')
    RETURNING id INTO test_user_id;

    INSERT INTO athlete_profiles (user_id, experience_level)
    VALUES (test_user_id, 'beginner')
    RETURNING id INTO test_profile_id;

    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status)
    VALUES (test_profile_id, CURRENT_DATE - 7, CURRENT_DATE + 7, 'active')
    RETURNING id INTO test_plan_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, CURRENT_DATE - 1, 'Treino perdido', 'Teste', 30,
        4, '{}'::jsonb, '{}'::jsonb, 'planned'
    ) RETURNING id INTO missed_workout_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, CURRENT_DATE - 2, 'Treino adaptado perdido', 'Teste', 30,
        4, '{}'::jsonb, '{}'::jsonb, 'adapted'
    ) RETURNING id INTO adapted_workout_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, CURRENT_DATE + 1, 'Treino futuro', 'Teste', 30,
        4, '{}'::jsonb, '{}'::jsonb, 'planned'
    ) RETURNING id INTO future_workout_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, CURRENT_DATE - 2, 'Treino encerrado', 'Teste', 30,
        4, '{}'::jsonb, '{}'::jsonb, 'completed'
    ) RETURNING id INTO completed_workout_id;

    UPDATE workouts
    SET status = 'skipped'
    WHERE id IN (missed_workout_id, adapted_workout_id)
      AND scheduled_on < CURRENT_DATE
      AND status IN ('planned', 'adapted');

    SELECT status INTO resulting_status
    FROM workouts WHERE id = missed_workout_id;
    SELECT status INTO adapted_status
    FROM workouts WHERE id = adapted_workout_id;
    IF adapted_status <> 'skipped' THEN
        RAISE EXCEPTION 'past adapted workout should be skipped, got %', adapted_status;
    END IF;
    SELECT status INTO future_status
    FROM workouts WHERE id = future_workout_id;
    SELECT status INTO completed_status
    FROM workouts WHERE id = completed_workout_id;
    SELECT COUNT(*) INTO session_count
    FROM workout_sessions WHERE workout_id = missed_workout_id;

    IF resulting_status <> 'skipped' THEN
        RAISE EXCEPTION 'past planned workout should be skipped, got %', resulting_status;
    END IF;
    IF future_status <> 'planned' THEN
        RAISE EXCEPTION 'future workout should remain planned, got %', future_status;
    END IF;
    IF completed_status <> 'completed' THEN
        RAISE EXCEPTION 'completed workout should remain completed, got %', completed_status;
    END IF;
    IF session_count <> 0 THEN
        RAISE EXCEPTION 'missed workout must not create a session, got %', session_count;
    END IF;
END;
$$;

ROLLBACK;
