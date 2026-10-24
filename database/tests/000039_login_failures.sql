-- Senhas erradas no login (000039): chave única, só no formato de resumo e
-- contagem nunca negativa.
BEGIN;

DO $$
BEGIN
    INSERT INTO login_failures (key, failures) VALUES ('ip:' || repeat('a', 64), 1);

    BEGIN
        INSERT INTO login_failures (key, failures) VALUES ('ip:' || repeat('a', 64), 2);
        RAISE EXCEPTION 'the same key was stored twice';
    EXCEPTION WHEN unique_violation THEN NULL;
    END;

    BEGIN
        INSERT INTO login_failures (key) VALUES ('ip:203.0.113.10');
        RAISE EXCEPTION 'a raw IP address was accepted as key';
    EXCEPTION WHEN check_violation THEN NULL;
    END;

    BEGIN
        INSERT INTO login_failures (key, failures) VALUES ('account:' || repeat('b', 64), -1);
        RAISE EXCEPTION 'a negative count was accepted';
    EXCEPTION WHEN check_violation THEN NULL;
    END;
END;
$$;

ROLLBACK;
