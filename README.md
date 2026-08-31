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
3. Aplique os arquivos `database/migrations/*.up.sql` ainda pendentes, em ordem numérica. O esquema atual chega à migração `000009`.
4. Execute a API com `pwsh -NoProfile -File scripts/run-api.ps1`.
5. Execute o frontend a partir de `frontend/` com `npm run dev`.

O frontend nunca se conecta diretamente ao PostgreSQL. Todo acesso passa pela API Go.
A configuração local deste projeto usa a porta `5433` no `.env`, pois a `5432` já estava ocupada no Windows. O `.env.example` mantém valores ilustrativos e não contém segredos.

## Fluxo implementado

- `POST /v1/auth/register`: cria usuário, aplica hash seguro à senha e inicia sessão.
- `POST /v1/auth/login`: autentica e cria uma nova sessão.
- `POST /v1/auth/logout`: revoga a sessão atual.
- `GET /v1/me`: retorna o usuário autenticado.
- `GET /v1/profile`: consulta o perfil básico do ciclista.
- `PUT /v1/profile`: cria ou atualiza o perfil básico.
- `GET /v1/onboarding`: consulta limitações, objetivos, disponibilidade e contexto opcional de ciclismo.
- `PUT /v1/onboarding/limitations`: salva informações de segurança.
- `PUT /v1/onboarding/goals`: salva até dois objetivos priorizados.
- `PUT /v1/onboarding/availability`: salva a disponibilidade semanal.
- `PUT /v1/onboarding/cycling-context`: salva histórico resumido, equipamento, terreno e meta opcional de prova.
- `GET /v1/assessments/current` e `POST /v1/assessments/submaximal`: consultam e registram o pedal de referência submáximo.
- `POST /v1/plans/generate`: gera e substitui o rascunho atual de quatro semanas.
- `GET /v1/plans/current`: consulta o plano ativo ou rascunho mais recente.
- `POST /v1/plans/{planID}/activate`: aprova um rascunho e mantém somente um plano ativo por atleta.
- `POST /v1/workouts/{workoutID}/start`: inicia uma sessão planejada do plano ativo.
- `POST /v1/workouts/{workoutID}/complete`: conclui a sessão e registra RPE, dificuldade, fadiga e dor.
- `POST /v1/workouts/{workoutID}/cancel`: cancela uma sessão em andamento e mantém esse histórico.
- `GET /v1/activities`: lista, para o atleta autenticado, as sessões concluídas e canceladas.

As sessões são opacas, armazenadas no PostgreSQL apenas como hash e enviadas ao navegador em cookie `HttpOnly`. Em produção, `APP_ENV=production` ativa também a exigência de HTTPS no cookie.

As rotas atuais do frontend são `/`, `/entrar`, `/perfil`, `/plano` e `/atividades`. A tela de atividades apresenta sessões concluídas e canceladas com data, duração, RPE e feedback. O perfil possui quatro etapas e retoma dados já salvos. Configure `frontend/.env` a partir de `frontend/.env.example` quando a URL da API for diferente de `http://localhost:8080`.

A tela `/plano` gera, apresenta e ativa ciclos de quatro semanas. O motor `rules-v1` é determinístico: considera experiência, objetivo, limitações, disponibilidade e o contexto opcional de ciclismo. Ele seleciona sessões específicas de forma gradual (cadência no indoor, subidas, sweet spot por potência/FTP e ritmo de prova), limita cada sessão ao tempo informado e reduz a intensidade quando há uma condição de segurança ativa. O dashboard usa o plano aprovado, explica a escala RPE e permite acompanhar a sessão do início ao feedback pós-treino.

O feedback de uma sessão concluída adapta de forma conservadora os próximos treinos planejados. Dor, fadiga, dificuldade e diferença entre RPE planejado e realizado podem reduzir duração ou esforço; uma resposta claramente fácil permite somente uma progressão pequena de duração. A decisão fica registrada no treino e é apresentada na interface. As regras completas estão em `docs/training-adaptation-rules.md`.

A rota `/avaliacao` permite registrar opcionalmente um pedal de referência submáximo, sem teste máximo ou diagnóstico. A referência serve para futuras progressões e não libera automaticamente treinos intensos.

Quando não existem mais sessões planejadas ou em andamento, o PostgreSQL marca o plano ativo como concluído. O usuário pode então gerar um novo ciclo sem apagar o histórico anterior. As regras de datas e estados estão em `docs/training-cycle-lifecycle.md`.

## Estado atual e próxima etapa

O fluxo completo de cadastro, onboarding, geração e ativação do plano, execução da sessão, feedback, adaptação e geração do próximo ciclo está implementado localmente. O frontend possui layout responsivo, PWA, rodapé com contatos, ajuda de RPE e bloqueio correto do scroll de fundo nos modais.

O histórico de atividades realizadas e canceladas está implementado localmente. O perfil também coleta, de forma opcional e condicional, contexto de ciclismo (volume, maior pedal, bicicleta, terreno, sensores, FTP e meta de prova). Esses dados já são persistidos; a próxima evolução do motor é usá-los para liberar prescrições específicas com regras conservadoras e explicáveis. Em paralelo, permanece a preparação da VPS Oracle com TLS, firewall, usuário PostgreSQL de privilégio mínimo, backups, restauração e Cloudflare Tunnel para expor somente o frontend e/ou a API — nunca o PostgreSQL.

## PWA

O frontend inclui manifesto, ícones, suporte à instalação e uma tela offline segura. O service worker armazena somente recursos estáticos; respostas da API e dados autenticados nunca entram no cache offline.

O registro do service worker ocorre apenas no build de produção. Para testar localmente, pare o servidor de desenvolvimento, execute `npm run build` e depois `npm run preview:pwa` dentro de `frontend/`. A prévia usa a porta 3000, já autorizada pela API local. A instalação exige HTTPS ou `localhost`/`127.0.0.1`.
