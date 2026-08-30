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
4. Execute a API com `pwsh -NoProfile -File scripts/run-api.ps1`.
5. Execute o frontend a partir de `frontend/` com `npm run dev`.

O frontend nunca se conecta diretamente ao PostgreSQL. Todo acesso passa pela API Go.

## Fluxo implementado

- `POST /v1/auth/register`: cria usuário, aplica hash seguro à senha e inicia sessão.
- `POST /v1/auth/login`: autentica e cria uma nova sessão.
- `POST /v1/auth/logout`: revoga a sessão atual.
- `GET /v1/me`: retorna o usuário autenticado.
- `GET /v1/profile`: consulta o perfil básico do ciclista.
- `PUT /v1/profile`: cria ou atualiza o perfil básico.

As sessões são opacas, armazenadas no PostgreSQL apenas como hash e enviadas ao navegador em cookie `HttpOnly`. Em produção, `APP_ENV=production` ativa também a exigência de HTTPS no cookie.

As telas locais ficam em `/entrar` e `/perfil`. Configure `frontend/.env` a partir de `frontend/.env.example` quando a URL da API for diferente de `http://localhost:8080`.
