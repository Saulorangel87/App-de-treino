# Estado atual do projeto Cadência

Última atualização: 2 de setembro de 2026.

Este é o documento principal de continuidade. Ele registra o que está implementado, validado, publicado e pendente. Não incluir senhas, tokens, chaves de API ou conteúdo de arquivos `.env`.

## Resumo executivo

O MVP de ciclismo está em produção real e foi validado no navegador e em um celular. O fluxo de cadastro, confirmação de e-mail, perfil, geração e ativação de plano, execução, feedback, adaptação, histórico, evolução e logout está funcionando.

O motor atual é determinístico (`rules-v1`), baseado em regras explícitas e referências científicas. A integração com um modelo externo de IA ainda não foi implementada; a próxima fase poderá usar IA para explicações e interpretação dentro dos limites do motor.

## Repositório e produção

- Repositório: <https://github.com/Saulorangel87/App-de-treino>
- Branch de produção: `master`.
- Frontend: <https://cadencia.devsaulo.com.br>
- API: <https://cadencia-api.devsaulo.com.br>
- VPS: Oracle Cloud, Ubuntu, acesso administrativo por SSH na porta 22.
- Código na VPS: `/home/ubuntu/apps/cadencia`.
- Commit implantado: `41638da fix: atualiza dependencias Go vulneraveis`.
- O Cloudflare Tunnel dedicado expõe somente frontend e API; o PostgreSQL não possui hostname, rota pública ou porta publicada.

## Arquitetura efetiva

```text
Navegador / PWA
       |
       | HTTPS pelo Cloudflare Tunnel
       v
frontend (cadencia_edge)
       |
       | REST / cookies HttpOnly
       v
api (cadencia_edge + cadencia_data)
       |
       | rede Docker interna
       v
PostgreSQL (cadencia_data, sem porta no host)
```

- Desenvolvimento usa PostgreSQL 17 em Docker, normalmente em `127.0.0.1:5433`.
- Produção usa PostgreSQL 17 na VPS Oracle, em rede Docker interna.
- O navegador nunca recebe credenciais nem acessa o PostgreSQL.
- A composição de produção não publica portas do Cadência no host.
- Configurações e segredos permanecem fora do Git, no `.env.production` da VPS.

## Organização

- `frontend/`: React/TypeScript com Vinext, PWA e interface responsiva.
- `backend/`: API REST em Go.
- `database/migrations/`: migrações PostgreSQL até `000012`.
- `database/tests/`: verificações SQL.
- `api/openapi.yaml`: contrato da API local e de produção.
- `infrastructure/cadencia/`: composição Docker, Dockerfile, migrações, backup e unidades systemd de produção.
- `docs/`: regras de produto, ciclo de vida, arquitetura e operação.

## Funcionalidades implementadas

### Conta e segurança

- Cadastro, login, logout e sessão por cookie `HttpOnly`.
- Senhas com bcrypt e tokens de sessão armazenados somente como hash.
- Confirmação de e-mail por token aleatório, expirável e de uso único.
- Recuperação de senha por token expirável, com revogação das sessões após a troca.
- Produção configurada com Resend e remetente verificado.
- Geração e ativação de planos bloqueadas enquanto o e-mail não estiver confirmado.

### Perfil e planejamento

- Perfil em quatro etapas: dados básicos, limitações, objetivos e disponibilidade.
- Até dois objetivos priorizados.
- Disponibilidade individual dos sete dias, com opções de duração até 8 horas.
- Contexto opcional de ciclismo: horas, pedais e distância semanal recente, semanas de regularidade, maior distância e pedal, preferências de sessão, equipamento, terreno, sensores, FTP e meta de prova.
- Resumo observado dos últimos 28 dias: sessões concluídas, minutos realizados, RPE e fadiga médios, dor relatada e check-ins de recuperação.
- Motor `rules-v1` com ciclos de quatro semanas, progressão, recuperação e datas calculadas para a semana corrente.
- Geração de rascunho, revisão, ativação e geração do próximo ciclo sem apagar o histórico.

### Treinos, feedback e evolução

- Sessões planejadas, iniciadas, concluídas e canceladas.
- Feedback com RPE, dificuldade, fadiga, dor e observações.
- Métricas opcionais: distância, elevação, frequência cardíaca média e potência média.
- Adaptação conservadora após feedback e check-in diário de sono, estresse e fadiga.
- Avaliação inicial submáxima, sem teste máximo ou diagnóstico.
- O histórico observado agora participa da geração: sinais recentes de dor, fadiga ou recuperação insuficiente protegem as sessões futuras de forma conservadora e ficam no snapshot do plano.
- Sessões específicas para perfis adequados: cadência, subidas, sweet spot, ritmo de prova e intervalos controlados.
- Histórico em `/atividades` e agregações observadas em `/evolucao`.
- Indicadores de consistência, carga semanal, prontidão e explicabilidade.

### Interface e PWA

- Dashboard, plano, atividades, avaliação, recuperação e evolução.
- Modal de sessão no mobile, check visual de treinos concluídos e logout.
- PWA instalável, manifesto, ícones e tela offline segura.
- Cache offline limitado a recursos estáticos; dados autenticados não entram no cache.
- Interface em português do Brasil, responsiva e sem rolagem horizontal no mobile.

## API disponível

As rotas estão descritas em `api/openapi.yaml`. Os grupos principais são:

- Saúde: `GET /health`, `GET /ready`.
- Conta: `/v1/auth/*` e `GET /v1/me`.
- Perfil/onboarding: `/v1/profile` e `/v1/onboarding/*`.
- Avaliação e recuperação: `/v1/assessments/*` e `/v1/recovery/today`.
- Planejamento: `/v1/plans/*`.
- Sessões: `/v1/workouts/{workoutID}/*`.
- Histórico: `GET /v1/activities`.
- Evolução: `GET /v1/evolution/summary`.

## Banco e migrações

- Migrações aplicadas no projeto: `000001` a `000012`.
- `000012` adiciona confirmação de e-mail e recuperação de senha.
- Produção possui registro de migrações em `cadencia_schema_migrations`.
- O usuário da API não é superusuário; o proprietário do banco é reservado para operações administrativas.
- Não há alteração de esquema pendente neste momento.

## Produção validada

Em 2 de setembro de 2026:

- Código `41638da` atualizado na VPS por fast-forward.
- Imagem da API reconstruída com Go 1.25.
- Somente o container `cadencia-api-1` foi recriado.
- `cadencia-api-1`, `cadencia-frontend-1` e `cadencia-postgres-1` ficaram saudáveis; o túnel permaneceu ativo.
- API interna: `/health` retornou `{"service":"cadencia-api","status":"ok"}`.
- API interna: `/ready` retornou `{"status":"ready"}`.
- API pública e frontend público retornaram HTTP 200.
- Nenhuma porta do Cadência foi publicada no host.
- Cadastro, confirmação de e-mail, ativação de plano e recuperação de senha foram testados em produção; uma mensagem de confirmação caiu em spam, sem falha funcional.

## Dependências e segurança

- GitHub Dependabot: 0 alertas abertos e 39 fechados após a atualização do commit `41638da`.
- `pgx` atualizado para `5.9.2`.
- `golang.org/x/crypto` atualizado para `0.55.0`.
- `golang.org/x/text` atualizado para `0.41.0`.
- Toolchain de build da API atualizado para Go `1.25`.
- O backend foi validado com `go test ./...`, build Docker e `govulncheck`.
- O `govulncheck` não encontrou vulnerabilidades alcançáveis pelo código; permanece uma advisory de `openpgp` não utilizado e sem correção upstream.

## Backups e operação

- Backup diário do Cadência ativo em `cadencia-backup.timer`, às 03:30 UTC.
- Retenção configurada: 14 dias.
- Dumps em formato customizado, validados por `pg_restore --list`.
- Diretório de produção: `/var/backups/cadencia`.
- Backup preventivo do deploy de 2 de setembro: `cadencia-20260902T104801Z.dump`.
- O teste de restauração completo foi concluído em 2 de setembro de 2026 com o dump `cadencia-20260902T104801Z.dump`: um PostgreSQL 17 temporário restaurou o arquivo com sucesso, apresentou 15 tabelas públicas e `cadencia_schema_migrations`, e foi removido ao final. O banco e os volumes de produção não foram alterados.

## Auditoria da VPS e pendências operacionais

O Cadência está isolado, mas a VPS hospeda outros aplicativos. Foram observadas portas Docker publicadas para serviços como Rotas, Despesas, Estoque, n8n, Immich, Uptime Kuma, Tecboard e Nginx Proxy Manager.

O teste externo realizado a partir do ambiente atual confirmou acesso à porta 22 e à porta 2283 (Immich); as demais portas testadas não responderam externamente. A política local de entrada ainda está permissiva (`accept`), não há `ufw` instalado e a cadeia `DOCKER-USER` está vazia.

Na auditoria somente leitura de 2 de setembro de 2026, o host também apresentou Home Assistant na porta 8123, além das portas já catalogadas. O processo nessa porta é o Python do Home Assistant (`/config`); a porta 8888 pertence ao `casaos-gateway` e a 2283 é a publicação Docker do Immich. O Nginx Proxy Manager possui os destinos `casaos.oraclecloud.com.br` → `100.67.151.30:8888` e `immich.photo.com.br` → `100.67.151.30:2283`, ambos por Tailscale. Existe um segundo container Cloudflare Tunnel, separado do túnel dedicado do Cadência, que encaminha o Immich; seus logs recentes registraram reconexão bem-sucedida e aviso de versão desatualizada. Os testes externos confirmaram inicialmente apenas 22 e 2283 acessíveis. Em seguida, foi aplicada uma regra persistente na cadeia `DOCKER-USER` para bloquear `tcp/2283` somente pela interface pública `enp0s6`; as regras TCP públicas `22`, `81`, `2283`, `8096` e `8097` também foram removidas da Oracle Cloud, restando somente ICMP. Immich continua respondendo via Tailscale e loopback, e Cadência foi validado após as alterações.

Pendências, sem executar bloqueios automáticos:

1. Completar o mapa de cada domínio, túnel e proxy dos aplicativos existentes.
2. Monitorar as regras de rede da Oracle Cloud; as regras TCP públicas desnecessárias já foram removidas.
3. Verificar periodicamente o acesso administrativo pelo Tailscale; novas conexões SSH pela interface pública já estão bloqueadas.
4. Fechar outras portas diretas desnecessárias, especialmente serviços que já usam proxy; `2283` do Immich já está bloqueada na interface pública.
5. Definir cópia externa dos backups e monitoramento de falhas.
6. Atualizar esta documentação após cada mudança de infraestrutura.

## Próximas etapas do produto

1. Concluir o hardening da VPS.
2. Validar visualmente e publicar o ajuste aplicado na tela inicial desktop: texto de privacidade mais amigável e menos altura/scroll.
3. Integrar IA no backend para explicações, interpretação de feedback e comunicação personalizada, sempre subordinada ao motor de regras e às proteções de segurança. A primeira camada de sessões estruturadas já foi implementada: o motor envia etapas com duração, RPE e instruções acionáveis.
4. Expandir a coleta progressiva de dados do ciclista. O motor já incorpora o resumo observado de sessões e recuperação sem transformar relatos em metas rígidas; a próxima ampliação deve adicionar métricas de desempenho com validação específica.
5. Evoluir sessões específicas de cadência, tiros, subidas, potência e preparação para provas com regras próprias e validação científica. As preferências agora orientam a sessão de qualidade quando há contexto compatível; a ampliação do catálogo de evidências específicas de ciclismo fica registrada como etapa futura, sem migração aplicada. Ainda falta ampliar a cobertura e revisar parâmetros com profissional habilitado.
6. Avaliar integrações externas, como Strava, somente depois de definir escopo, consentimento, custos e segurança dos tokens.

## Como iniciar localmente

1. Inicie o Docker Desktop.
2. Na raiz, execute `docker compose up -d postgres`.
3. Inicie a API com `pwsh -NoProfile -File scripts/run-api.ps1`.
4. Em outro terminal, entre em `frontend/` e execute `npm run dev`.
5. Acesse `http://localhost:3000`.
6. API local: `http://localhost:8080/health` e `http://localhost:8080/ready`.

Para testar o PWA localmente, pare o servidor de desenvolvimento e execute `npm run build` e `npm run preview:pwa` dentro de `frontend/`.

## Como retomar

Antes de alterar o projeto:

1. Leia este arquivo, `README.md`, `docs/architecture-decisions.md`, `docs/training-cycle-lifecycle.md` e `docs/training-adaptation-rules.md`.
2. Confira `git status` e os commits recentes.
3. Preserve os bancos PostgreSQL local e da VPS.
4. Não publique, faça commit ou altere infraestrutura sem autorização explícita.
