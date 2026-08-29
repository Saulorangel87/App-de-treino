# Cadência

Aplicação de planejamento adaptativo de treinos de ciclismo.

## Estrutura

- `frontend/`: aplicação React/TypeScript publicada com Sites.
- `backend/`: API REST em Go.
- `database/`: migrações PostgreSQL versionadas.
- `api/`: contrato OpenAPI.
- `infrastructure/`: configuração futura da VPS Oracle.
- `docs/`: decisões e regras do produto.

## Ambiente local

1. Copie `.env.example` para `.env` e use somente credenciais locais.
2. Inicie o PostgreSQL com `docker compose up -d postgres`.
3. Aplique `database/migrations/000001_initial_schema.up.sql` no banco local.
4. Execute a API a partir de `backend/` com as variáveis do `.env` carregadas.
5. Execute o frontend a partir de `frontend/` com `npm run dev`.

O frontend nunca se conecta diretamente ao PostgreSQL. Todo acesso passa pela API Go.
