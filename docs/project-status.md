# Estado atual do projeto Cadência

Última atualização: 5 de setembro de 2026.

Este é o documento principal de continuidade. Ele registra o que está implementado, validado, publicado e pendente. Não incluir senhas, tokens, chaves de API ou conteúdo de arquivos `.env`.

## Resumo executivo

O MVP de ciclismo está em produção real e foi validado no navegador e em um celular. O fluxo de cadastro, confirmação de e-mail, perfil, geração e ativação de plano, execução, feedback, adaptação, histórico, evolução e logout está funcionando.

O motor atual é determinístico (`rules-v1`), baseado em regras explícitas e referências científicas. A camada opcional de IA explicativa foi preparada no backend, validada por uma rota remota protegida e preparada para uso local com Ollama. O padrão do código continua desligado; na produção, o Worker remoto está temporariamente selecionado como provedor para evitar consumo elevado da VPS.

## Repositório e produção

- Repositório: <https://github.com/Saulorangel87/App-de-treino>
- Branch de produção: `master`.
- Frontend: <https://cadencia.devsaulo.com.br>
- API: <https://cadencia-api.devsaulo.com.br>
- VPS: Oracle Cloud, Ubuntu, acesso administrativo por SSH na porta 22.
- Código na VPS: `/home/ubuntu/apps/cadencia`.
- Commit implantado: `c768ef7 chore: registra novidades do catalogo de ciclismo`; a API funcional foi reconstruída a partir de `5fbc668`.
- O Cloudflare Tunnel dedicado expõe somente frontend e API; o PostgreSQL não possui hostname, rota pública ou porta publicada.

## Estado do checkout local

- Checkout local no commit `a4e5f9f feat: registra histórico de aderência e carga em 7 28 e 42 dias`. A terceira fatia descrita abaixo está modificada localmente, sem commit nem deploy; não confundir nenhuma delas com a versão publicada na VPS.
- A sequência recente inclui `5461cd6` (contexto por modalidade), `49f1dbd` (catálogo de evidências), `4683999` (piloto de estrada), `5fbc668` (adaptação de recuperação) e `c768ef7` (nota de atualização).
- A migração `000015`, o catálogo inicial e o protocolo `road_moderate_intervals` foram aplicados e publicados na produção após revisão, backup, validação e autorização explícita.
- Protocolos adicionais continuam exigindo revisão própria de elegibilidade, segurança, evidência e atualização das notas de versão do produto.

## Arquitetura efetiva

### Primeira fatia de melhorias — prontidão observacional (local)

- Novos rascunhos registram `prescription_snapshot.readiness_assessment`, com versão `readiness-v1`, modo `observation`, horário UTC, estado, motivos, dados ausentes, cobertura e fatores ainda não avaliados.
- Classificação independente de experiência e avaliação submáxima: `insufficient_data`, `caution`, `recovery_needed` ou `stable`. `stable` significa apenas ausência de alertas nos agregados disponíveis, não aptidão atual ou liberação de carga.
- O repositório passou a contar campos válidos de sessões e fadiga dos check-ins. Médias que ignoram campos nulos não são tratadas como cobertura completa.
- `progression_eligible` permanece `false` nesta leitura. O campo não é uma nova trava do motor: os treinos e as proteções existentes continuam sendo prescritos pelo `rules-v1`, sem alteração de duração, RPE ou blocos nesta entrega.
- Baixa consistência e prontidão para progressão não são inferidas da quantidade de registros. Faltam aderência, tolerância, destreinamento, tendências 7/28/42 dias, recência da avaliação e sinais recentes de sono/estresse/fadiga.
- Planos antigos não são reclassificados; o snapshot só nasce ao gerar outro rascunho. Não há migração, nova infraestrutura nem mudança visual, portanto a nota do produto continua em `0.7.0`. A próxima mudança visível deverá atualizar `APP_VERSION` e `UPDATE_NOTES`.
- Validação: `go test -count=1 ./...` e `go vet ./...` passaram, incluindo cenários de dados ausentes/inconsistentes, dor, fadiga, independência de experiência, serialização do snapshot e regressão do plano. O YAML do OpenAPI foi carregado com sucesso e a referência de `ReadinessAssessment` foi conferida. A indisponibilidade temporária do Go não persistiu na execução final; sua instalação não foi alterada por esta tarefa.
- Após o proprietário ligar o Docker, `pwsh -NoProfile -File scripts/test-readiness-queries.ps1` passou nos quatro cenários: completo, campos incompletos/zero/nulos, ausência de histórico e somente check-ins. O script extrai as consultas reais do repositório e as executa no PostgreSQL com CTEs fictícias em transação somente leitura, verificando também exclusão de sessões canceladas, antigas e de outro atleta. Nenhum dado real foi lido ou alterado. A persistência ponta a ponta também foi confirmada: o mesmo plano, `assessed_at`, classificação, motivos, lacunas e treinos retornaram no `GET /v1/plans/current` após atualizar a tela sem gerar outro plano.
- Detalhes e limites em [`training-adaptation-rules.md`](training-adaptation-rules.md#leitura-observacional-de-prontidão-readiness-v1).

Arquivos desta fatia: `backend/internal/planning/readiness.go`, `backend/internal/planning/readiness_test.go`, `backend/internal/planning/service.go`, `backend/internal/repository/planning.go`, `scripts/test-readiness-queries.ps1`, `api/openapi.yaml`, `docs/README.md`, `docs/project-status.md`, `docs/architecture-decisions.md`, `docs/training-adaptation-rules.md` e `planejamento.md`. `melhorias.md`, as migrações e os arquivos de infraestrutura não foram alterados.

### Segunda fatia de melhorias — histórico 7/28/42 e aderência (commit local)

- Novos rascunhos registram `prescription_snapshot.training_history`, versão `training-history-v1`, com janelas cumulativas de 7, 28 e 42 dias. `used_for_prescription` permanece `false`; `rules-v1` não foi alterado.
- Aderência observada separa sessões previstas já fechadas em concluídas, canceladas, pendentes vencidas e em andamento vencidas. Treino aberto na data atual não reduz antecipadamente a taxa. Rascunhos e planos cancelados ficam fora do denominador.
- Carga realizada usa sessões concluídas na janela móvel de `completed_at`. O cálculo `duração positiva × RPE real de 1 a 10` é salvo em unidades arbitrárias de session-RPE, junto da quantidade de sessões com e sem cobertura válida.
- Aderência e carga têm bases temporais diferentes e isso fica explícito no snapshot. Pedais fora do app não são conhecidos; portanto a taxa descreve somente o plano registrado e não representa toda a rotina do ciclista.
- Não foram criados limites de “boa” ou “má” aderência, inferência de destreinamento, ACWR, progressão ou regressão. As janelas são medições sobrepostas para futura comparação contextual, não autorização de carga.
- Sem migração, mudança visual ou infraestrutura. `APP_VERSION` e `UPDATE_NOTES` permanecem inalterados nesta fatia interna.
- Validação concluída: `go test -count=1 ./...`, `go vet ./...`, build do frontend e carregamento do OpenAPI passaram. Depois das últimas alterações defensivas, `internal/planning` e `internal/repository` passaram novamente. Uma repetição final de `internal/httpapi`, que não foi alterado nesta fatia, foi bloqueada pelo Controle de Aplicativos do Windows ao executar o binário temporário; o mesmo pacote havia passado na suíte completa anterior, sem falha de teste. `scripts/test-training-history-query.ps1` executou a consulta real com CTEs sintéticas em transação somente leitura e confirmou 7/28/42 dias, fechamento do dia atual, vencidos, cancelados, sessão futura, campos nulos e isolamento de outro atleta. O teste detectou e levou à correção da contagem de duração nula. Nenhum dado real foi lido ou alterado.
- Verificação manual concluída no `GET /v1/plans/current`: retornaram `training-history-v1`, as três janelas 7/28/42, `data_issues: []` e `used_for_prescription: false`. Na conta de teste, as cinco sessões realizadas tinham duração zero e nenhuma sessão planejada já fechada dentro das janelas; por isso a taxa veio `null`, a carga veio zero e as lacunas de histórico/cobertura foram registradas. Esse resultado é esperado e evita interpretar dados ausentes como carga real igual a zero.

Arquivos desta fatia: `backend/internal/planning/history.go`, `backend/internal/planning/history_test.go`, `backend/internal/planning/readiness.go`, `backend/internal/planning/readiness_test.go`, `backend/internal/planning/service.go`, `backend/internal/repository/planning.go`, `frontend/lib/planning.ts`, `api/openapi.yaml`, `scripts/test-training-history-query.ps1`, `README.md`, `docs/README.md`, `docs/project-status.md`, `docs/architecture-decisions.md`, `docs/training-adaptation-rules.md` e `planejamento.md`. Não foram alterados `melhorias.md`, migrações, infraestrutura ou notas de versão.

### Terceira fatia de melhorias — qualidade temporal e sinais protetivos (local)

- Novos rascunhos usam `training-history-v2`. Planos já persistidos com `training-history-v1` continuam válidos e não são recalculados.
- `temporal_quality` registra a última sessão concluída, a última sessão com carga session-RPE válida, o último check-in e os dias desde cada registro. Sessões concluídas no futuro e check-ins com data futura são excluídos e sinalizados em `data_issues`.
- Cada janela passa a separar registros de feedback, feedbacks completos, dor, fadiga de 4 a 5 e RPE real pelo menos dois pontos acima do planejado. Também conta check-ins totais/completos, check-ins com ao menos um sinal protetivo e aqueles que já atendem à regra existente de necessidade de recuperação.
- Os limiares são somente os já usados pelo `rules-v1`; não foi criado um novo corte fisiológico. Dor continua contada mesmo se outro campo do feedback estiver ausente, e a falta parcial entra em `missing_data`.
- `not_evaluated` mantém explícitos tolerância à carga, destreinamento, mudança de condicionamento, atividades fora do Cadência, fuso do atleta e progressão. A ausência de atividade no app significa apenas lacuna de registros, não interrupção comprovada do treino.
- `used_for_prescription` continua `false`; o `rules-v1`, a prontidão, os treinos e as adaptações não foram alterados. Não há ACWR, mudança visual, migração ou infraestrutura; `APP_VERSION` e `UPDATE_NOTES` permanecem em `0.7.0`.
- Validação concluída: `go test -count=1 ./...`, `go vet ./...`, build do frontend, carregamento do OpenAPI e a consulta PostgreSQL com fixtures sintéticas em transação somente leitura passaram. A consulta cobre janelas, cobertura incompleta, sinais protetivos, registros futuros e isolamento de outro atleta. O lint geral mantém erros preexistentes em telas/componentes fora desta fatia; `frontend/lib/planning.ts` não acrescentou erro. Na conferência manual, um novo rascunho retornou `training-history-v2`, `data_issues: []` e `used_for_prescription: false`; os cinco feedbacks apareceram completos, o check-in entrou como sinal protetivo/necessidade de recuperação e a carga permaneceu indisponível porque as cinco sessões realizadas tinham duração zero.

Arquivos desta fatia: `backend/internal/planning/history.go`, `backend/internal/planning/history_test.go`, `backend/internal/repository/planning.go`, `frontend/lib/planning.ts`, `api/openapi.yaml`, `scripts/test-training-history-query.ps1`, `README.md`, `docs/README.md`, `docs/project-status.md`, `docs/architecture-decisions.md`, `docs/training-adaptation-rules.md` e `planejamento.md`. Não foram alterados `melhorias.md`, migrações, infraestrutura ou notas de versão.

### Topologia mantida

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
- `database/migrations/`: migrações PostgreSQL até `000015`; as `000013` e `000014` sustentam feedback e resumo semanal, e a `000015` registra as fontes científicas do catálogo. Todas estão aplicadas na produção.
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
- Sessões específicas para perfis adequados: cadência, subidas, sweet spot, ritmo de prova e intervalos controlados. O piloto `road_moderate_intervals` e suas referências estão publicados, com a elegibilidade e os limites documentados.
- Histórico em `/atividades` e agregações observadas em `/evolucao`.
- A aba `/evolucao` também compara as últimas sessões concluídas com a prescrição original (tempo, RPE e métricas registradas), sem transformar a diferença em ajuste automático.
- O plano exibe o resumo do contexto observado usado na geração do ciclo, com sessões, minutos, RPE, check-ins e alertas conservadores de recuperação quando aplicável.
- Indicadores de consistência, carga semanal, prontidão e explicabilidade.
- Aba `/feedback` para o atleta registrar uma experiência, problema ou sugestão com nota de 1 a 5. O relato fica vinculado à conta no PostgreSQL, sem coleta adicional de contato nesta primeira versão.
- Resumo semanal de feedback implementado como comando separado (`cadencia-feedback-digest`). Ele busca até 50 relatos ainda não enviados, envia um único e-mail pelo Resend ao endereço `FEEDBACK_DIGEST_TO` e marca os registros somente depois de um envio bem-sucedido. O serviço não é iniciado junto da API e fica desativado quando o destinatário não está configurado.
- Contrato inicial da IA explicativa no backend, com Ollama opcional, limites de recurso e fallback determinístico para as regras. O cliente do Worker Cloudflare está preparado como provedor remoto, e a rota protegida `/cadencia/explanation` foi publicada sem alterar o endpoint legado. O segredo `CADENCIA_WORKER_TOKEN` foi configurado no Worker e na VPS; uma chamada sintética autenticada respondeu `200` usando `openai/gpt-oss-20b`. O serviço Ollama foi instalado na composição de produção, sem porta pública, e o modelo `qwen3:4b-instruct` foi baixado e testado, mas permanece parado para não pressionar a VPS. A variável `AI_ENABLED` está `true` na VPS com `AI_PROVIDER=worker`; o valor seguro padrão permanece `false`.

### Interface e PWA

- Dashboard, plano, atividades, avaliação, recuperação e evolução.
- Modal de sessão no mobile, check visual de treinos concluídos e logout.
- O painel principal prioriza a sessão em andamento antes de procurar o próximo treino planejado, mantendo o estado consistente após iniciar pela tela inicial.
- Os gráficos semanais da Evolução exibem o intervalo completo de cada semana para deixar claro que os valores são agrupados por período de sete dias.
- Informativo de novidades versionado no primeiro acesso autenticado: aparece uma vez por conta e versão neste navegador, com linguagem simples e os principais recursos da atualização. Cada funcionalidade visível deve atualizar `APP_VERSION` e `UPDATE_NOTES` na mesma entrega; essa exigência ainda fica pendente para qualquer publicação do piloto local.
- PWA instalável, manifesto, ícones e tela offline segura.
- Cache offline limitado a recursos estáticos; dados autenticados não entram no cache.
- Interface em português do Brasil, responsiva e sem rolagem horizontal indevida no mobile; os gráficos que precisam mostrar oito períodos usam rolagem interna controlada.

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
- Feedback de produto: `POST /v1/feedback`.

## Banco e migrações

- Migrações versionadas no checkout local: `000001` a `000015`. A `000013` cria os relatos de feedback, a `000014` adiciona o controle de envio do resumo semanal e a `000015` registra fontes científicas do catálogo de ciclismo.
- Em produção, estão aplicadas `000001` a `000015`. A `000015` foi executada pelo perfil `maintenance` após backup verificável, revisão e autorização explícita; novas migrações devem continuar seguindo essa ordem operacional.
- `000012` adiciona confirmação de e-mail e recuperação de senha.
- Produção possui registro de migrações em `cadencia_schema_migrations`.
- O usuário da API não é superusuário; o proprietário do banco é reservado para operações administrativas.
- Não há alterações de esquema pendentes na produção. Novas migrações devem continuar sendo executadas em ordem pelo perfil `maintenance`, após backup verificável.

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

Em 3 de setembro de 2026, o commit `33de28a` foi atualizado na VPS por fast-forward. Foi criado e verificado o backup preventivo `cadencia-20260904T023630Z.dump` (UTC), somente o container `cadencia-frontend-1` foi reconstruído e recriado, e o túnel permaneceu ativo. O frontend interno e as rotas públicas `/` e `/evolucao` retornaram HTTP 200; API, frontend, PostgreSQL e túnel permaneceram saudáveis. O ajuste corrige a sobreposição dos períodos e da barra de rolagem nos gráficos em telas pequenas.

Em 4 de setembro de 2026, o commit `5fbc668` foi atualizado na VPS por fast-forward. O backup preventivo `cadencia-20260905T003553Z.dump` (UTC) foi criado e verificado, a migração `000015` foi aplicada pelo perfil `maintenance` e as imagens de API e frontend foram reconstruídas. Os containers `cadencia-api-1` e `cadencia-frontend-1` foram recriados; PostgreSQL e túnel permaneceram ativos. A API respondeu `{"status":"ready"}` no endpoint interno `/ready`, e os domínios públicos retornaram HTTP 200.

Na sequência, o commit `c768ef7` atualizou somente o frontend para publicar a versão `0.7.0` e a nota sobre o catálogo baseado em evidências. A página pública passou a entregar a versão `0.7.0`; a tela de novidades foi confirmada em uma conta autenticada.

Durante uma tentativa inicial, uma cópia privada do frontend foi publicada por engano no ambiente Sites, fora da infraestrutura oficial. Ela foi excluída manualmente pelo proprietário e não tinha acesso público. O Sites não faz parte do fluxo de produção do Cadência; futuras publicações devem usar exclusivamente a VPS Oracle e o Cloudflare Tunnel dedicado.

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
- Backup preventivo do deploy de 3 de setembro: `cadencia-20260904T023630Z.dump` (UTC).
- O backup preventivo anterior, `cadencia-20260902T104801Z.dump`, permanece registrado e validado.
- O teste de restauração completo foi concluído em 2 de setembro de 2026 com o dump `cadencia-20260902T104801Z.dump`: um PostgreSQL 17 temporário restaurou o arquivo com sucesso, apresentou 15 tabelas públicas e `cadencia_schema_migrations`, e foi removido ao final. O banco e os volumes de produção não foram alterados.

## Auditoria da VPS e pendências operacionais

O Cadência está isolado, mas a VPS hospeda outros aplicativos. Foram observadas portas Docker publicadas para serviços como Rotas, Despesas, Estoque, n8n, Immich, Uptime Kuma, Tecboard e Nginx Proxy Manager.

O teste externo realizado a partir do ambiente atual confirmou acesso à porta 22 e à porta 2283 (Immich); as demais portas testadas não responderam externamente. A política local de entrada ainda está permissiva (`accept`), não há `ufw` instalado e a cadeia `DOCKER-USER` está vazia.

Na auditoria somente leitura de 2 de setembro de 2026, o host também apresentou Home Assistant na porta 8123, além das portas já catalogadas. O processo nessa porta é o Python do Home Assistant (`/config`); a porta 8888 pertence ao `casaos-gateway` e a 2283 é a publicação Docker do Immich. O Nginx Proxy Manager possui os destinos `casaos.oraclecloud.com.br` → `100.67.151.30:8888` e `immich.photo.com.br` → `100.67.151.30:2283`, ambos por Tailscale. Existe um segundo container Cloudflare Tunnel, separado do túnel dedicado do Cadência, que encaminha o Immich; seus logs recentes registraram reconexão bem-sucedida e aviso de versão desatualizada. Os testes externos confirmaram inicialmente apenas 22 e 2283 acessíveis. Em seguida, foi aplicada uma regra persistente na cadeia `DOCKER-USER` para bloquear `tcp/2283` somente pela interface pública `enp0s6`; as regras TCP públicas `22`, `81`, `2283`, `8096` e `8097` também foram removidas da Oracle Cloud, restando somente ICMP. Immich continua respondendo via Tailscale e loopback, e Cadência foi validado após as alterações.

Em 3 de setembro de 2026, uma auditoria somente leitura pelo Tailscale mediu 2 vCPUs, 11 GiB de RAM total, 8,2 GiB disponíveis, nenhum swap e 118 GiB livres no disco raiz (40% usado). A carga estava baixa (0,13 / 0,05 / 0,01), mas a ausência de swap exige cautela. Havia 23 containers ativos; os maiores consumos observados foram Immich (servidor e machine learning), n8n e Home Assistant. O Ollama foi instalado isoladamente no serviço `cadencia-ollama-1`, limitado a 4 GiB de memória, 1 CPU e uma chamada simultânea, somente na rede Docker interna e sem porta publicada. O modelo `qwen3:4b-instruct` foi baixado e uma inferência simples retornou `OK`. Com o modelo carregado, o container chegou a aproximadamente 3,3 GiB e uma chamada levou cerca de 72 segundos, consumindo praticamente 100% do limite de CPU; por isso o serviço foi parado após o teste. A API foi recriada com `AI_ENABLED=true` e `AI_PROVIDER=worker`, permaneceu saudável e respondeu `/ready`; a VPS voltou a aproximadamente 8,1 GiB disponíveis.

Também em 3 de setembro, a versão ativa do Worker `flat-rice-6724` foi atualizada para o modelo Groq disponível `openai/gpt-oss-20b`, após o identificador anterior deixar de existir. A rota protegida foi testada com payload mínimo e token válido, sem expor o segredo; o endpoint legado do Worker permaneceu preservado.

Após o teste autenticado mostrar respostas interrompidas, a versão `35f24685` do Worker passou a usar `max_completion_tokens: 512` e `reasoning_effort: 'low'`. A rota também rejeita respostas cujo `finish_reason` não seja `stop`, permitindo que a API use o fallback determinístico em vez de exibir texto incompleto. Os testes autenticados de 3 de setembro para “Giro de base” e “Subidas controladas” retornaram explicações completas.

Pendências, sem executar bloqueios automáticos:

1. Completar o mapa de cada domínio, túnel e proxy dos aplicativos existentes.
2. Monitorar as regras de rede da Oracle Cloud; as regras TCP públicas desnecessárias já foram removidas.
3. Verificar periodicamente o acesso administrativo pelo Tailscale; novas conexões SSH pela interface pública já estão bloqueadas.
4. Fechar outras portas diretas desnecessárias, especialmente serviços que já usam proxy; `2283` do Immich já está bloqueada na interface pública.
5. Definir cópia externa dos backups e monitoramento de falhas.
6. Atualizar esta documentação após cada mudança de infraestrutura.

## Feedback de produto e recebimento dos relatos

O fluxo inicial foi desenhado para a divulgação do MVP em grupos de ciclismo: cada pessoa cria uma conta, abre a aba `Feedback` e envia uma categoria, uma nota e um relato livre. O backend exige autenticação, valida tamanho e categoria e armazena o registro em `user_feedback` ligado ao usuário; ele não altera o plano nem dispara IA.

Nesta primeira etapa, os relatos continuam centralizados no banco e não geram um e-mail individual. O recebimento escolhido é um resumo semanal por e-mail, enviado pelo job separado ao endereço administrativo `FEEDBACK_DIGEST_TO`; os relatos são marcados como enviados para não reaparecerem no próximo resumo. O timer systemd está ativo na VPS e o primeiro ciclo será observado quanto a entrega e utilidade. Uma tela administrativa protegida ou exportação controlada pode ser avaliada depois, se o volume justificar. Não expor o banco nem liberar uma listagem administrativa ao usuário final faz parte do escopo de segurança.

## Próximas etapas do produto

1. Preparar a comparação observacional entre períodos não sobrepostos, sem transformar uma razão isolada em autorização de carga ou chamar ausência de registros de destreinamento.
2. Evoluir as regras em versão paralela, preservando `rules-v1` até que a nova versão esteja testada, comparável e auditável.
3. Trabalhar adaptação em ciclo fechado, carga/progressão e integridade dos dados antes de ampliar a prescrição.
4. Reforçar segurança, feedback pós-treino, explicabilidade e auditabilidade das decisões.
5. Ampliar o catálogo de modalidades e protocolos de ciclismo somente com critérios de elegibilidade e referências próprias revisados; o catálogo inicial e o piloto atual já estão publicados.
6. Avaliar integrações externas, como Strava, somente depois de definir escopo, consentimento, custos e segurança dos tokens.
7. Manter o escopo desta fase em ciclismo; corrida e força não entram no próximo ciclo sem nova decisão.

O feedback real e o primeiro envio automático do Resend seguem em paralelo, sem bloquear as melhorias. Os testes manuais de e-mail e de latência/limites/fallback já foram realizados e não precisam ser repetidos como condição para avançar. A operação de produção não foi alterada nesta etapa.

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

1. Leia este arquivo, [`docs/README.md`](README.md), `README.md`, `docs/architecture-decisions.md`, `docs/training-cycle-lifecycle.md`, `docs/training-adaptation-rules.md` e `docs/cycling-evidence-catalog.md`.
2. Confira `git status` e os commits recentes.
3. Preserve os bancos PostgreSQL local e da VPS.
4. Não publique, faça commit ou altere infraestrutura sem autorização explícita.
