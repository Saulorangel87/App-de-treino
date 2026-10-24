-- Senhas erradas no login, contadas por IP e por conta, para o bloqueio valer
-- também depois de um reinício ou deploy da API (antes a contagem ficava só na
-- memória e zerava a cada deploy). A chave é um resumo SHA-256: nem o IP nem o
-- e-mail ficam gravados em claro. A própria API apaga linhas sem erro há mais
-- de um dia.
CREATE TABLE login_failures (
    key TEXT PRIMARY KEY CHECK (key ~ '^(ip|account):[0-9a-f]{64}$'),
    failures INTEGER NOT NULL DEFAULT 0 CHECK (failures >= 0),
    last_failure_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    blocked_until TIMESTAMPTZ
);

CREATE INDEX login_failures_last_failure_at_idx ON login_failures (last_failure_at);
