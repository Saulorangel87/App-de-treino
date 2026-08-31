BEGIN;

DO $$
DECLARE
    test_user_id UUID;
    test_profile_id UUID;
    test_plan_id UUID;
    pending_workout_id UUID;
    resulting_status TEXT;
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('plan-completion-test@example.invalid', 'test-only', 'Plan Test')
    RETURNING id INTO test_user_id;

    INSERT INTO athlete_profiles (user_id, experience_level)
    VALUES (test_user_id, 'beginner')
    RETURNING id INTO test_profile_id;

    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status)
    VALUES (test_profile_id, DATE '2099-03-02', DATE '2099-03-29', 'active')
    RETURNING id INTO test_plan_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, DATE '2099-03-02', 'Concluído', 'Teste', 30,
        4, '{}'::jsonb, '{}'::jsonb, 'completed'
    );

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, DATE '2099-03-04', 'Último treino', 'Teste', 30,
        4, '{}'::jsonb, '{}'::jsonb, 'planned'
    ) RETURNING id INTO pending_workout_id;

    UPDATE workouts SET status = 'completed' WHERE id = pending_workout_id;

    SELECT status INTO resulting_status
    FROM training_plans WHERE id = test_plan_id;

    IF resulting_status <> 'completed' THEN
        RAISE EXCEPTION 'plan should be completed, got %', resulting_status;
    END IF;
END;
$$;

ROLLBACK;
