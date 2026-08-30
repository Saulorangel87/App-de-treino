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
- `GET /v1/onboarding`: consulta limitações, objetivos e disponibilidade.
- `PUT /v1/onboarding/limitations`: salva informações de segurança.
- `PUT /v1/onboarding/goals`: salva até dois objetivos priorizados.
- `PUT /v1/onboarding/availability`: salva a disponibilidade semanal.

As sessões são opacas, armazenadas no PostgreSQL apenas como hash e enviadas ao navegador em cookie `HttpOnly`. Em produção, `APP_ENV=production` ativa também a exigência de HTTPS no cookie.

As telas locais ficam em `/entrar` e `/perfil`. O perfil possui quatro etapas e retoma dados já salvos. Configure `frontend/.env` a partir de `frontend/.env.example` quando a URL da API for diferente de `http://localhost:8080`.

## PWA

O frontend inclui manifesto, ícones, suporte à instalação e uma tela offline segura. O service worker armazena somente recursos estáticos; respostas da API e dados autenticados nunca entram no cache offline.

O registro do service worker ocorre apenas no build de produção. Para testar localmente, pare o servidor de desenvolvimento, execute `npm run build` e depois `npm run preview:pwa` dentro de `frontend/`. A prévia usa a porta 3000, já autorizada pela API local. A instalação exige HTTPS ou `localhost`/`127.0.0.1`.
