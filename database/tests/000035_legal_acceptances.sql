-- Aceite dos termos (000035): único por usuário e versão, formato de versão
-- validado e apagado em cascata com a conta.
BEGIN;

DO $$
DECLARE
    test_user_id UUID;
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('legal-acceptance-test@example.invalid', 'test-only', 'Legal Acceptance Test')
    RETURNING id INTO test_user_id;

    INSERT INTO legal_acceptances (user_id, terms_version) VALUES (test_user_id, '2026-09-30');
    INSERT INTO legal_acceptances (user_id, terms_version) VALUES (test_user_id, '2026-12-01');

    BEGIN
        INSERT INTO legal_acceptances (user_id, terms_version) VALUES (test_user_id, '2026-09-30');
        RAISE EXCEPTION 'the same version was accepted twice';
    EXCEPTION WHEN unique_violation THEN NULL;
    END;

    BEGIN
        INSERT INTO legal_acceptances (user_id, terms_version) VALUES (test_user_id, 'ontem');
        RAISE EXCEPTION 'a malformed version was accepted';
    EXCEPTION WHEN check_violation THEN NULL;
    END;

    DELETE FROM users WHERE id = test_user_id;
    IF EXISTS (SELECT 1 FROM legal_acceptances WHERE user_id = test_user_id) THEN
        RAISE EXCEPTION 'deleting the account must delete its acceptances';
    END IF;
END;
$$;

ROLLBACK;
