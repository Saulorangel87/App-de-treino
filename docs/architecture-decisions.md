# Decisões de arquitetura

Última revisão: 4 de setembro de 2026.

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

As migrações `000001` a `000015` estão versionadas e aplicadas na produção. A `000013` cria os relatos de feedback, a `000014` adiciona o controle de envio do resumo semanal e a `000015` registra fontes científicas do catálogo de ciclismo. O catálogo inicial e o piloto `road_moderate_intervals` foram publicados após revisão e autorização; antes de qualquer nova mudança estrutural em produção, deve existir backup verificável e a migração deve ser executada pelo perfil `maintenance`.

## ADR-006 — Feedback de produto

**Status:** Implementada e publicada; migrações `000013` e `000014` aplicadas na produção.

O feedback solicitado aos primeiros ciclistas é separado do feedback pós-treino. A rota autenticada `POST /v1/feedback` aceita somente uma categoria (`experience`, `bug` ou `suggestion`), uma nota de 1 a 5 e uma mensagem entre 10 e 2000 caracteres. O registro é vinculado ao usuário na tabela `user_feedback`, sem armazenar um e-mail duplicado ou permitir conteúdo anônimo nesta primeira versão.

A tela `/feedback` é acessível pelo menu principal no desktop e por um atalho no cabeçalho em telas menores. O envio não chama a IA e não altera o plano. Os relatos são consolidados no PostgreSQL e não geram e-mail individual; um job separado reúne até 50 relatos pendentes em um único resumo semanal via Resend, destinado somente ao endereço administrativo configurado em `FEEDBACK_DIGEST_TO`.

## ADR-007 — Resumo semanal de feedback

**Status:** Implementada e ativa na produção.

O resumo é executado fora do processo HTTP, como um comando de curta duração (`cadencia-feedback-digest`) acionado semanalmente por systemd. O comando consulta no máximo 50 relatos ainda não enviados, escapa o conteúdo no HTML, envia uma mensagem HTML e texto pelo remetente Resend já verificado e só grava `digest_sent_at` depois de o envio retornar sucesso. Se `FEEDBACK_DIGEST_TO` estiver vazio, a execução termina sem consultar o banco nem enviar mensagens. O serviço acessa o PostgreSQL pela rede Docker interna e o Resend por uma rede de saída dedicada; não publica portas e não fica residente, reduzindo consumo na VPS. A migração e o timer já foram aplicados; o primeiro envio será acompanhado em operação.

O timer está agendado para segunda-feira às 11:00 UTC (08:00 no horário de São Paulo) e possui `Persistent=true` para executar após uma indisponibilidade. O endereço administrativo é único nesta primeira versão; a quantidade de mensagens permanece previsível e deve ser acompanhada junto das demais aplicações que usam a mesma conta Resend. Um painel administrativo protegido continua como possível evolução, não como dependência do MVP.

## ADR-008 — Evolução pós-MVP do motor

**Status:** Aceita como direção planejada; ainda não implementada como substituição do `rules-v1`.

A próxima fase seguirá o roadmap de `melhorias.md` com foco exclusivo em ciclismo: classificação explícita de prontidão, regras versionadas, adaptação em ciclo fechado, progressão/carga, integridade dos dados, segurança, feedback e auditabilidade. O `rules-v1` continuará preservado até que uma evolução paralela esteja testada, comparável e auditável. Dados ausentes ou inconsistentes não devem ser tratados como autorização para aumentar carga. A IA continua explicativa e subordinada às regras validadas; não prescreve, inventa evidências ou diagnostica.

Primeira fatia local, em 5 de setembro de 2026: `readiness-v1` é uma função determinística em `backend/internal/planning/readiness.go`, salva no snapshot do plano em modo `observation`. Não substitui `rules-v1` nem o nível de prontidão calculado pelo check-in diário. Usa somente limitações ativas, os agregados de 28 dias já existentes e contagens explícitas de cobertura dos campos. Não depende do nível de experiência, de aprovação antiga na avaliação ou de volume autodeclarado para concluir que há prontidão. Não há migração nem recálculo retroativo de planos. A ausência de sessões não permite inferir baixa consistência; a progressão permanece não avaliada até haver aderência, tolerância e qualidade temporal dos dados. O recebimento de feedback real e do Resend é acompanhamento paralelo, não pré-requisito de desenvolvimento.

Segunda fatia no commit local `a4e5f9f`, ainda sem publicação: `training-history-v1` mede janelas cumulativas de 7, 28 e 42 dias em modo observacional. A aderência usa a data planejada e apenas sessões já fechadas: treinos de hoje entram depois de concluídos ou cancelados; datas anteriores ainda `planned`/`adapted` são pendências vencidas, e `in_progress` vencido fica separado. A carga realizada usa `completed_at`, duração positiva e RPE real entre 1 e 10; o produto registra `duração × RPE` em unidades arbitrárias e explicita quantas sessões não têm dados suficientes. Planos `draft` ou `cancelled` não compõem o denominador de aderência. Sessões concluídas do atleta continuam compondo a carga realizada independentemente do plano de origem. As duas bases temporais, as lacunas e as referências ficam no snapshot. As janelas não formam ACWR, não geram limiares de aderência e não participam da prescrição nesta fatia.

Terceira fatia local, ainda sem commit ou publicação: `training-history-v2` acrescenta recência, exclusão explícita de registros futuros, cobertura de feedback/check-in e contagens dos sinais protetivos já usados pelo motor. O snapshot diferencia feedback inexistente de feedback incompleto e preserva dor mesmo quando falta outro campo. Ausência de registros no Cadência é rotulada apenas como lacuna de atividade registrada; destreinamento, perda de condicionamento, tolerância e progressão permanecem em `not_evaluated`. A decisão evita extrapolar estudos de cessação total para redução de treino ou falta de sincronização do atleta. `used_for_prescription` permanece `false`, e planos antigos `training-history-v1` continuam legíveis sem recálculo.

Quarta fatia no commit local `810183c`, ainda sem publicação: `training-history-v3` acrescenta `period_comparison` com seis blocos semanais não sobrepostos, do mais recente ao mais antigo. Cada bloco repete medições brutas de aderência, sessões, carga session-RPE, feedback, sinais protetivos e recuperação; não calcula tendência, ACWR ou qualquer razão que autorize carga. Sessões e carga usam intervalos de `completed_at` no relógio do banco, enquanto aderência e recuperação usam datas de `CURRENT_DATE`, mantendo as bases explícitas. O contrato registra `period-comparison-v1`, lacunas e inconsistências, mantém `used_for_prescription: false` e deixa a tendência para prescrição em `not_evaluated`. Snapshots antigos continuam legíveis sem recálculo. A conferência manual via API foi concluída.

Quinta fatia no commit local `64e554d`, ainda sem publicação: `rules-v2` é executado em modo `shadow` durante a geração do plano. Ele avalia integridade dos períodos, sinais protetivos e evidência mínima de dois períodos recentes, registrando regras avaliadas, regras adiadas, motivos, lacunas e uma resposta candidata. Os estados são `protective_signal`, `observation_only` e `not_evaluated`; nenhum deles altera a prescrição. `engine_version` continua `rules-v1`, e `progression_eligible`, `applied` e `used_for_prescription` permanecem `false`. A conferência manual no `GET /v1/plans/current` e a matriz controlada confirmaram esse isolamento; o teste regressivo adicional ainda está sem commit. A decisão permite comparar o comportamento antes de qualquer migração para um motor prescritivo.

## ADR-009 — Comunicação de atualizações no produto

**Status:** Aceita e aplicada.

Toda atualização com funcionalidade visível deve atualizar `frontend/lib/release.ts`, incrementando `APP_VERSION` e registrando a mudança em `UPDATE_NOTES`. O componente `UpdateNotice` apresenta as notas no primeiro acesso autenticado após a versão mudar e registra a confirmação por conta, versão e navegador usando armazenamento local. As notas não devem conter segredos. A versão `0.7.0` registrou o catálogo de ciclismo baseado em evidências e foi confirmada na produção.

## Estado de produção

Em 3 de setembro de 2026, o commit `57c241a` foi implantado na VPS Oracle. A imagem da API, do job de digest e do frontend foi reconstruída com Go 1.25; as migrações `000013` e `000014` foram aplicadas após backup preventivo. O serviço Ollama permanece isolado e parado após a medição de capacidade, e a API usa temporariamente o Worker remoto com `AI_ENABLED=true` e `AI_PROVIDER=worker`.

Ainda em 3 de setembro, o commit `33de28a` foi publicado por fast-forward na mesma VPS para corrigir a legibilidade dos períodos e da barra de rolagem dos gráficos da Evolução em telas pequenas. Somente `cadencia-frontend-1` foi reconstruído e recriado; API, PostgreSQL e túnel permaneceram ativos e saudáveis. As rotas públicas principal e `/evolucao` retornaram HTTP 200. O backup preventivo correspondente foi `cadencia-20260904T023630Z.dump` (timestamp em UTC).

Em 4 de setembro de 2026, o commit `5fbc668` foi publicado por fast-forward na mesma VPS. O backup preventivo `cadencia-20260905T003553Z.dump` (UTC) foi criado e verificado, a migração `000015` foi aplicada pelo perfil `maintenance` e as imagens da API e do frontend foram reconstruídas. Os containers de API e frontend foram recriados; PostgreSQL e túnel permaneceram ativos e saudáveis. A API interna respondeu `{"status":"ready"}` e os dois domínios públicos retornaram HTTP 200.

Na sequência, o commit `c768ef7` atualizou somente o frontend para publicar a versão `0.7.0` e a nota do catálogo. A nota apareceu no primeiro acesso autenticado de teste; o fluxo funcional do check-in de recuperação e os testes de latência, limites e fallback do Worker já haviam sido validados.

O destino oficial de produção é a composição Docker na VPS Oracle, exposta pelos hostnames `cadencia.devsaulo.com.br` e `cadencia-api.devsaulo.com.br` no Cloudflare Tunnel dedicado. Uma publicação privada acidental no Sites, feita durante uma tentativa de deploy, foi excluída pelo proprietário. O Sites não é um destino autorizado para futuras publicações do Cadência.

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
- O backup preventivo do último deploy funcional é `cadencia-20260905T003553Z.dump` (UTC).
- O backup anterior, `cadencia-20260902T104801Z.dump`, permanece registrado e validado.
- O teste completo de restauração foi concluído em 2 de setembro de 2026 com `cadencia-20260902T104801Z.dump`: a restauração em PostgreSQL 17 temporário terminou sem erro, validou 15 tabelas públicas e `cadencia_schema_migrations`, e o container temporário foi removido sem tocar a produção.

## Segurança operacional da VPS

A composição do Cadência está isolada, mas a VPS também hospeda outros aplicativos. A auditoria encontrou portas Docker publicadas para Rotas, Despesas, Estoque, n8n, Immich, Uptime Kuma, Tecboard e Nginx Proxy Manager.

O teste externo confirmou a porta SSH `22` e a porta `2283` do Immich; as demais portas testadas não responderam externamente. A política local de entrada ainda é permissiva (`accept`), não há `ufw` instalado e a cadeia `DOCKER-USER` está vazia.

A auditoria somente leitura de 2 de setembro de 2026 encontrou também Home Assistant na porta `8123` (processo Python em `/config`), `casaos-gateway` na `8888` e a publicação Docker do Immich na `2283`. O Nginx Proxy Manager encaminha `casaos.oraclecloud.com.br` para `100.67.151.30:8888` e `immich.photo.com.br` para `100.67.151.30:2283`, usando Tailscale. O container `cloudflared-tunnel` é separado do túnel dedicado do Cadência e possui o Immich como origem; seus logs indicaram reconexão normal e uma versão desatualizada do cliente. Os testes externos confirmaram inicialmente apenas as portas `22` e `2283` acessíveis. Uma regra persistente na cadeia `DOCKER-USER` passou a bloquear `tcp/2283` somente pela interface pública `enp0s6`; a porta continua acessível por Tailscale e loopback, e o Immich permaneceu respondendo. Em seguida, novas conexões SSH foram bloqueadas na interface pública `enp0s6`, mantendo o acesso administrativo por Tailscale; as regras TCP públicas `22`, `81`, `2283`, `8096` e `8097` foram removidas da Oracle Cloud, restando somente ICMP.

Nenhuma porta de outro aplicativo deve ser bloqueada sem mapear antes seus domínios, túneis, proxies e necessidade de acesso. O próximo hardening deve preservar Tailscale, Cloudflare e os aplicativos existentes.

## Próximas decisões

1. Observar os primeiros relatos reais em `/feedback` e confirmar a entregabilidade/utilidade do resumo semanal pelo Resend.
2. Monitorar a explicação autenticada pelo Worker remoto, sua latência, limites e acionamento do fallback determinístico; manter o Ollama parado.
3. Comparar os resultados shadow em cenários controlados, preservando `rules-v1` e sem alterar carga; essa validação já foi concluída no commit `64e554d` e deve ser mantida como regressão.
4. Só então definir adaptação em ciclo fechado, progressão/carga e barreiras de integridade e segurança.
5. Evoluir o catálogo de protocolos de ciclismo com elegibilidade e evidências próprias; o catálogo inicial e o piloto de estrada já estão em produção, enquanto novos protocolos permanecem condicionados à revisão.
6. Definir cópia externa dos backups, monitoramento de falhas e alertas de saúde.
7. Escolher a política de firewall e a lista mínima de portas públicas da VPS, preservando Tailscale, Cloudflare e os demais aplicativos.
