# Decisões de arquitetura

Última revisão: 3 de setembro de 2026.

## ADR-001 — Banco de dados próprio

**Status:** Aceita e aplicada.

O Cadência usa PostgreSQL controlado pelo proprietário do projeto:

- Desenvolvimento e testes: PostgreSQL 17 em Docker Compose, exposto somente em loopback (`127.0.0.1:5433`).
- Produção: PostgreSQL 17 na VPS Oracle, dentro da rede Docker `cadencia_data`.
- O backend Go é o único componente autorizado a acessar o banco.
- O navegador nunca recebe credenciais, hostname ou porta do PostgreSQL.
- O usuário da API não possui privilégios administrativos.
- Migrações SQL são versionadas e registradas em `cadencia_schema_migrations`.

## ADR-002 — Frontend, backend e transporte

**Status:** Aceita e aplicada.

```text
Navegador / PWA
      |
      | HTTPS via Cloudflare Tunnel dedicado
      v
Frontend (cadencia_edge)
      |
      | REST e cookie HttpOnly
      v
API Go (cadencia_edge + cadencia_data)
      |
      | rede interna Docker
      v
PostgreSQL (cadencia_data)
```

- Frontend: React/TypeScript com Vinext.
- Backend: Go, API REST.
- O Cloudflare Tunnel encaminha somente frontend e API.
- Produção usa `https://cadencia.devsaulo.com.br` e `https://cadencia-api.devsaulo.com.br`.
- Nenhum serviço da composição de produção publica portas do Cadência no host.

## ADR-003 — Motor de treinamento e IA

**Status:** Motor aplicado; camada explicativa local e fallback remoto implementados, Worker remoto selecionado em produção para preservar a capacidade da VPS.

O planejamento é gerado pelo motor determinístico `rules-v1`, com regras explícitas, limitações de segurança, disponibilidade e evidências científicas. As prescrições também carregam etapas operacionais estruturadas para que cada sessão seja executável e explicável. Uma futura integração de IA ficará no backend e poderá explicar decisões, interpretar feedback e adaptar a comunicação, mas não poderá inventar estudos, ultrapassar as regras ou diagnosticar condições clínicas.

Os formatos das sessões são mantidos em uma biblioteca de protocolos com chaves estáveis e referências associadas. A biblioteca define a forma do estímulo; o motor ainda aplica nível, disponibilidade, progressão, recuperação e limitações antes de gerar cada duração final. A IA explicativa recebe somente fatos já validados do treino, suas regras e o escopo da evidência; ela não pode alterar a prescrição.

O contexto de ciclismo permanece em JSONB para evoluir sem migrações a cada pergunta opcional. Atualmente inclui horas semanais, pedais por semana, distância semanal recente, semanas de regularidade, maior distância e pedal, preferências de sessão, equipamento, terreno e sensores. Além dele, o motor consulta um resumo agregado dos últimos 28 dias de sessões concluídas e check-ins de recuperação. Esses dados são preservados no snapshot do plano; sinais de dor, fadiga elevada ou recuperação insuficiente apenas protegem a sessão de forma conservadora, sem criar metas rígidas ou diagnósticos. Novas fórmulas de carga só serão ativadas após revisão e testes específicos.

A primeira implementação de IA usa Ollama local como provedor opcional. `AI_ENABLED=false` é o padrão; quando habilitado, o backend aplica timeout de até 60 segundos, saída limitada a 512 tokens e no máximo duas chamadas simultâneas (padrão: uma). A API local do Ollama não é exposta ao navegador ou à internet. Uma rota separada e protegida do Worker Cloudflare (`/cadencia/explanation`) foi publicada para fallback, com autenticação por segredo, allowlist de campos, limite de corpo, timeout, limite de requisições por janela e resposta sanitizada. O Worker usa `openai/gpt-oss-20b` na Groq, com o segredo mantido somente no Worker e no `.env.production` da VPS; uma chamada sintética autenticada foi validada. O contrato legado do Worker permanece inalterado para não interromper outros projetos. Se os provedores falharem, a API retorna o resumo determinístico do motor.

## ADR-004 — Autenticação e e-mail

**Status:** Aceita e aplicada.

- Senhas com bcrypt.
- Sessões opacas com hash SHA-256 armazenado no banco.
- Cookies `HttpOnly`; em produção também `Secure`.
- Confirmação de e-mail e recuperação de senha com tokens aleatórios, expiráveis, de uso único e armazenados somente como hash.
- E-mails de produção enviados pelo Resend, com remetente verificado.
- Troca de senha revoga todas as sessões existentes.

## ADR-005 — Migrações e produção

**Status:** Aplicada.

As migrações `000001` a `000012` foram aplicadas localmente e em produção. A migração `000013` cria os relatos de feedback de produto e a `000014` adiciona o controle de envio do resumo semanal; ambas estão versionadas e aguardam o próximo deploy de produção. A proposta de catálogo ampliado de evidências específicas de ciclismo fica registrada como etapa futura e ainda não foi aplicada. Antes de uma mudança estrutural, deve existir backup verificável e a migração deve ser executada pelo perfil `maintenance`.

## ADR-006 — Feedback de produto

**Status:** Implementada no código; migrações `000013` e `000014` pendentes de aplicação no próximo deploy.

O feedback solicitado aos primeiros ciclistas é separado do feedback pós-treino. A rota autenticada `POST /v1/feedback` aceita somente uma categoria (`experience`, `bug` ou `suggestion`), uma nota de 1 a 5 e uma mensagem entre 10 e 2000 caracteres. O registro é vinculado ao usuário na tabela `user_feedback`, sem armazenar um e-mail duplicado ou permitir conteúdo anônimo nesta primeira versão.

A tela `/feedback` é acessível pelo menu principal no desktop e por um atalho no cabeçalho em telas menores. O envio não chama a IA e não altera o plano. Os relatos são consolidados no PostgreSQL e não geram e-mail individual; um job separado reúne até 50 relatos pendentes em um único resumo semanal via Resend, destinado somente ao endereço administrativo configurado em `FEEDBACK_DIGEST_TO`.

## ADR-007 — Resumo semanal de feedback

**Status:** Implementada no código; aguardando migrações `000013`/`000014`, configuração do destinatário e habilitação do timer na produção.

O resumo é executado fora do processo HTTP, como um comando de curta duração (`cadencia-feedback-digest`) acionado semanalmente por systemd. O comando consulta no máximo 50 relatos ainda não enviados, escapa o conteúdo no HTML, envia uma mensagem HTML e texto pelo remetente Resend já verificado e só grava `digest_sent_at` depois de o envio retornar sucesso. Se `FEEDBACK_DIGEST_TO` estiver vazio, a execução termina sem consultar o banco nem enviar mensagens. O serviço usa a rede Docker interna, não publica portas e não fica residente, reduzindo consumo na VPS.

O timer está agendado para segunda-feira às 11:00 UTC (08:00 no horário de São Paulo) e possui `Persistent=true` para executar após uma indisponibilidade. O endereço administrativo é único nesta primeira versão; a quantidade de mensagens permanece previsível e deve ser acompanhada junto das demais aplicações que usam a mesma conta Resend. Um painel administrativo protegido continua como possível evolução, não como dependência do MVP.

## Estado de produção

Em 3 de setembro de 2026, o commit `bef4e60` foi implantado na VPS Oracle. A imagem da API foi reconstruída com Go 1.25; o serviço Ollama foi adicionado isoladamente e o modelo foi preparado sem publicar novas portas. Após a medição de capacidade, o Ollama foi parado e a API passou a usar temporariamente o Worker remoto com `AI_ENABLED=true` e `AI_PROVIDER=worker`.

Na mesma data, a versão `35f24685` do Worker ajustou o provedor Groq para `max_completion_tokens: 512` e `reasoning_effort: 'low'`. O Worker rejeita respostas com `finish_reason` diferente de `stop`, mantendo o fallback determinístico como proteção contra truncamento. A sessão autenticada confirmou explicações completas para treinos de base e subidas.

Validações realizadas:

- API interna `/health`: `200`.
- API interna `/ready`: `200`.
- API pública e frontend público: `200`.
- API, frontend e PostgreSQL saudáveis; túnel ativo.
- PostgreSQL sem porta publicada pelo Cadência.
- Dependabot do GitHub: 0 alertas abertos e 39 fechados.
- `go test ./...`, build Docker e `govulncheck` concluídos.

## Backups

- `cadencia-backup.timer` está habilitado na VPS e executa diariamente às 03:30 UTC.
- Os dumps ficam em `/var/backups/cadencia`, com retenção de 14 dias.
- O script valida cada arquivo com `pg_restore --list`.
- O backup preventivo do último deploy foi `cadencia-20260902T104801Z.dump`.
- O teste completo de restauração foi concluído em 2 de setembro de 2026 com `cadencia-20260902T104801Z.dump`: a restauração em PostgreSQL 17 temporário terminou sem erro, validou 15 tabelas públicas e `cadencia_schema_migrations`, e o container temporário foi removido sem tocar a produção.

## Segurança operacional da VPS

A composição do Cadência está isolada, mas a VPS também hospeda outros aplicativos. A auditoria encontrou portas Docker publicadas para Rotas, Despesas, Estoque, n8n, Immich, Uptime Kuma, Tecboard e Nginx Proxy Manager.

O teste externo confirmou a porta SSH `22` e a porta `2283` do Immich; as demais portas testadas não responderam externamente. A política local de entrada ainda é permissiva (`accept`), não há `ufw` instalado e a cadeia `DOCKER-USER` está vazia.

A auditoria somente leitura de 2 de setembro de 2026 encontrou também Home Assistant na porta `8123` (processo Python em `/config`), `casaos-gateway` na `8888` e a publicação Docker do Immich na `2283`. O Nginx Proxy Manager encaminha `casaos.oraclecloud.com.br` para `100.67.151.30:8888` e `immich.photo.com.br` para `100.67.151.30:2283`, usando Tailscale. O container `cloudflared-tunnel` é separado do túnel dedicado do Cadência e possui o Immich como origem; seus logs indicaram reconexão normal e uma versão desatualizada do cliente. Os testes externos confirmaram inicialmente apenas as portas `22` e `2283` acessíveis. Uma regra persistente na cadeia `DOCKER-USER` passou a bloquear `tcp/2283` somente pela interface pública `enp0s6`; a porta continua acessível por Tailscale e loopback, e o Immich permaneceu respondendo. Em seguida, novas conexões SSH foram bloqueadas na interface pública `enp0s6`, mantendo o acesso administrativo por Tailscale; as regras TCP públicas `22`, `81`, `2283`, `8096` e `8097` foram removidas da Oracle Cloud, restando somente ICMP.

Nenhuma porta de outro aplicativo deve ser bloqueada sem mapear antes seus domínios, túneis, proxies e necessidade de acesso. O próximo hardening deve preservar Tailscale, Cloudflare e os aplicativos existentes.

## Próximas decisões

1. Definir cópia externa dos backups e o monitoramento de falhas.
2. Escolher a política de firewall e a lista mínima de portas públicas da VPS.
3. Registrar monitoramento e alertas de saúde/backup.
4. Monitorar o Worker remoto na sessão autenticada e seus limites; manter o Ollama parado até existir capacidade adequada ou uma otimização de inferência. Reverter para `AI_ENABLED=false` se o provedor remoto não for aprovado.
5. Evoluir coleta de dados e sessões específicas de ciclismo.
