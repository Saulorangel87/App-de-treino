BEGIN;

DO $$
DECLARE
    test_user_id UUID;
    test_profile_id UUID;
    stored_symptoms JSONB;
    stored_restriction BOOLEAN;
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('limitation-safety-test@example.invalid', 'test-only', 'Limitation Safety Test')
    RETURNING id INTO test_user_id;

    INSERT INTO athlete_profiles (user_id, experience_level)
    VALUES (test_user_id, 'intermediate')
    RETURNING id INTO test_profile_id;

    INSERT INTO injuries_or_limitations (
        athlete_profile_id, kind, description, symptoms_during_after,
        medical_restriction
    ) VALUES (
        test_profile_id, 'pain', 'Teste de segurança',
        '["dizziness", "extreme_fatigue"]'::jsonb, true
    );

    SELECT symptoms_during_after, medical_restriction
    INTO stored_symptoms, stored_restriction
    FROM injuries_or_limitations
    WHERE athlete_profile_id = test_profile_id;

    IF stored_symptoms <> '["dizziness", "extreme_fatigue"]'::jsonb OR NOT stored_restriction THEN
        RAISE EXCEPTION 'limitation safety context was not stored: symptoms %, restriction %',
            stored_symptoms, stored_restriction;
    END IF;

    BEGIN
        INSERT INTO injuries_or_limitations (
            athlete_profile_id, kind, description, symptoms_during_after
        ) VALUES (
            test_profile_id, 'pain', 'Sintomas demais',
            '["dizziness", "unusual_shortness_of_breath", "malaise", "extreme_fatigue", "other", "extra"]'::jsonb
        );
        RAISE EXCEPTION 'more than five safety symptoms should fail';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;
END;
$$;

ROLLBACK;
