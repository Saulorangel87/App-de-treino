BEGIN;

DO $$
DECLARE
    test_user_id UUID;
    test_profile_id UUID;
    stored_trend TEXT;
    stored_surgery BOOLEAN;
    stored_prohibited BOOLEAN;
    stored_condition BOOLEAN;
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('profile-safety-context-test@example.invalid', 'test-only', 'Profile Safety Context Test')
    RETURNING id INTO test_user_id;

    INSERT INTO athlete_profiles (
        user_id, experience_level, waist_cm, body_fat_percent, weight_trend
    ) VALUES (
        test_user_id, 'intermediate', 82.5, 17.2, 'stable'
    ) RETURNING id, weight_trend INTO test_profile_id, stored_trend;

    IF stored_trend <> 'stable' THEN
        RAISE EXCEPTION 'weight trend was not stored: %', stored_trend;
    END IF;

    INSERT INTO injuries_or_limitations (
        athlete_profile_id, kind, description, recent_surgery,
        exercise_prohibited, condition_affecting_exercise
    ) VALUES (
        test_profile_id, 'medical_condition', 'Contexto seguro de teste', true,
        true, true
    ) RETURNING recent_surgery, exercise_prohibited, condition_affecting_exercise
        INTO stored_surgery, stored_prohibited, stored_condition;

    IF NOT stored_surgery OR NOT stored_prohibited OR NOT stored_condition THEN
        RAISE EXCEPTION 'safety flags were not stored';
    END IF;

    BEGIN
        UPDATE athlete_profiles SET body_fat_percent = 71 WHERE id = test_profile_id;
        RAISE EXCEPTION 'out-of-range body fat should fail';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;
END;
$$;

ROLLBACK;
