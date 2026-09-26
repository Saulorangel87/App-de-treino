# Estado atual do Cadência

Última atualização: 26 de setembro de 2026.

Este é o documento de continuidade: curto e sempre atual. O diário cronológico das fatias de trabalho (com datas, validações e decisões) está em [`changelog/project-status-history.md`](changelog/project-status-history.md). Não inclua senhas, tokens, chaves de API nem conteúdo de `.env`.

## Produção (VPS Oracle)

| Item | Valor |
| --- | --- |
| Frontend | <https://cadencia.devsaulo.com.br> |
| API | <https://cadencia-api.devsaulo.com.br> |
| Versão publicada | `0.33.0` (merge `f925753` + correção `bf87ad2`, deploy de 26/09/2026) |
| Migrações aplicadas | `000001` a `000030` |
| Motor prescritivo | `rules-v1` (único); `rules-v2` e demais shadows são somente observacionais |
| Último backup preventivo | `cadencia-20260915T113044Z.dump` |
| Validação pós-deploy | login, `GET /v1/plans/current` e `logout-others` = 200 com a conta de smoke test; `/ready` verifica o schema |

Escopo: somente ciclismo (estrada, MTB XCO/XCM, gravel e indoor). Corrida e musculação são produtos separados.

## Endurecimento operacional — publicado na `0.33.0`

Conjunto de melhorias de robustez publicado em 26/09/2026 (PR #5 e #18), a partir do diagnóstico do incidente de 14/09 (schema defasado atrás de healthchecks verdes) e da revisão geral do projeto. A versão local do frontend passou para `0.33.0`.

**Schema e deploy**
- A API confere na inicialização se todas as migrações de `database.RequiredMigrations` estão em `cadencia_schema_migrations`. Em produção, a divergência é fatal; fora dela, apenas um aviso. `GET /ready` retorna `503 schema_behind` enquanto faltar migração (em produção). `TestRequiredMigrationsMatchFiles` impede esquecer de atualizar a lista ao criar uma migração.
- O `HEALTHCHECK` do container da API usa `/ready`; o tunnel só sobe com a API pronta.
- `migrate.sh`: lock consultivo por transação, checksum das migrações aplicadas (coluna aditiva `checksum`), `DRY_RUN=1` e `MIGRATE_STRICT=1`.
- `infrastructure/cadencia/scripts/deploy.sh` automatiza fast-forward, backup, build, migrações, recriação, espera por `/ready` e smoke test (público e autenticado).

**CI**
- `.github/workflows/ci.yml`: `gofmt`, `go vet`, `go test -race`, `govulncheck`; migrações + fixtures SQL em PostgreSQL 17; `tsc`, `oxlint`, `vitest`, build e `npm audit` do frontend. Dependabot para gomod, npm, docker e actions.
- O job `database` roda as fixtures SQL pela primeira vez em CI: se falhar, o motivo provável é uma fixture que dependia de dados do banco de desenvolvimento, não da migração.

**Autenticação**
- Login gasta o mesmo custo de bcrypt para e-mails inexistentes; falhas de login são limitadas também por conta (10 em 15 min), além do limite por IP.
- Novos endpoints `POST /v1/auth/change-password` e `POST /v1/auth/logout-others`, com cartão “Segurança do acesso” em `/configuracoes`.
- Sessões e tokens de e-mail expirados são removidos por uma rotina horária na API.
- Os links de desenvolvimento (`development_*_url`) só existem fora de produção **e** com `APP_BASE_URL` em loopback (falha fechada). `APP_ENV` só aceita `development`, `test` ou `production`; em produção `APP_BASE_URL` e `ALLOWED_ORIGIN` precisam ser `https`.

**Observabilidade e infraestrutura**
- Middleware com `X-Request-ID`, log de acesso estruturado (sondas saudáveis não são registradas) e recuperação de panic com JSON 500.
- Compose: imagens fixadas por digest, `mem_limit` (`API_MEM_LIMIT`, `FRONTEND_MEM_LIMIT`, `POSTGRES_MEM_LIMIT`) e rotação de logs.
- Backup: alerta de falha/sucesso por URL de ping, cópia externa via `rclone` e teste mensal de restauração (`test-restore.sh` + timer). Ambos opcionais e desligados até serem configurados na VPS.
- Pool do PostgreSQL configurável (`DB_MAX_CONNS`, `DB_MIN_CONNS`); `WriteTimeout` acompanha o timeout da IA.

**Frontend**
- `apiRequest` com timeout, novas tentativas para GET e redirecionamento ao login quando a sessão expira em qualquer rota protegida.
- CSP e demais cabeçalhos de segurança em `next.config.ts` (verificados com `vinext start`; sem violações no Chrome).
- Service worker não falha mais a instalação quando `/offline` não existe (o precache falhava com `vinext start`).
- Removidos 56 componentes `ui/*` e 8 dependências sem uso; `perfil/page.tsx` (1.946 linhas) dividido em modelo, hook e quatro etapas; lint sem erros; `npm test` (vitest) e e2e Playwright (`npm run e2e`, exige API e banco locais).

**Validação executada neste checkout:** `go vet`, `go test ./...`, `tsc --noEmit`, `oxlint`, `vitest`, `npm run build`, validação do Compose de produção e cabeçalhos no `vinext start`. Com Docker, em um PostgreSQL 17 descartável (o banco de desenvolvimento não foi tocado):
- `migrate.sh`: 30 migrações aplicadas, reexecução idempotente, `DRY_RUN`, aviso de checksum e falha com `MIGRATE_STRICT=1`;
- as 14 fixtures SQL passaram e as migrações `down` reverteram em ordem inversa;
- imagem da API construída (48 MB): em produção, com schema completo, `/ready` = 200 e container `healthy`; removendo o registro da `000030`, `/ready` = 503 e uma nova inicialização falha com o nome da migração ausente (o cenário do incidente de 14/09); o log de acesso e o `X-Request-ID` funcionaram;
- imagem do frontend construída: `healthy`, com CSP incluindo a origem da API;
- e2e Playwright (2 testes) contra a API real: cadastro, confirmação, senha atual errada recusada, troca de senha, “sair dos outros dispositivos”, login antigo recusado e novo aceito, e redirecionamento ao login com sessão expirada;
- `test-restore.sh` restaurou um dump real (30 migrações verificadas).

**Ainda não executado:** `backup-postgres.sh` e `deploy.sh` de ponta a ponta (dependem do Compose de produção e da VPS; só `sh -n`), o job de CI no GitHub (`-race` e `govulncheck`) e o build multi-arquitetura `linux/arm64` da VPS.

**Achado do e2e:** clicar em enviar antes da hidratação fazia o navegador enviar o formulário por GET, com a senha na URL. Os formulários com senha agora usam `method="post"`.

## Deploy de 26/09/2026

- Primeiro uso do `deploy.sh`. O backup preventivo foi `cadencia-20260926T233525Z.dump`; o container do PostgreSQL foi recriado pela mudança de digest da imagem (volume intacto).
- A primeira tentativa parou no `DRY_RUN` do `migrate.sh` (a tabela de produção não tinha a coluna `checksum`); corrigido no PR #18. A coluna foi adicionada nesta publicação; os registros antigos permanecem sem checksum, por desenho.
- O smoke test público falhou com 530 porque o tunnel ainda reconectava; o `deploy.sh` agora espera até ~90 s.
- Conta de smoke test criada em produção (`smoke-test@cadencia.devsaulo.com.br`, e-mail confirmado no banco); credenciais em `/etc/cadencia/smoke.env` na VPS, lidas pelo `deploy.sh`.
- Dependabot passou a abrir PRs (#6 a #13). Os saltos de versão maior (Node 26, Go 1.27, actions) exigem teste próprio antes de mesclar.

## Pendências operacionais (fora do código)

- Configurar na VPS (opcionais, ainda desligados): cópia externa dos backups (`CADENCIA_OFFSITE_REMOTE`), monitor de ping (`CADENCIA_HEALTHCHECK_URL`) e o timer `cadencia-restore-test`.
- Hardening da VPS e limpeza gradual do que restar de dívida técnica.
- Coleta longitudinal de dados reais antes de dar autoridade adicional aos shadows.
- Decisão de produto: a hospedagem do frontend usa `vinext` (beta) com dependências herdadas do ambiente de criação (`wrangler`, `@openai/sites-vite-plugin`); avaliar migração para uma base mais estável.
