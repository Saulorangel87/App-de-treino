BEGIN;

DO $$
DECLARE
    test_user_id UUID;
    test_profile_id UUID;
    test_plan_id UUID;
    test_workout_id UUID;
    test_session_id UUID;
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('account-deletion-test@example.invalid', 'test-only', 'Account Deletion Test')
    RETURNING id INTO test_user_id;

    INSERT INTO athlete_profiles (user_id, experience_level)
    VALUES (test_user_id, 'intermediate')
    RETURNING id INTO test_profile_id;

    INSERT INTO goals (athlete_profile_id, goal_type, priority)
    VALUES (test_profile_id, 'health', 1);

    INSERT INTO availability (athlete_profile_id, weekday, available_minutes)
    VALUES (test_profile_id, 1, 60);

    INSERT INTO recovery_data (athlete_profile_id, recorded_on, sleep_quality, stress_level, fatigue_level)
    VALUES (test_profile_id, DATE '2099-01-01', 4, 2, 2);

    INSERT INTO injuries_or_limitations (athlete_profile_id, kind, description)
    VALUES (test_profile_id, 'pain', 'Teste de exclusão');

    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status)
    VALUES (test_profile_id, DATE '2099-01-01', DATE '2099-01-28', 'active')
    RETURNING id INTO test_plan_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, status
    ) VALUES (
        test_plan_id, DATE '2099-01-01', 'Teste de exclusão', 'Teste', 30,
        4, '{}'::jsonb, 'completed'
    ) RETURNING id INTO test_workout_id;

    INSERT INTO workout_sessions (
        workout_id, athlete_profile_id, started_at, completed_at,
        duration_minutes, actual_rpe, status
    ) VALUES (
        test_workout_id, test_profile_id, now() - interval '30 minutes', now(),
        30, 4, 'completed'
    ) RETURNING id INTO test_session_id;

    INSERT INTO feedback (workout_session_id, difficulty, pain_reported, fatigue_after)
    VALUES (test_session_id, 'moderate', false, 2);

    INSERT INTO cycling_assessments (athlete_profile_id, duration_minutes, actual_rpe)
    VALUES (test_profile_id, 20, 5);

    INSERT INTO auth_sessions (user_id, token_hash, expires_at)
    VALUES (test_user_id, decode(repeat('ab', 32), 'hex'), now() + interval '1 hour');

    INSERT INTO auth_email_tokens (user_id, purpose, token_hash, expires_at)
    VALUES (test_user_id, 'verify_email', decode(repeat('cd', 32), 'hex'), now() + interval '1 hour');

    INSERT INTO user_feedback (user_id, category, rating, message)
    VALUES (test_user_id, 'suggestion', 5, 'Teste de exclusão da conta');

    DELETE FROM users WHERE id = test_user_id;

    IF EXISTS (SELECT 1 FROM athlete_profiles WHERE user_id = test_user_id)
       OR EXISTS (SELECT 1 FROM auth_sessions WHERE user_id = test_user_id)
       OR EXISTS (SELECT 1 FROM auth_email_tokens WHERE user_id = test_user_id)
       OR EXISTS (SELECT 1 FROM user_feedback WHERE user_id = test_user_id)
       OR EXISTS (SELECT 1 FROM goals WHERE athlete_profile_id = test_profile_id)
       OR EXISTS (SELECT 1 FROM availability WHERE athlete_profile_id = test_profile_id)
       OR EXISTS (SELECT 1 FROM recovery_data WHERE athlete_profile_id = test_profile_id)
       OR EXISTS (SELECT 1 FROM injuries_or_limitations WHERE athlete_profile_id = test_profile_id)
       OR EXISTS (SELECT 1 FROM training_plans WHERE athlete_profile_id = test_profile_id)
       OR EXISTS (SELECT 1 FROM workouts WHERE training_plan_id = test_plan_id)
       OR EXISTS (SELECT 1 FROM workout_sessions WHERE id = test_session_id)
       OR EXISTS (SELECT 1 FROM feedback WHERE workout_session_id = test_session_id)
       OR EXISTS (SELECT 1 FROM cycling_assessments WHERE athlete_profile_id = test_profile_id)
    THEN
        RAISE EXCEPTION 'account deletion did not cascade to every user-owned table';
    END IF;
END;
$$;

ROLLBACK;
