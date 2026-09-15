# Estado atual do projeto Cadência

Última atualização: 15 de setembro de 2026.

Este é o documento principal de continuidade. Ele registra o que está implementado, validado, publicado e pendente. Não incluir senhas, tokens, chaves de API ou conteúdo de arquivos `.env`.

## Resumo executivo

O MVP de ciclismo está em produção real e foi validado no navegador e em um celular. O fluxo de cadastro, confirmação de e-mail, perfil, geração e ativação de plano, execução, feedback, adaptação, histórico, evolução e logout está funcionando.

O motor atual é determinístico (`rules-v1`), baseado em regras explícitas e referências científicas. A camada opcional de IA explicativa foi preparada no backend, validada por uma rota remota protegida e preparada para uso local com Ollama. O padrão do código continua desligado; na produção, o Worker remoto está temporariamente selecionado como provedor para evitar consumo elevado da VPS.

### Estado canônico publicado — 15 de setembro de 2026

- O escopo do Cadência está encerrado em ciclismo. Corrida e musculação serão produtos separados e não são pendências deste repositório.
- A versão funcional publicada é `0.31.0`, incluindo configurações, encerramento seguro de conta e correções responsivas posteriores.
- As migrações `000001` a `000029` estão aplicadas na produção. O backup preventivo mais recente da sincronização de schema é `cadencia-20260914T224807Z.dump`.
- A leitura autenticada de `GET /v1/plans/current` foi validada após as migrações, além de `/health`, `/ready`, frontend público, API pública e os quatro serviços da composição Docker.
- O MVP de ciclismo está implementado e validado. O que permanece é operação e evolução controlada: feedback real e resumo semanal, cópia externa de backups, hardening da VPS, correção gradual do lint e coleta longitudinal antes de dar autoridade adicional aos shadows.

### Configurações da conta — publicada na versão `0.31.0`

- A rota frontend `/configuracoes` foi implementada com acesso ao perfil, status do e-mail e encerramento definitivo da conta.
- O `DELETE /v1/auth/account` exige a senha atual e a confirmação `ENCERRAR CONTA`, remove o usuário em uma operação atômica e invalida o cookie da sessão.
- O teste transacional `database/tests/account_deletion.sql` confirmou a remoção em cascata de todos os dados pessoais existentes no schema. Não foi criada migração nova.
- A interface possui layout responsivo e acesso compacto no cabeçalho mobile. A validação manual de desktop/mobile e de encerramento com conta descartável foi concluída antes da publicação.

### Entrega local validada — questionário e contexto seguro

- O checkout adiciona o contrato `cycling-onboarding-v2`, consultável em `GET /v1/onboarding/questionnaire`; a tela usa seus gates condicionais para exibir segurança, potência e evento, sem criar modalidades fora do escopo de ciclismo.
- Perfil, segurança e contexto de ciclismo receberam campos opcionais validados; medidas corporais são somente contexto e não alteram automaticamente a prescrição. Rotina sedentária ou ocasional bloqueia qualidade, e objetivo secundário apenas desempata opções já elegíveis.
- A evolução passa a expor objetivos observados em 28 dias, carga sessão-RPE e velocidade média calculada quando há duração e distância reais. Não há porcentagem inventada, diagnóstico ou estimativa de desempenho.
- A migração aditiva `000030_profile_safety_context` ainda não está aplicada na produção, mas foi aplicada e validada no PostgreSQL local. `go test ./...`, `go vet ./...`, `npm run build`, a fixture SQL transacional, a conferência do schema, o teste autenticado do questionário e `scripts/test-local-visual.ps1` passaram no ambiente local. O smoke test visual cobriu `/perfil` em mobile/desktop, `/plano` em mobile com a auditoria expandida e `/evolucao` em mobile; as capturas foram inspecionadas. A conta temporária foi removida pela API, conferida no banco e os artefatos temporários foram limpos. O pacote foi registrado no commit desta etapa; restam backup, publicação autorizada, migração `000030` em produção e leitura autenticada pós-deploy.

### Correção operacional da produção — 14 de setembro de 2026

- A tela de novidades e os endpoints públicos permaneciam disponíveis, mas a conta autenticada recebia `500` em `GET /v1/plans/current`. API, frontend, PostgreSQL, Tunnel, `/health` e `/ready` continuavam saudáveis.
- A causa foi uma defasagem de schema introduzida no último deploy: o código consultava `feedback.satisfaction`, `feedback.terrain` e `feedback.external_conditions`, campos da migração `000023_feedback_context`, enquanto `cadencia_schema_migrations` estava somente até `000021`.
- Foi criado e validado o backup `/var/backups/cadencia/cadencia-20260914T112526Z.dump`. Em seguida, o perfil `maintenance` aplicou em ordem `000022_limitation_context` e `000023_feedback_context`.
- A correção foi confirmada pela presença das três colunas, pelo registro das duas migrações, pelos serviços saudáveis e pelo carregamento do plano na sessão autenticada do navegador. Não houve alteração de prescrição nem geração de plano novo.
- Regra operacional permanente: healthchecks não comprovam compatibilidade do schema. Antes de recriar a API, conferir `cadencia_schema_migrations`, aplicar todos os `.up.sql` pendentes em ordem, validar o registro aplicado e só então executar a checagem autenticada de `/v1/plans/current`.

### Décima sexta fatia de melhorias — distribuição observacional dos estímulos (local)

- Novos rascunhos passam a usar `training-history-v4` e `period-comparison-v2`. Os seis períodos semanais preservam contagem de sessões de qualidade, cobertura de carga, minutos realizados de qualidade, densidade e alta intensidade.
- O bloco `stimulus-distribution-v1` resume a distribuição em 7/14/28/42 dias e observa proximidade entre estímulos por pares em dias consecutivos, menor intervalo e sessão mais recente. O corte operacional é RPE-alvo `>= 6,0` para qualidade e `>= 7,0` para alta intensidade; não representa zona fisiológica.
- Sessões futuras e registros explicitamente inelegíveis pelo `data-integrity-v1` ficam fora. Datas são normalizadas para UTC porque o produto ainda não possui fuso individual. Lacunas de duração/RPE e de datas de qualidade ficam explícitas, sem reclassificar ou apagar atividades.
- `rules-v1` continua sendo a única fonte prescritiva. `mode: observation`, `progression_eligible`, `applied` e `used_for_prescription` não recebem autoridade por causa desta medição; não há mudança visual, release, migração, deploy ou infraestrutura.
- As alterações desta fatia ainda estão no checkout local e aguardam commit autorizado; a produção permanece em `0.20.0` sem essa medição.

Validação desta fatia: `go test -count=1 ./...`, `go vet ./...`, `npm run build`, lint direcionado de `frontend/lib/planning.ts`, validação de referências do OpenAPI, a fixture PostgreSQL somente leitura e `git diff --check` passaram. O lint geral ainda aponta débitos preexistentes em componentes/páginas não tocados nesta fatia. Não é necessário teste no navegador porque nenhum comportamento visível foi alterado.

### Décima sétima fatia de melhorias — gate observacional de distribuição dos estímulos (publicada sem mudança de versão)

- O `rules-v2` e o `rules-v2-adaptation-v1` agora consomem o bloco `stimulus-distribution-v1` em paralelo, sem substituir o `rules-v1`.
- A avaliação observa dois padrões operacionais: pelo menos duas sessões de qualidade nos últimos 7 dias e sessões de qualidade em dias consecutivos dentro dos 42 dias observados. Esses padrões geram `prefer_recovery` somente no shadow; não são diagnóstico nem limiar fisiológico universal.
- Se a cobertura de períodos, datas de qualidade ou consistência estiver incompleta, a progressão shadow fica `not_evaluated`/`defer_progression` e explicita as lacunas. Ausência de sessões de qualidade, com histórico íntegro, não é tratada como bloqueio.
- O `decision_audit` passa a registrar `stimulus_distribution` como dado observado e `stimulus_distribution_gate` como restrição quando o gate é acionado. `progression_eligible`, `applied` e `used_for_prescription` continuam falsos.
- A alteração é somente backend/contrato: não há migração, mudança visual ou nota de versão. O commit `d1cc7e4` foi implantado pelo proprietário; a produção continua na versão visível `0.20.0`.

Validação desta fatia: `go test -count=1 ./...`, `go vet ./...`, `npm run build`, lint direcionado de `frontend/lib/planning.ts` e `git diff --check` passaram. O lint geral mantém apenas pendências preexistentes fora dos arquivos tocados. Não é necessário teste no navegador porque nenhum comportamento visível foi alterado.

### Verificação pós-deploy da fatia shadow

- A tela autenticada `/plano` em produção carregou após o deploy, manteve `Regras V1` como motor visível e não apresentou regressão aparente no fluxo do plano.
- A tentativa de abrir diretamente a rota autenticada da API foi bloqueada pelo cliente do navegador; a tentativa equivalente no PowerShell local falhou na negociação TLS. Portanto, esta sessão confirma o carregamento da aplicação, mas não reivindica uma inspeção independente do JSON de `stimulus_distribution` em produção.
- A validação controlada local confirmou os campos `stimulus_distribution_gate`, `mode: shadow` e `used_for_prescription: false` pelos testes do motor. A inspeção independente do JSON em produção ainda depende de uma requisição da própria sessão autenticada; não é necessário gerar treino artificial nem modificar dados reais.

### Décima oitava fatia de melhorias — auditoria observacional da periodização (publicada sem mudança de versão)

- Cada rascunho passa a registrar `periodization-shadow-v1` no `prescription_snapshot`, resumindo as quatro semanas planejadas, as fases amplas, o volume, a qualidade, a recuperação, os treinos longos, o taper e o espaçamento entre estímulos exigentes.
- O auditor verifica a presença das quatro semanas, a ausência de qualidade na semana de recuperação, a redução do volume dessa semana, a quantidade de sessões de qualidade por semana e o espaçamento entre elas. Lacunas ficam em `missing_data` e incoerências em `data_issues`.
- O resultado é exclusivamente observacional: `mode: shadow`, `progression_eligible: false`, `applied: false` e `used_for_prescription: false`. O `rules-v1`, os treinos gerados, o calendário e o comportamento prescritivo permanecem inalterados.
- Foram adicionadas regressões para ciclo coerente, semanas ausentes, qualidade indevida na recuperação e presença do snapshot sem substituir o `rules-v1`. Não há migração, mudança visual ou atualização de `APP_VERSION`. O commit `b32a3c9` foi publicado após backup e validação operacional; a versão visível permanece `0.20.0`.

Validação desta fatia: `go test -count=1 ./...`, `go vet ./...`, `npm run build`, lint direcionado de `frontend/lib/planning.ts`, validação das referências do OpenAPI e `git diff --check` passaram. A aplicação pública e a API permaneceram saudáveis após o deploy. A inspeção independente do campo autenticado `periodization_shadow` ainda depende de uma requisição da própria sessão do navegador; isso não bloqueia a documentação nem transforma o shadow em prescrição.

### Décima nona fatia de melhorias — seleção observacional de estímulos (publicada sem mudança de versão)

- Cada novo rascunho passa a registrar `stimulus-selection-shadow-v1`, relacionando a necessidade inferida do contexto às famílias de estímulos selecionadas pelo `rules-v1`.
- A auditoria observa proteção/recuperação, aderência, especificidade de evento, progressão de qualidade, estímulos esperados e estímulos selecionados. Incoerências ficam em `data_issues` e não geram troca automática de sessão.
- O resultado permanece `mode: shadow`, com `progression_eligible: false`, `applied: false` e `used_for_prescription: false`. O `rules-v1` continua sendo a única fonte prescritiva.
- O contrato foi atualizado no OpenAPI e no tipo compartilhado do frontend. Não há migração, mudança visual, atualização de `APP_VERSION` ou nota de versão.

Validação local: `go test -count=1 ./...`, `go vet ./...`, `npm run build`, `git diff --check` e referências do OpenAPI passaram. O commit `c5d8822` foi publicado após backup e validação operacional; a versão visível permanece `0.20.0`.

### Vigésima fatia de melhorias — coerência integrada dos shadows (publicada; validada)

- O novo bloco `planning-coherence-shadow-v1` resume periodização, distribuição histórica e seleção de estímulos no mesmo resultado de geração de plano.
- A auditoria registra os estados dos três componentes, verificações de coerência, sinais de densidade e a semana de recuperação. Divergências produzem `observed_mismatch`/`review_coherence`; cobertura insuficiente produz `not_evaluated`/`defer_evaluation`.
- A integração preserva `progression_eligible: false`, `applied: false` e `used_for_prescription: false` em todos os componentes. O `rules-v1` permanece como única fonte prescritiva.
- O contrato foi atualizado no OpenAPI e no tipo compartilhado do frontend. Não há migração, mudança visual, atualização de `APP_VERSION`, release, deploy ou alteração de infraestrutura.

Validação local: `go test -count=1 ./...`, `go vet ./...`, `npm run build`, validação das referências do OpenAPI e `git diff --check` passaram. O commit `f0fec8b` foi publicado após backup preventivo; não houve migração nova, mudança visual ou atualização de versão. API, frontend, PostgreSQL e Tunnel ficaram saudáveis, `/ready` respondeu corretamente e os dois domínios públicos retornaram HTTP 200.

## Repositório e produção

- Repositório: <https://github.com/Saulorangel87/App-de-treino>
- Branch de produção: `master`.
- Frontend: <https://cadencia.devsaulo.com.br>
- API: <https://cadencia-api.devsaulo.com.br>
- VPS: Oracle Cloud, Ubuntu, acesso administrativo por SSH na porta 22.
- Código na VPS: `/home/ubuntu/apps/cadencia`.
- Linha funcional implantada: versão `0.31.0`, incluindo configurações, encerramento seguro de conta e correções responsivas posteriores.
- O backup preventivo mais recente é `cadencia-20260914T224807Z.dump`. Após a aplicação ordenada das migrações `000024` a `000029`, `/ready`, os quatro serviços e os dois domínios públicos retornaram estado saudável.
- A leitura autenticada de `GET /v1/plans/current` foi executada com sucesso e a interface carregou o plano, a aba de novidades e o detalhamento da decisão.
- O arquivo `melhorias.md` foi removido; `planejamento.md` é a única fonte de roadmap ativa.
- O Cloudflare Tunnel dedicado expõe somente frontend e API; o PostgreSQL não possui hostname, rota pública ou porta publicada.

## Estado do checkout local

- A produção está na versão `0.31.0`, com as migrações `000001` a `000029` aplicadas. O checkout local adiciona o questionário adaptativo, contexto seguro e indicadores observacionais; a migração `000030` foi validada no PostgreSQL local e permanece pendente somente de publicação autorizada.
- A sequência recente inclui `49f1dbd` (catálogo de evidências), `4683999` (piloto de estrada), `5fbc668` (adaptação de recuperação), `c768ef7` (nota de atualização), `810183c` (comparação observacional por períodos), `64e554d` (avaliação shadow do `rules-v2`), `2359c3f` (matriz de validação ampliada), `de23add` (avaliação shadow pós-treino), `b6ea8bd` (observação transacional e inicialização local), `9034287` (matriz comparativa), `61d7939` (pin do digest do Tunnel), `1358ac1` (status da versão `0.12.0`), `53cbadc` (acesso ao perfil no mobile), `66f70ed` (decisão do taper pré-prova), `0eb34d6` (implementação local do taper), `01875c9` (piloto local de VO₂max de estrada), `1eab2c8` (piloto local de intervalos curtos), `3b3639a` (exclusão de modalidades fora do produto), `9aff39f` (sincronização documental), `6fdbe45` (estado do catálogo) e o deploy autorizado da versão `0.16.0`.
- As migrações `000015` e `000016`, o catálogo inicial, o protocolo `road_moderate_intervals` e o piloto `xco_aerobic_intervals` foram aplicados e publicados na produção após revisão, backup, validação e autorização explícita.
- Protocolos adicionais continuam exigindo revisão própria de elegibilidade, segurança, evidência e atualização das notas de versão do produto.

### Coerência da precedência dos gates no shadow — versão local (validado; sem publicação)

- A revisão confirmou que sinais protetivos têm precedência sobre conclusão parcial e que uma resposta fácil sem cobertura completa de tolerância permanece `not_evaluated`/`defer_progression`.
- O `decision_audit` agora registra `load_tolerance_gate` também quando a tolerância está incompleta e a candidata é adiada, não apenas quando existe um sinal protetivo. Isso torna explícita a diferença entre proteção observada e evidência ainda insuficiente.
- Foram adicionadas regressões para a precedência entre dor e conclusão parcial e para a preservação do gate de tolerância incompleto. `go test -count=1 ./...`, `go vet ./...` e `git diff --check` passaram.
- A alteração permanece observacional: `rules-v1` continua prescritivo, `progression_eligible`, `applied` e `used_for_prescription` continuam falsos. Não houve migração, mudança visual, atualização de `APP_VERSION`, deploy ou alteração de infraestrutura.

### Consistência transacional da auditoria shadow — versão local (validado; sem publicação)

- A finalização da transação agora reconstrói o `decision_audit` depois de anexar `planned_vs_actual` e depois de registrar uma eventual falha recuperável da consulta histórica.
- Lacunas da comparação planejado versus realizado passam a aparecer também na auditoria consolidada. Quando o histórico não pode ser consultado, `history_query_failed` não é apresentado como dado utilizado e o audit registra `history_query_gate`.
- Foram adicionados testes para a atualização da auditoria após anexar observações e para a não utilização de histórico quando a consulta falha. A suíte Go completa, `go vet`, as fixtures PostgreSQL somente leitura e `git diff --check` passaram.
- A alteração permanece observacional: não muda `rules-v1`, prescrição, interface, `APP_VERSION`, migrações, infraestrutura ou produção.

### Cobertura HTTP do shadow — versão local (validado; sem publicação)

- O endpoint autenticado `GET /v1/plans/current` foi coberto com um plano sintético e confirmou a serialização de `adaptation_shadow`, `planned_vs_actual` e `decision_audit` até o cliente.
- O teste também confirmou que `history_query_gate` e as lacunas observacionais permanecem visíveis e que `used_for_prescription` continua falso.
- Não foi necessário usar conta real, navegador, migração ou alteração de versão. A cobertura é de contrato HTTP e não transforma o shadow em motor prescritivo.

### Matriz final de não autoridade do shadow — versão local (validado; sem publicação)

- A matriz cobre feedback inválido, sinal protetivo, conclusão parcial, evidência incompleta, baixa aderência, histórico inconsistente, integridade da sessão atual e candidata com evidência completa.
- Em todos os cenários, `progression_eligible`, `applied` e `used_for_prescription` permanecem falsos, inclusive no `decision_audit`, e `prescription_isolation_gate` permanece avaliado.
- A suíte Go completa e `go vet` passaram. Esta etapa encerra a revisão técnica local do shadow; progressão em ciclo fechado continua condicionada a dados reais, calibração e nova revisão.

### Deploy da versão `0.20.0` — 13 de setembro de 2026

- O commit `84b653b` foi atualizado na VPS por fast-forward após confirmação do checkout remoto. O backup preventivo `cadencia-20260913T161932Z.dump` foi criado e verificado; as migrações `000020_completion_context` e `000021_post_workout_context` foram aplicadas em ordem.
- As imagens da API e do frontend foram reconstruídas e os dois serviços foram recriados. PostgreSQL e o Cloudflare Tunnel permaneceram ativos; nenhum serviço opcional foi iniciado.
- A API interna respondeu `{"service":"cadencia-api","status":"ok"}` em `/health` e `{"status":"ready"}` em `/ready`. `https://cadencia-api.devsaulo.com.br/health`, `/ready`, `https://cadencia.devsaulo.com.br/` e `/novidades` retornaram HTTP 200. O endpoint autenticado `/v1/plans/current` retornou HTTP 401 sem sessão, conforme esperado.
- O HTML público contém a versão `0.20.0`. A tela de novidades e o contexto de conclusão/feedback pós-treino passam a estar publicados; o `rules-v1` continua sendo o motor prescritivo e o shadow permanece sem autoridade sobre a prescrição.

### Correção de layout e novidades — versão local `0.20.0`

- A barra lateral passou a preservar o tamanho original dos menus e a usar rolagem própria somente quando a altura disponível não comporta todo o conteúdo. O bloco inferior não é comprimido e o rodapé fixo não cobre mais o acesso ao perfil.
- A versão principal do produto foi corrigida para `0.20.0`, mantendo a nota correspondente em `frontend/lib/release.ts`. A validação visual local confirmou o aviso `NOVIDADES · V0.20.0`, a mensagem `Plano explicável` no menu lateral e o rodapé com a versão atual.
- A validação do treino concluído confirmou `planned_vs_actual-v1` com 3 minutos realizados de 31 planejados, `data_issues: []`, `status: observed`, `progression_eligible: false` e `used_for_prescription: false`. O RPE realizado acima do alvo acionou somente a proteção observacional de esforço alto.
- `npm run build` e `git diff --check` passaram. A implementação foi registrada no commit `2828049`; produção permanece em `0.16.0`, sem deploy, migração ou alteração de infraestrutura.

### Filtro de integridade no histórico observado — versão local (validado; sem publicação)

- O `data-integrity-v1` já marcava sessões concluídas como elegíveis ou inelegíveis, mas as consultas de histórico ainda podiam contar uma sessão explicitamente inelegível em sinais agregados de carga, dor, fadiga e recência.
- `backend/internal/repository/planning.go` e `backend/internal/repository/evolution.go` agora filtram sessões com `workouts.explanation.data_integrity.eligible_for_history = false` no resumo observado usado na geração do plano, nas janelas de 7/28/42 dias, nos seis períodos semanais do shadow e nos agregados da tela de Evolução. A aderência planejada continua separada e o registro original não é apagado.
- Treinos legados sem o bloco de integridade permanecem legíveis para não quebrar históricos anteriores; apenas registros que carregam explicitamente `eligible_for_history: false` são excluídos dessas métricas observacionais.
- A fixture de `scripts/test-training-history-query.ps1` passou a incluir uma sessão inelegível e confirmou que ela não contamina minutos, carga session-RPE, dor ou fadiga nas janelas e períodos.
- `go test -count=1 ./...`, `go vet ./...`, `npm run build`, a consulta PostgreSQL somente leitura e `git diff --check` passaram. Não houve migração, mudança visual, atualização de `APP_VERSION`, deploy ou alteração de infraestrutura.

### Consistência temporal do resumo observado — versão local (validado; sem publicação)

- O resumo observado de 28 dias usado para montar o contexto de prontidão agora aceita somente sessões concluídas até `now()`. Isso alinha essa consulta às janelas cumulativas, aos períodos não sobrepostos e à qualidade temporal do histórico.
- Sessões futuras não entram em minutos, RPE médio, fadiga, dor ou cobertura de dados. Sessões explicitamente inelegíveis por `data-integrity-v1` continuam fora; a aderência planejada permanece separada.
- A regressão SQL sintética cobre sessão elegível, sessão inelegível, sessão futura, sessão fora da janela, sessão cancelada e conta sem registros. A validação passou em transação somente leitura; `go test -count=1 ./...`, `go vet ./...` e `git diff --check` também passaram.
- Esta é uma correção somente de backend e teste: não muda interface, `rules-v1`, migração, release, infraestrutura ou produção. O próximo passo é revisar os agregados observacionais restantes com a mesma referência temporal, sem transferir autoridade ao `rules-v2`.

### Consistência temporal da tela de Evolução — versão local (validado; sem publicação)

- Os agregados da Evolução agora descartam sessões concluídas ou canceladas no futuro do relógio do banco, preservando somente eventos já ocorridos.
- O resumo total, o agrupamento semanal, a lista de sessões recentes e os check-ins exibidos na Evolução usam a mesma barreira temporal. O filtro de integridade continua aplicado às sessões realizadas e a conta de outro atleta permanece isolada.
- `scripts/test-evolution-queries.ps1` executa as quatro consultas reais com CTEs sintéticas em transação somente leitura e confirmou a exclusão de sessões futuras, cancelamento futuro, check-in futuro e dados de outro atleta.
- Esta fatia é somente de backend e testes: não altera interface, `rules-v1`, migração, release, infraestrutura ou produção. O próximo passo é revisar os demais agregados observacionais e manter qualquer evolução prescritiva no shadow até haver evidência suficiente.

### Integração do gate de tolerância no shadow — versão local (validado; sem publicação)

- O `rules-v2-adaptation-v1` agora trata o resultado protetivo de `load-tolerance-v1` como gate efetivo da própria avaliação shadow. Um esforço acima do alvo no período mais recente não pode coexistir com uma candidata de progressão baseada apenas na sessão atual.
- O motivo `recent_above_target_rpe` passa a aparecer na decisão shadow e `load_tolerance_gate` fica listado entre as regras avaliadas. `rules-v1`, o trigger SQL e a prescrição ativa permanecem inalterados.
- A matriz comparativa recebeu o cenário de resposta fácil atual com esforço acima do alvo recente e confirmou `protective_signal`/`prefer_recovery`, mantendo `progression_eligible`, `applied` e `used_for_prescription` como `false`.
- Os testes direcionados de planejamento, repositório e HTTP, `go vet` e `git diff --check` passaram. Não houve mudança visual, migração, release, infraestrutura ou deploy; não é necessário teste no navegador nesta fatia.

### Coerência da conclusão parcial no shadow — versão local (validado; sem publicação)

- O `rules-v2-adaptation-v1` agora avalia explicitamente `completion_status` antes de considerar uma candidata de progressão. Após priorizar sinais protetivos reais, uma sessão `partial` fica como `not_evaluated`/`defer_progression` com o motivo `partial_completion`.
- A decisão não transforma falta de tempo, interrupção ou outra conclusão parcial em sinal fisiológico de recuperação; apenas impede que a sessão seja confundida com tolerância ao treino completo. O histórico e o `load-tolerance-v1` continuam exigindo feedback completo para evidência de carga tolerada.
- A auditoria passou a registrar `load_tolerance_gate` quando a tolerância observada está protetiva. A matriz automatizada cobre a conclusão parcial e a auditoria do gate; `go test -count=1 ./...`, `go vet` e `git diff --check` passaram.
- Esta é uma correção somente de shadow, testes e documentação: `rules-v1`, trigger, banco, interface, release, migrações, infraestrutura e produção permanecem inalterados.

### Fidelidade da proveniência na auditoria shadow — versão local (validado; sem publicação)

- `adaptation-audit-v1` agora consolida as lacunas dos blocos `post_workout_context`, `load_tolerance` e `planned_vs_actual` quando presentes, sem substituir os dados específicos de cada bloco.
- Os campos básicos entram em `data_used` somente quando o feedback e o RPE passaram pela validação mínima. Feedback inválido não aparece como se tivesse sustentado a avaliação; períodos históricos avaliados ficam identificados como `training_history_periods`.
- Foram adicionados testes para lacunas aninhadas, campos inválidos e registro do gate de tolerância. `go test -count=1 ./...`, `go vet` e `git diff --check` passaram.
- A alteração é somente observacional e documental: `rules-v1`, trigger, banco, interface, release, migrações, infraestrutura e produção permanecem inalterados.

### Coerência da evidência de recuperação por período — versão local (validado; sem publicação)

- A candidata de progressão do `rules-v2-adaptation-v1` agora só avança na avaliação quando `load-tolerance-v1` está em `observation_only`. Isso exige carga, feedback completo e um check-in de recuperação completo em cada um dos dois períodos não sobrepostos.
- Quando a recuperação existe apenas no período anterior, a lacuna do período recente é propagada e o resultado fica `not_evaluated`/`defer_progression`, sem tratar a ausência de registro como diagnóstico ou baixa tolerância.
- Foi adicionada regressão para impedir progressão com recuperação incompleta em um dos períodos. `go test -count=1 ./...`, `go vet` e `git diff --check` passaram.
- Esta alteração é somente de shadow e testes: `rules-v1`, trigger, banco, interface, release, migrações, infraestrutura e produção permanecem inalterados.

### Gate de aderência na progressão shadow — versão local (validado; sem publicação)

- A candidata de progressão agora consulta os dois períodos de evidência e bloqueia quando há sessão prevista perdida (`missed_sessions`) ou treino em andamento vencido (`overdue_in_progress_sessions`). O resultado fica `not_evaluated`/`defer_progression` com `low_adherence`.
- Cancelamentos explícitos continuam separados dos treinos perdidos e não são interpretados isoladamente como baixa aderência. A regra não cria reagendamento, compensação ou alteração no plano ativo.
- Foi adicionada regressão para uma sessão perdida no período recente. `go test -count=1 ./...`, `go vet` e `git diff --check` passaram.
- Esta alteração é somente de shadow e testes: `rules-v1`, trigger, banco, interface, release, migrações, infraestrutura e produção permanecem inalterados.

### Décima quarta fatia de melhorias — taper pré-prova orientado por evento (local)

- O plano local passa a registrar `prescription_snapshot.event_taper` com `taper-v1`. A redução só é prescritiva para evento futuro entre 7 e 21 dias, atleta avançado, avaliação submáxima apta, pelo menos 8 semanas e 3 pedais semanais, sem limitação, dor ou necessidade recente de recuperação.
- Sessões anteriores ao evento e dentro de 14 dias recebem redução de 50% na duração, com mínimo de 20 minutos; frequência, RPE-alvo, semana de recuperação e proteções do `rules-v1` são preservados. O dia do evento fica fora da redução.
- A migração `000017` registra três fontes do taper e foi aplicada somente no PostgreSQL de desenvolvimento existente. O trabalho está no commit local `0eb34d6`; a produção continua aplicada até `000016`, sem deploy, backup de produção, alteração de infraestrutura ou publicação da release `v0.13.0`.
- A versão local foi atualizada para `0.13.0` e a tela de novidades informa a mudança. `go test -count=1 ./...`, `go vet ./...`, `npm run build` e `git diff --check` passaram. A validação positiva no navegador confirmou evento em `20/09/2026`, oito dias de distância, `status: eligible`, `applied: true`, `used_for_prescription: true`, multiplicador `0.5` e cinco sessões com `event_taper_applied`; os testes automatizados também confirmaram a não aplicação fora da janela e diante de dor/recuperação. O rascunho permaneceu sem aceite; produção continua sem a migração `000017` e sem publicação da versão `0.13.0`.

### Décima quinta fatia de melhorias — piloto de intervalos VO₂max de estrada (local)

- O catálogo local passa a reconhecer a preferência explícita `vo2max` e pode selecionar `road_vo2_intervals`, apresentado como **Intervalos VO₂max de estrada**.
- A escolha exige disciplina `road`, atleta avançado, objetivo `performance` ou `event`, avaliação submáxima apta, pelo menos oito semanas de treino recente, três pedais semanais, 60 minutos disponíveis, semana de construção e ausência de proteções. Perfis intermediários, outras modalidades, histórico insuficiente e preferência não informada permanecem fora do piloto.
- A sessão usa quatro blocos de 4 minutos em RPE 8 com quatro minutos leves, sem sprint máximo, cadência baixa obrigatória, meta fixa de potência ou frequência cardíaca tratada como equivalente a VO₂max. A dor, limitação, recuperação insuficiente e o taper continuam vencendo a seleção.
- A migração `000018_road_vo2_catalog_evidence` foi aplicada somente no PostgreSQL local e registra `road-vo2-intervention-2024` e `road-vo2-response-2024`. O `rules-v1` continua sendo o único motor prescritivo; produção permanece em `0.12.0`/`000016`, sem deploy ou mudança de infraestrutura.
- A versão local passou para `0.14.0` e a tela de novidades foi atualizada. `go test -count=1 ./...`, `go vet ./...`, `npm run build` e `git diff --check` passaram. A validação ponta a ponta confirmou no navegador o salvamento da preferência `VO₂max`, a aceitação do contexto pela API local após reiniciar o binário atual e a geração de um plano com **Intervalos VO₂max de estrada**. Produção permanece em `0.12.0`/`000016`, sem deploy ou mudança de infraestrutura.

### Piloto local de catálogo — intervalos curtos autorregulados (validado localmente; fora da produção)

- A pesquisa de 12 de setembro de 2026 recomenda avaliar intervalos curtos autorregulados para estrada ou indoor, sem sprint máximo. O estudo de Hesketh et al. (2025) comparou `4–8 × 30 s` com 120 segundos de recuperação e `6–10 × 1 min` com 1 minuto de recuperação em adultos anteriormente inativos; ambos melhoraram o VO₂peak, mas essa população não representa automaticamente os atletas do Cadência.
- O estudo de Rønnestad et al. (2020) em ciclistas de elite informa que intervalos de 30 segundos podem produzir adaptações favoráveis, porém a amostra, o nível e o esforço repetido limitam a transferência. Uma meta-análise de 2025 reforça a heterogeneidade entre HIIT, SIT e repeated-sprint training.
- O protocolo local `short_self_regulated_intervals` foi implementado com a preferência explícita `short_intervals`, seis repetições de 1 minuto em RPE 7,5 e recuperação leve autorregulada. A migração `000019_short_intervals_evidence` registra as duas fontes do piloto e a versão local passou para `0.15.0`, com nota na tela de novidades. A suíte Go, o `go vet`, o build e a aplicação/verificação da migração local passaram. A validação manual confirmou no navegador o salvamento da preferência, a atualização do plano e a apresentação de **Intervalos curtos autorregulados**. Produção permanece em `0.12.0`/`000016`, sem deploy ou mudança de infraestrutura.

### Próxima pesquisa do catálogo — resistência específica na bicicleta (candidato bloqueado)

- O estudo randomizado de Barranco-Gil et al. (2025) avaliou 10 semanas de resistência fora e na bicicleta em 37 ciclistas bem treinados. O protocolo na bicicleta usou resistência muito alta, cadência muito baixa e carga calibrada por força dinâmica máxima; os grupos melhoraram força e potência, mas não VO₂max. [Fonte no PubMed](https://pubmed.ncbi.nlm.nih.gov/39231694/)
- A evidência sustenta investigar o estímulo, mas não permite uma conversão honesta para RPE nem uma liberação para perfis gerais: o Cadência não mede força dinâmica máxima nem calibra a carga estudada. O candidato exige medição, elegibilidade avançada e travas próprias antes de qualquer código prescritivo.
- Nenhuma preferência, protocolo, migração ou nota de versão foi criada nesta etapa. Gravel/XCM continuam contextos de endurance sem protocolo próprio; sprint/pista/BMX e downhill/enduro estão fora do produto.

### Decisão de escopo de modalidades — 12 de setembro de 2026

- Sprint/pista/BMX e downhill/enduro não fazem parte do Cadência. As opções foram removidas do perfil e a API passou a rejeitar novos valores `track_sprint` e `dh_enduro`.
- Registros legados dessas modalidades, se existirem, são tratados como disciplina não informada no carregamento do perfil e não liberam protocolos. Não houve migração ou alteração de produção.
- A versão local passou para `0.16.0` e a tela de novidades comunica a decisão. A produção permanece em `0.12.0`/`000016`.

## Arquitetura efetiva

### Primeira fatia de melhorias — prontidão observacional (local)

- Novos rascunhos registram `prescription_snapshot.readiness_assessment`, com versão `readiness-v1`, modo `observation`, horário UTC, estado, motivos, dados ausentes, cobertura e fatores ainda não avaliados.
- Classificação independente de experiência e avaliação submáxima: `insufficient_data`, `caution`, `recovery_needed` ou `stable`. `stable` significa apenas ausência de alertas nos agregados disponíveis, não aptidão atual ou liberação de carga.
- O repositório passou a contar campos válidos de sessões e fadiga dos check-ins. Médias que ignoram campos nulos não são tratadas como cobertura completa.
- `progression_eligible` permanece `false` nesta leitura. O campo não é uma nova trava do motor: os treinos e as proteções existentes continuam sendo prescritos pelo `rules-v1`, sem alteração de duração, RPE ou blocos nesta entrega.
- Baixa consistência e prontidão para progressão não são inferidas da quantidade de registros. Ainda faltam critérios validados de tolerância, destreinamento, tendência e progressão; a leitura de períodos agora existe apenas para comparação observacional.
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

### Terceira fatia de melhorias — qualidade temporal e sinais protetivos (commit local)

- Novos rascunhos usam `training-history-v2`. Planos já persistidos com `training-history-v1` continuam válidos e não são recalculados.
- `temporal_quality` registra a última sessão concluída, a última sessão com carga session-RPE válida, o último check-in e os dias desde cada registro. Sessões concluídas no futuro e check-ins com data futura são excluídos e sinalizados em `data_issues`.
- Cada janela passa a separar registros de feedback, feedbacks completos, dor, fadiga de 4 a 5 e RPE real pelo menos dois pontos acima do planejado. Também conta check-ins totais/completos, check-ins com ao menos um sinal protetivo e aqueles que já atendem à regra existente de necessidade de recuperação.
- Os limiares são somente os já usados pelo `rules-v1`; não foi criado um novo corte fisiológico. Dor continua contada mesmo se outro campo do feedback estiver ausente, e a falta parcial entra em `missing_data`.
- `not_evaluated` mantém explícitos tolerância à carga, destreinamento, mudança de condicionamento, atividades fora do Cadência, fuso do atleta e progressão. A ausência de atividade no app significa apenas lacuna de registros, não interrupção comprovada do treino.
- `used_for_prescription` continua `false`; o `rules-v1`, a prontidão, os treinos e as adaptações não foram alterados. Não há ACWR, mudança visual, migração ou infraestrutura; `APP_VERSION` e `UPDATE_NOTES` permanecem em `0.7.0`.
- Validação concluída: `go test -count=1 ./...`, `go vet ./...`, build do frontend, carregamento do OpenAPI e a consulta PostgreSQL com fixtures sintéticas em transação somente leitura passaram. A consulta cobre janelas, cobertura incompleta, sinais protetivos, registros futuros e isolamento de outro atleta. O lint geral mantém erros preexistentes em telas/componentes fora desta fatia; `frontend/lib/planning.ts` não acrescentou erro. Na conferência manual, um novo rascunho retornou `training-history-v2`, `data_issues: []` e `used_for_prescription: false`; os cinco feedbacks apareceram completos, o check-in entrou como sinal protetivo/necessidade de recuperação e a carga permaneceu indisponível porque as cinco sessões realizadas tinham duração zero.

Arquivos desta fatia: `backend/internal/planning/history.go`, `backend/internal/planning/history_test.go`, `backend/internal/repository/planning.go`, `frontend/lib/planning.ts`, `api/openapi.yaml`, `scripts/test-training-history-query.ps1`, `README.md`, `docs/README.md`, `docs/project-status.md`, `docs/architecture-decisions.md`, `docs/training-adaptation-rules.md` e `planejamento.md`. Não foram alterados `melhorias.md`, migrações, infraestrutura ou notas de versão.

### Quarta fatia de melhorias — comparação entre períodos (commit local)

- Novos rascunhos passam a usar `training-history-v3`. Snapshots já persistidos em `training-history-v1` ou `training-history-v2` continuam legíveis e não são recalculados.
- `period_comparison` registra seis períodos semanais não sobrepostos: `last_7d`, `days_8_14`, `days_15_21`, `days_22_28`, `days_29_35` e `days_36_42`. Cada período mantém as mesmas medições brutas de aderência, sessões realizadas, carga session-RPE, feedback, sinais protetivos e check-ins, sem agregar semanas diferentes.
- Sessões realizadas e carga usam intervalos contínuos de `completed_at` baseados no relógio do PostgreSQL; aderência e recuperação usam blocos de datas baseados em `CURRENT_DATE`. Essa diferença temporal continua explícita e não é tratada como erro por si só.
- O contrato informa `period-comparison-v1`, `mode: observation`, lacunas, inconsistências e `used_for_prescription: false`. Não há cálculo de tendência, razão aguda:crônica, limiar de carga, progressão, regressão ou conclusão de destreinamento. `period_trend_for_prescription` permanece em `not_evaluated`.
- Nenhuma prescrição, adaptação, migração, infraestrutura ou tela foi alterada. Como não há mudança visível no produto, `APP_VERSION` e `UPDATE_NOTES` permanecem em `0.7.0`; a próxima funcionalidade visível deverá atualizar a tela de novidades.
- Validação concluída: `go test -count=1 ./internal/planning ./internal/repository ./internal/httpapi`, `go vet ./...`, build do frontend, validação estrutural do OpenAPI e `scripts/test-training-history-query.ps1` passaram. O script confirmou os seis períodos com fixtures sintéticas em transação somente leitura; nenhum dado real foi lido ou alterado.
- A conferência manual ponta a ponta foi concluída em um novo rascunho: o `GET /v1/plans/current` retornou a versão `training-history-v3`, os seis períodos e a preservação do modo observacional, sem alteração indevida dos treinos.

Arquivos desta fatia: `backend/internal/planning/history.go`, `backend/internal/planning/history_test.go`, `backend/internal/planning/service.go`, `backend/internal/repository/planning.go`, `frontend/lib/planning.ts`, `api/openapi.yaml`, `scripts/test-training-history-query.ps1`, `README.md`, `docs/README.md`, `docs/project-status.md`, `docs/architecture-decisions.md`, `docs/training-adaptation-rules.md` e `planejamento.md`. Não foram alterados `melhorias.md`, migrações, infraestrutura ou notas de versão.

### Quinta fatia de melhorias — avaliação shadow do `rules-v2` (commit local; validação ampliada)

- Novos rascunhos registram `prescription_snapshot.rules_v2_shadow` com versão `rules-v2`, modo `shadow` e escopo `plan_generation_only`. O `engine_version` continua sendo `rules-v1`.
- A avaliação aplica somente gates determinísticos de integridade dos períodos, sinais protetivos e evidência mínima para progressão. Ela pode registrar `protective_signal`, `observation_only` ou `not_evaluated`, além de uma resposta candidata explicitamente não aplicada.
- A resposta protetiva considera limitações ativas, dor, fadiga elevada, necessidade de recuperação e sinais do período mais recente. Para chegar a `observation_only`, exige dois períodos recentes com sessões, carga session-RPE e feedback completos, além de check-in de recuperação completo nos últimos 14 dias. Isso é um critério de validação do shadow, não uma regra de prescrição.
- Dados insuficientes ou inconsistentes permanecem em `missing_data`/`data_issues`. A avaliação não conclui destreinamento, tolerância, mudança de condicionamento, progressão ou resposta fisiológica; `progression_eligible`, `applied` e `used_for_prescription` permanecem `false`.
- Nenhum treino, duração, RPE, protocolo, migração, infraestrutura ou tela foi alterado. Como não há mudança visual, `APP_VERSION` e `UPDATE_NOTES` continuam em `0.7.0`.
- Validação concluída com uma ressalva operacional: os testes específicos e a suíte Go passaram nos pacotes executados; a execução agregada foi bloqueada apenas pelo Controle de Aplicativos do Windows ao abrir o executável temporário de `internal/repository`, mas o mesmo pacote passou quando compilado e executado dentro do workspace. `go vet`, build do frontend, validação do OpenAPI e a consulta PostgreSQL sintética também passaram. A conferência manual via API confirmou `protective_signal`, `prefer_recovery`, `applied: false`, `used_for_prescription: false`, `progression_eligible: false`, `missing_data: []` e `data_issues: []`. A matriz controlada confirmou cenários de evidência completa, sinal protetivo, dados insuficientes, período inconsistente e invariância dos treinos do `rules-v1` com ou sem dados do shadow.

Arquivos desta fatia: `backend/internal/planning/rules_v2.go`, `backend/internal/planning/rules_v2_test.go`, `backend/internal/planning/service.go`, `frontend/lib/planning.ts`, `api/openapi.yaml`, `README.md`, `docs/README.md`, `docs/project-status.md`, `docs/architecture-decisions.md`, `docs/training-adaptation-rules.md` e `planejamento.md`. Não foram alterados `melhorias.md`, migrações, infraestrutura ou notas de versão.

### Sexta fatia de melhorias — adaptação pós-treino em shadow (commit local; sem publicação)

- `backend/internal/planning/rules_v2_adaptation.go` cria `rules-v2-adaptation-v1`, uma avaliação determinística da resposta ao feedback pós-treino. Ela foi desenhada como observação paralela e não altera o trigger ativo do `rules-v1` nem os campos prescritivos.
- Dor, esforço muito alto, fadiga máxima ou sinal protetivo recente produzem a candidata `prefer_recovery`. Uma resposta dentro do esperado produz `maintain_observed`.
- Uma resposta claramente fácil sozinha produz `defer_progression`; somente dois períodos recentes íntegros, com carga session-RPE, feedback completo e recuperação registrada, permitem registrar a candidata `progress_duration_5pct`. Mesmo nesse caso, `progression_eligible`, `applied` e `used_for_prescription` permanecem `false`.
- Dados insuficientes ou inconsistentes ficam em `missing_data`/`data_issues`. Não são avaliados destreinamento, tolerância, mudança fisiológica, atividades fora do Cadência ou efeito da prescrição.
- Testes direcionados passaram nos cinco cenários da adaptação shadow e nos cenários anteriores do `rules-v2`; os testes e a documentação dessa fatia foram versionados no commit `de23add`. Não houve migração, mudança visual, infraestrutura, publicação nem alteração de `APP_VERSION`/`UPDATE_NOTES`.

Arquivos desta fatia: `backend/internal/planning/rules_v2_adaptation.go`, `backend/internal/planning/rules_v2_adaptation_test.go`, `README.md`, `docs/README.md`, `docs/project-status.md`, `docs/architecture-decisions.md`, `docs/training-adaptation-rules.md` e `planejamento.md`. Não foram alteradas migrações, infraestrutura ou notas de versão.

### Sétima fatia de melhorias — observação transacional pós-treino (commit local; sem publicação)

- Ao concluir uma sessão, o repositório consulta os períodos históricos dentro da mesma transação que grava o feedback e chama o avaliador `rules-v2-adaptation-v1`. O resultado é armazenado apenas em `workouts.explanation.adaptation_shadow` no treino concluído, ficando disponível no `GET /v1/plans/current`.
- O `rules-v1` continua sendo a única fonte prescritiva: o trigger existente e seus campos de duração, RPE, estímulo e status não foram substituídos. O shadow permanece com `mode: shadow`, `progression_eligible: false`, `applied: false` e `used_for_prescription: false`.
- A consulta histórica fica protegida por um savepoint. Se ela falhar, o feedback principal pode continuar sendo salvo e o shadow registra `not_evaluated` com `history_query_failed`, sem transformar uma observação em bloqueio do fluxo.
- A tipagem do frontend e o contrato OpenAPI expõem a nova chave, mas não há mudança visual nem funcionalidade visível para o atleta; `APP_VERSION` e `UPDATE_NOTES` continuam em `0.7.0`.
- A inicialização local da API foi ajustada em `scripts/run-api.ps1`: em vez de executar o binário transitório do `go run`, o script compila em `backend/.gotmp`, pasta ignorada pelo Git. Isso contorna o bloqueio do Smart App Control do Windows sem desativar a proteção do sistema e não altera a produção.
- Validação automatizada: `go test -count=1 ./internal/planning ./internal/repository ./internal/httpapi`, `go vet ./internal/planning ./internal/repository ./internal/httpapi`, build do frontend e validação estrutural do OpenAPI passaram. A validação manual confirmou a chave no `GET /v1/plans/current`, com `status: protective_signal`, `candidate_response: prefer_recovery` e todas as barreiras de não aplicação em `false`.

Arquivos desta fatia: `backend/internal/planning/rules_v2_adaptation.go`, `backend/internal/repository/planning.go`, `backend/internal/repository/workout_sessions.go`, `frontend/lib/planning.ts`, `api/openapi.yaml`, `scripts/run-api.ps1`, `docs/README.md`, `docs/project-status.md`, `docs/architecture-decisions.md`, `docs/training-adaptation-rules.md` e `planejamento.md`. Não foram alteradas migrações, infraestrutura ou notas de versão.

### Oitava fatia de melhorias — matriz controlada entre `rules-v1` e shadow (commit local; sem publicação)

- `backend/internal/planning/rules_v2_adaptation_comparison_test.go` compara cenários determinísticos de dor, esforço alto, resposta neutra, necessidade recente de recuperação, resposta fácil sem evidência, resposta fácil com evidência completa e histórico inconsistente.
- A matriz exige que a proteção do `rules-v1` e a candidata protetiva do shadow coincidam. Para progressão, exige que o shadow seja mais restritivo: sem evidência suficiente ou com inconsistência, a candidata fica adiada ou não avaliada; com evidência completa, continua somente como proposta.
- Cada caso confirma `progression_eligible: false`, `applied: false` e `used_for_prescription: false`. O teste é regressivo e não altera geração de plano, trigger SQL, banco ou interface.
- A validação automatizada passou em `go test -count=1 ./internal/planning ./internal/repository ./internal/httpapi` e `go vet ./internal/planning ./internal/repository ./internal/httpapi`. A matriz representa as regras atuais e permanece como regressão antes de qualquer integração prescritiva.

### Nona fatia de melhorias — integridade observacional dos dados (commit local; sem publicação)

- `backend/internal/planning/data_integrity.go` cria `data-integrity-v1` para classificar sessões concluídas como `valid`, `incomplete` ou `inconsistent`, separando ausência de dados de valores incompatíveis.
- O gate verifica duração positiva, RPE realizado, feedback, fadiga, faixas das métricas opcionais e combinações como duração zero com distância ou elevação registradas. O registro original não é rejeitado nem sobrescrito; a leitura é armazenada em `workouts.explanation.data_integrity` para auditoria.
- O shadow do pós-treino consulta esse resultado: uma sessão incompleta ou inconsistente não produz candidata de adaptação. O `rules-v1` e o trigger SQL continuam inalterados nesta fatia, portanto a observação ainda não é uma barreira prescritiva ativa.
- O contrato OpenAPI e a tipagem do frontend expõem a nova leitura. Os testes cobrem sessão coerente, duração zero, distância sem tempo, feedback ausente, métrica fora da faixa e RPE não finito. A validação manual no `GET /v1/plans/current` confirmou uma sessão curta como `incomplete` e uma sessão com duração positiva como `valid`, mantendo `used_for_prescription: false`.

Arquivos desta fatia: `backend/internal/planning/data_integrity.go`, `backend/internal/planning/data_integrity_test.go`, `backend/internal/planning/rules_v2_adaptation.go`, `backend/internal/planning/rules_v2_adaptation_test.go`, `backend/internal/repository/workout_sessions.go`, `frontend/lib/planning.ts`, `api/openapi.yaml`, `README.md`, `docs/README.md`, `docs/project-status.md`, `docs/architecture-decisions.md`, `docs/training-adaptation-rules.md` e `planejamento.md`. Não foram alteradas migrações, infraestrutura ou notas de versão.

### Observação de tolerância à carga — versão local `load-tolerance-v1` (validada)

- A nova avaliação observa os dois períodos semanais mais recentes e exige, em cada um, sessões realizadas com carga session-RPE, feedback completo e ao menos um check-in de recuperação completo.
- Sinais de dor, fadiga alta, recuperação necessária ou esforço atual pelo menos dois pontos acima do alvo produzem resposta protetiva. Com evidência completa e sem esses sinais, o resultado é somente `observation_only`/`maintain_observed`.
- O resultado registra `evidence_periods`, regras, motivos, lacunas e inconsistências. Continua com `progression_eligible: false`, `applied: false` e `used_for_prescription: false`; não calcula ACWR, não infere tolerância fisiológica e não altera o `rules-v1`.
- A leitura foi anexada a `workouts.explanation.adaptation_shadow`, com tipos do frontend e contrato OpenAPI atualizados. Não houve migração, mudança visual, alteração de infraestrutura ou atualização de `APP_VERSION`/`UPDATE_NOTES`.
- Após reiniciar a API local, uma sessão concluída confirmou no `GET /v1/plans/current` a presença de `load_tolerance` com `version: "load-tolerance-v1"`, `progression_eligible: false` e `used_for_prescription: false`. O estado `not_evaluated` foi aceito como esperado para a cobertura histórica da conta de teste. A implementação foi registrada no commit `36cfabb` e ainda não foi publicada em produção.
- A suíte Go, `go vet`, build do frontend e `git diff --check` passaram. O lint geral continua bloqueado pelas pendências anteriores já registradas, sem erro apontado nos arquivos desta fatia.

Arquivos desta fatia: `backend/internal/planning/load_tolerance.go`, `backend/internal/planning/load_tolerance_test.go`, `backend/internal/planning/rules_v2_adaptation.go`, `frontend/lib/planning.ts`, `api/openapi.yaml`, `docs/training-adaptation-rules.md`, `docs/project-status.md`, `docs/architecture-decisions.md` e `planejamento.md`. Não foram alteradas migrações, infraestrutura ou notas de versão.

### Comparação planejado versus realizado — versão local `planned-vs-actual-v1` (validada)

- A conclusão de uma sessão passa a registrar no shadow a duração planejada e realizada, o RPE-alvo e realizado, suas diferenças e os campos de execução observados.
- Métricas opcionais presentes são identificadas em `observed_fields`; ausência de distância, elevação, potência, frequência cardíaca ou cadência média fica em `missing_data`. Sono, estresse, recuperação, extensão da conclusão e motivo de não conclusão permanecem em `not_evaluated` porque ainda não fazem parte deste fluxo.
- O estado `observed` descreve uma comparação mínima válida; `not_evaluated` fica reservado a dados essenciais ausentes ou inválidos. Em ambos os casos, `progression_eligible` e `used_for_prescription` permanecem `false`.
- A implementação não altera duração, RPE, estímulo ou status de sessões e não cria migração, mudança visual ou atualização de `APP_VERSION`/`UPDATE_NOTES`. A validação manual ponta a ponta confirmou no `GET /v1/plans/current` uma sessão com 3 minutos realizados de 35 planejados, `status: "observed"`, `duration_completion_percent: 8.57`, `progression_eligible: false` e `used_for_prescription: false`. As métricas opcionais ausentes e os campos ainda não coletados permaneceram explicitamente classificados.

Arquivos desta fatia: `backend/internal/planning/execution_comparison.go`, `backend/internal/planning/execution_comparison_test.go`, `backend/internal/planning/rules_v2_adaptation.go`, `backend/internal/repository/workout_sessions.go`, `frontend/lib/planning.ts`, `api/openapi.yaml`, `docs/training-adaptation-rules.md` e `docs/project-status.md`. Não houve deploy ou publicação.

### Registro de conclusão parcial — versão local `0.17.0` (validado localmente; fora da produção)

- O encerramento de uma sessão agora exige `completion_status`: `complete` para o treino inteiro ou `partial` para parte da sessão. Quando parcial, `partial_reason` é obrigatório e usa um conjunto controlado: falta de tempo, fadiga/recuperação, dor/desconforto, equipamento/clima/terreno ou outro motivo.
- O contexto aparece no formulário de conclusão e no resumo da sessão; também é exibido no histórico de `/atividades`. Conclusão parcial não é tratada como tolerância ao treino completo e não pode gerar progressão. Dor e fadiga alta continuam podendo acionar as proteções de segurança/recuperação existentes.
- `planned-vs-actual-v1` passa a registrar `completion_status` e `partial_reason` como observação, retirando extensão/motivo de conclusão de `not_evaluated`. A leitura continua com `progression_eligible: false` e `used_for_prescription: false`.
- A migração local `000020_completion_context` adiciona os campos com `complete` como padrão para registros antigos, valida a combinação status/motivo e atualiza o trigger sem alterar o comportamento protetivo do `rules-v1`. Ela foi aplicada no PostgreSQL local e o teste SQL transacional passou.
- A versão local passou para `0.17.0` e a tela de novidades informa a mudança. A validação manual confirmou uma sessão com `Fiz apenas parte` e `Fiquei sem tempo`, o retorno `Parcial · sem tempo` no histórico e a ausência de progressão automática. O contraste dos menus e do placeholder foi corrigido. `go test -count=1 ./...`, `go vet ./...`, `npm run build` e `git diff --check` passaram. O lint geral mantém pendências preexistentes fora desta entrega. A implementação foi registrada em `58e3887`, com os ajustes visuais em `b5ef413` e `d7ce5bb`; a versão `0.17.0` e a migração `000020` permanecem fora da produção.

Arquivos desta fatia: `backend/internal/httpapi/workout_session_handlers.go`, `backend/internal/planning/adaptation.go`, `backend/internal/planning/data_integrity.go`, `backend/internal/planning/execution_comparison.go`, testes de `backend/internal/planning`, `backend/internal/repository/planning.go`, `backend/internal/repository/workout_sessions.go`, `database/migrations/000020_completion_context.*`, `database/tests/000020_completion_context.sql`, `frontend/components/workout-session-actions.tsx`, `frontend/app/atividades/page.tsx`, `frontend/app/globals.css`, `frontend/lib/planning.ts`, `frontend/lib/release.ts`, `api/openapi.yaml` e esta documentação. A implementação está nos commits `58e3887`, `b5ef413` e `d7ce5bb`; não houve deploy ou alteração de infraestrutura.

### Feedback pós-treino com recuperação e confiança — versão local `0.18.0` (validado localmente; fora da produção)

- O formulário de conclusão passa a registrar `recovery_after` (recuperação percebida) e `repeat_confidence` (confiança para repetir), ambos em escala de 1 a 5. Os valores aparecem no resumo da sessão e no histórico de `/atividades`.
- Os campos são opcionais no banco para preservar feedbacks antigos, mas o formulário atual envia a seleção neutra `3 de 5`. A API, o banco e a integridade observacional rejeitam valores fora da faixa.
- `planned-vs-actual-v1` registra os sinais quando presentes em `observed_fields`, sem interpretar tolerância, prontidão ou efeito da prescrição. `rules-v1`, o trigger pós-feedback e a carga das próximas sessões não foram alterados.
- A migração `000021_post_workout_context` foi aplicada e registrada no PostgreSQL local; o teste SQL transacional, `go test -count=1 ./...`, `go vet ./...`, `npm run build` e `git diff --check` passaram. A verificação direcionada do lint não apontou erro no componente de conclusão; a página de atividades mantém avisos anteriores de navegação por `<a>`.
- A versão local passou para `0.18.0` e a tela de novidades informa a mudança. A validação manual confirmou o registro e a exibição dos dois campos no resumo/histórico e no retorno observacional do plano. A implementação foi registrada no commit `34f17b3`; produção permanece em `0.16.0`/`000019`, sem deploy ou alteração de infraestrutura.

Arquivos desta fatia: `backend/internal/httpapi/workout_session_handlers.go`, `backend/internal/planning/service.go`, `backend/internal/planning/data_integrity.go`, `backend/internal/planning/execution_comparison.go`, testes de `backend/internal/planning`, `backend/internal/repository/planning.go`, `backend/internal/repository/workout_sessions.go`, `database/migrations/000021_post_workout_context.*`, `database/tests/000021_post_workout_context.sql`, `frontend/components/workout-session-actions.tsx`, `frontend/app/atividades/page.tsx`, `frontend/lib/planning.ts`, `frontend/lib/release.ts`, `api/openapi.yaml`, `README.md`, `docs/training-adaptation-rules.md`, `docs/architecture-decisions.md`, `infrastructure/cadencia/README.md` e esta documentação. A implementação foi registrada no commit `34f17b3`; não houve deploy ou alteração de infraestrutura.

### Histórico permanente de novidades — versão local `0.19.0` (validado localmente; fora da produção)

- O aviso inicial agora apresenta somente as notas da versão atual, evitando que o primeiro acesso fique excessivamente longo.
- A nova rota autenticada `/novidades` apresenta o histórico completo agrupado por versão, com versões anteriores recolhidas para facilitar a leitura.
- O acesso foi incluído no menu lateral, no menu móvel e no cabeçalho das telas internas. O histórico usa `frontend/lib/release.ts` como fonte única, sem duplicar as descrições.
- O build do frontend passou e a rota `/novidades` foi reconhecida. O lint específico dos arquivos novos e a formatação dos arquivos da interface passaram; o lint geral mantém pendências antigas de navegação por `<a>` no dashboard.
- A validação visual foi concluída pelo proprietário. Produção permanece em `0.16.0`/`000019`, sem deploy ou alteração de infraestrutura.

Arquivos desta fatia: `frontend/app/novidades/page.tsx`, `frontend/app/globals.css`, `frontend/app/page.tsx`, `frontend/components/account-actions.tsx`, `frontend/components/update-notice.tsx`, `frontend/lib/release.ts`, `README.md`, `docs/architecture-decisions.md` e esta documentação. A implementação foi registrada no commit `8798ea3`; ainda não foi publicada.

### Contexto pós-treino integrado ao shadow — versão local (sem mudança visual)

- `post-workout-context-v1` classifica a cobertura de `recovery_after` e `repeat_confidence` dentro de `workouts.explanation.adaptation_shadow`. O bloco distingue os dois sinais completos, contexto parcial, ausência de registro e valores fora da faixa, mantendo `observed_fields`, `missing_data`, `data_issues` e motivos explícitos.
- O bloco é estritamente observacional: `progression_eligible` e `used_for_prescription` permanecem `false`, `rules-v1` continua sendo o único motor ativo e a resposta `maintain_observed` significa apenas que os dados foram registrados, não que uma prescrição foi validada.
- Foram adicionados testes para contexto completo, parcial, inválido e para o isolamento da avaliação dentro do shadow. `go test -count=1 ./...`, `go vet ./...`, `npm run build` e `git diff --check` passaram.
- A validação manual via API local confirmou os dois valores no bloco `post_workout_context`, com `status: "observed"`, e preservou `progression_eligible: false` e `used_for_prescription: false`. A implementação foi registrada no commit `46a900f`; não houve migração, mudança visual, atualização de versão, alteração de infraestrutura ou deploy. Produção permanece em `0.16.0`/`000019`.

Arquivos desta fatia: `backend/internal/planning/post_workout_context.go`, `backend/internal/planning/post_workout_context_test.go`, `backend/internal/planning/rules_v2_adaptation.go`, `frontend/lib/planning.ts`, `api/openapi.yaml` e esta documentação.

### Auditoria da decisão do shadow — versão local (sem mudança visual)

- `adaptation-audit-v1` registra, junto da avaliação `rules-v2-adaptation-v1`, os dados considerados, as lacunas, as restrições aplicadas, os caminhos de adaptação não selecionados e as condições informativas para uma futura revisão.
- O campo `confidence` permanece `not_calibrated`; nenhum nível de confiança é inventado. O bloco também mantém `used_for_prescription: false`, e as condições registradas não funcionam como gatilhos automáticos.
- Foram adicionados testes para proveniência, lacunas de evidência e serialização JSON estável. `go test -count=1 ./...`, `go vet ./...`, `npm run build`, `oxlint` da tipagem alterada e `git diff --check` passaram.
- Não houve migração, mudança visual, atualização de versão, alteração de infraestrutura, commit ou deploy nesta fatia. Produção permanece em `0.16.0`/`000019`.

Arquivos desta fatia: `backend/internal/planning/adaptation_audit.go`, `backend/internal/planning/adaptation_audit_test.go`, `backend/internal/planning/rules_v2_adaptation.go`, `frontend/lib/planning.ts`, `api/openapi.yaml` e esta documentação.

### Décima fatia de melhorias — piloto publicado de intervalos aeróbicos XCO

- O catálogo passa a selecionar `xco_aerobic_intervals` somente quando a disciplina `mtb_xco` é informada explicitamente, o atleta é avançado, o objetivo é performance ou prova, a avaliação submáxima está apta, há pelo menos 75 minutos disponíveis, a semana não é de recuperação e não há proteção ativa por limitação, dor ou sinais recentes de recuperação insuficiente.
- A sessão usa cinco blocos de 4 minutos com 4 minutos leves, alvo RPE 7 e uma única sessão de qualidade no ciclo. É uma adaptação conservadora do HIT estudado em mountain bikers treinados; não inclui sprint máximo, técnica de trilha, descida, salto ou meta rígida de potência.
- A migração `000016` registra `xco-hit-2016`, um ensaio randomizado de 2016. A revisão sistemática contemporânea de XCO de 2026 orienta a especificidade intermitente, mas ressalta a escassez de avaliações diretas de desempenho; por isso, o protocolo permanece um piloto publicado apenas para o perfil elegível.
- Gravel continua apenas como contexto de endurance, sem protocolo próprio baseado em um único estudo de campo. Sprint/pista/BMX e downhill/enduro estão fora do produto.
- Como a seleção é visível ao atleta, `frontend/lib/release.ts` foi atualizado para a versão `0.8.0` e a tela de novidades passou a explicar o piloto XCO. A migração `000016`, a nova regra e a nota foram validadas localmente e publicadas após backup e autorização.

Arquivos desta fatia: `backend/internal/planning/protocols.go`, `backend/internal/planning/service.go`, `backend/internal/planning/service_test.go`, `database/migrations/000016_xco_catalog_evidence.up.sql`, `database/migrations/000016_xco_catalog_evidence.down.sql`, `docs/cycling-evidence-catalog.md`, `docs/training-adaptation-rules.md`, `docs/architecture-decisions.md`, `frontend/lib/release.ts`, `README.md`, `docs/README.md` e `planejamento.md`. Não foram alteradas infraestrutura ou regras prescritivas do shadow.

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
- `database/migrations/`: migrações PostgreSQL até `000029`; as `000013` e `000014` sustentam feedback e resumo semanal, a `000015` registra as fontes do catálogo inicial, a `000016` registra a fonte do piloto XCO, a `000017` registra as fontes do taper pré-prova, a `000018` registra as fontes do piloto VO₂max de estrada, a `000019` registra as fontes do piloto de intervalos curtos, a `000020` registra o contexto de conclusão parcial, a `000021` registra o contexto adicional pós-treino, a `000022` registra o contexto de limitações, a `000023` registra o feedback estruturado, as `000024`/`000025` registram equipamento e sinais de segurança, a `000026` amplia os metadados científicos, a `000027` protege a adaptação contra dados inválidos e sinais protetivos recentes, a `000028` registra evidências de recuperação pós-prova e a `000029` registra cadência média observacional. Em produção, todas estão aplicadas até `000029`.
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
- Contexto opcional de ciclismo: horas, pedais e distância semanal recente, semanas de regularidade, situação atual do treino, maior distância e pedal, preferências de sessão, equipamento, terreno, sensores, FTP e meta de prova.
- Resumo observado dos últimos 28 dias: sessões concluídas, minutos realizados, RPE e fadiga médios, dor relatada e check-ins de recuperação.
- Motor `rules-v1` com ciclos de quatro semanas, progressão, recuperação e datas calculadas para a semana corrente.
- Geração de rascunho, revisão, ativação e geração do próximo ciclo sem apagar o histórico.

### Treinos, feedback e evolução

- Sessões planejadas, iniciadas, concluídas e canceladas.
- Feedback com conclusão completa/parcial, motivo controlado quando parcial, RPE, dificuldade, fadiga, dor e observações.
- Métricas opcionais: distância, elevação, frequência cardíaca média e potência média.
- Adaptação conservadora após feedback e check-in diário de sono, estresse e fadiga.
- Avaliação inicial submáxima, sem teste máximo ou diagnóstico.
- O histórico observado agora participa da geração: sinais recentes de dor, fadiga ou recuperação insuficiente protegem as sessões futuras de forma conservadora e ficam no snapshot do plano.
- Sessões específicas para perfis adequados: cadência, subidas, sweet spot, ritmo de prova e intervalos controlados. O piloto `road_moderate_intervals` e suas referências estão publicados, com a elegibilidade e os limites documentados.
- Histórico em `/atividades` e agregações observadas em `/evolucao`.
- A aba `/evolucao` também compara as últimas sessões concluídas com a prescrição original (tempo, RPE e métricas registradas), sem transformar a diferença em ajuste automático.
- O plano exibe o resumo do contexto observado usado na geração do ciclo, com sessões, minutos, RPE, check-ins e alertas conservadores de recuperação quando aplicável.
- Indicadores de consistência, carga semanal, prontidão e explicabilidade.
- A adaptação shadow pós-treino também registra `load-tolerance-v1`, uma leitura observacional dos dois períodos recentes com session-RPE, feedback e recuperação completos. Ela permanece sem autoridade prescritiva e protege diante de dor, fadiga alta, recuperação necessária ou esforço acima do alvo.
- Aba `/feedback` para o atleta registrar uma experiência, problema ou sugestão com nota de 1 a 5. O relato fica vinculado à conta no PostgreSQL, sem coleta adicional de contato nesta primeira versão.
- Resumo semanal de feedback implementado como comando separado (`cadencia-feedback-digest`). Ele busca até 50 relatos ainda não enviados, envia um único e-mail pelo Resend ao endereço `FEEDBACK_DIGEST_TO` e marca os registros somente depois de um envio bem-sucedido. O serviço não é iniciado junto da API e fica desativado quando o destinatário não está configurado.
- Contrato inicial da IA explicativa no backend, com Ollama opcional, limites de recurso e fallback determinístico para as regras. O cliente do Worker Cloudflare está preparado como provedor remoto, e a rota protegida `/cadencia/explanation` foi publicada sem alterar o endpoint legado. O segredo `CADENCIA_WORKER_TOKEN` foi configurado no Worker e na VPS; uma chamada sintética autenticada respondeu `200` usando `openai/gpt-oss-20b`. O serviço Ollama foi instalado na composição de produção, sem porta pública, e o modelo `qwen3:4b-instruct` foi baixado e testado, mas permanece parado para não pressionar a VPS. A variável `AI_ENABLED` está `true` na VPS com `AI_PROVIDER=worker`; o valor seguro padrão permanece `false`.

### Interface e PWA

- Dashboard, plano, atividades, avaliação, recuperação e evolução.
- Modal de sessão no mobile, check visual de treinos concluídos e logout.
- O painel principal prioriza a sessão em andamento antes de procurar o próximo treino planejado, mantendo o estado consistente após iniciar pela tela inicial.
- Os gráficos semanais da Evolução exibem o intervalo completo de cada semana para deixar claro que os valores são agrupados por período de sete dias.
- Informativo de novidades versionado no primeiro acesso autenticado: aparece uma vez por conta e versão neste navegador, com linguagem simples e os principais recursos da atualização. Cada funcionalidade visível deve atualizar `APP_VERSION` e `UPDATE_NOTES` na mesma entrega; essa exigência foi cumprida na publicação do piloto XCO em `0.8.0`.
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

- Migrações versionadas no checkout local e aplicadas em produção: `000001` a `000029`. As migrações mais recentes foram executadas em ordem pelo perfil `maintenance` após o backup verificável `cadencia-20260914T224807Z.dump`, revisão e autorização explícita.
- `000012` adiciona confirmação de e-mail e recuperação de senha.
- Produção possui registro de migrações em `cadencia_schema_migrations`.
- O usuário da API não é superusuário; o proprietário do banco é reservado para operações administrativas.
- Não há alterações de esquema aplicadas parcialmente na produção. Novas migrações devem continuar sendo executadas em ordem pelo perfil `maintenance`, após backup verificável.

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

### Décima primeira fatia de melhorias — rotação segura da recuperação (commit `ac8ed3b`)

- O catálogo geral passa a ter o protocolo `active_recovery`, apresentado como **Recuperação ativa**. Ele é selecionado de forma determinística para uma sessão de base na quarta semana do ciclo, que já usa multiplicador de recuperação, esforço-alvo RPE 3,5 e instrução de pedal leve contínuo.
- A escolha reutiliza a evidência geral `acsm-1998` e não copia dose de um estudo específico. A sessão não cria modalidade nova, não aumenta carga, não usa sprint, potência obrigatória ou metas fisiológicas universais.
- Limitação ativa, dor e sinais recentes de recuperação insuficiente continuam vencendo a variação e produzem `Giro leve protegido`. A regra também não altera a seleção dos protocolos específicos de estrada ou XCO.
- A tela de novidades foi atualizada para `0.9.0`, conforme a decisão de comunicar toda funcionalidade visível. Esta fatia foi validada localmente e está no commit `ac8ed3b`; ainda não foi publicada ou aplicada na produção.

Validação desta fatia: `go test -count=1 ./...`, `go vet ./...` e `npm run build` passaram. O build manteve apenas o aviso preexistente do Vite sobre importação JSON e classificação de rotas dinâmicas. Não foi feita validação visual no navegador nesta etapa.

Arquivos desta fatia: `backend/internal/planning/protocols.go`, `backend/internal/planning/service.go`, `backend/internal/planning/service_test.go`, `frontend/lib/release.ts`, `docs/cycling-evidence-catalog.md`, `docs/training-adaptation-rules.md`, `docs/project-status.md`, `docs/architecture-decisions.md`, `docs/README.md` e `planejamento.md`. Não foram alteradas migrações, infraestrutura ou produção.

### Décima segunda fatia de melhorias — registro explícito de treino não realizado

- A interface passa a oferecer **Não realizei** para treinos `planned` ou `adapted` cuja data já passou. A confirmação registra o treino como `skipped` sem criar uma sessão artificial.
- A nova rota `POST /v1/workouts/{workoutID}/missed` valida o plano ativo, o estado e a data no PostgreSQL. Treinos futuros, iniciados, concluídos ou já encerrados não podem usar essa transição.
- O registro fecha a pendência vencida para a leitura de aderência, mas não reage automaticamente: não há reagendamento, sessão substituta, aumento/redução de carga ou inferência de destreinamento.
- A tela de novidades foi atualizada para `0.10.0`. A proteção contra divergência de hidratação calcula a data local da ação somente depois do primeiro render. Não houve migração, mudança de infraestrutura, deploy ou alteração do motor prescritivo `rules-v1`.

Validação desta fatia: `go test -count=1 ./...`, `go vet ./...`, `npm run build`, `oxlint` do componente alterado, OpenAPI, teste transacional PostgreSQL com `ROLLBACK`, consulta de histórico e confirmação visual no navegador local passaram. O lint geral ainda aponta débitos preexistentes em arquivos não tocados. O commit `051d285` foi feito; não foi feito deploy.

### Décima terceira fatia de melhorias — piloto de intervalos intensos para estrada

- O catálogo local passa a ter `road_high_intensity_intervals`, apresentado como **Intervalos intensos de estrada**. O estímulo usa até cinco blocos de 8 minutos com 4 minutos leves, alvo RPE 8 e uma única sessão de qualidade por semana; a estrutura reduz blocos quando a duração adaptada não comporta o formato completo.
- A seleção exige disciplina `road`, nível avançado, pelo menos oito semanas e três pedais semanais recentes, avaliação submáxima apta, objetivo de performance/evento, pelo menos 75 minutos disponíveis, ciclo alternado e semana de construção. Dor, limitação, recuperação insuficiente, dados inelegíveis, baixa consistência e perfis abaixo do avançado continuam impedindo o piloto.
- A fonte principal é `road-block-comparison-2025`, complementada por `rosenblat-2020`. O estudo recente envolveu ciclistas bem treinados e blocos concentrados; o Cadência não copia sua frequência, seu RPE ou sua carga. A implementação local é uma adaptação conservadora, sem potência obrigatória, sem sprint máximo e sem alteração do `rules-v1` fora dessa escolha contextual.
- A funcionalidade é visível, então `frontend/lib/release.ts` foi atualizado para `0.11.0` e a tela de novidades informa a disponibilidade restrita. A implementação foi registrada no commit local `4312fa9`; não houve deploy, migração ou mudança de infraestrutura.

### Correção da semana de recuperação — versão local 0.11.1

- A quarta semana do ciclo não transforma mais o slot de qualidade em sessão de qualidade. O pedal mais longo continua como `Endurance contínuo` e os demais passam a `Recuperação ativa`, preservando o multiplicador reduzido.
- Isso impede que `Ritmo de prova controlado`, `Tempo controlado` ou intervalos apareçam na semana marcada como `RECUPERAÇÃO`, inclusive quando há meta de prova, avaliação apta e preferência por intervalos.
- A tela de novidades foi atualizada para `0.11.1`. O teste de regressão para meta de prova e a suíte Go completa passaram, assim como `go vet`; não houve commit, deploy, migração ou mudança de infraestrutura nesta correção.

### Auditoria de segurança e correções — versão 0.12.0 publicada

Após a revisão do código, foram corrigidos os bloqueadores funcionais e de segurança que podiam afetar esta etapa: sessões `adapted` agora podem ser iniciadas; corpos JSON da API têm limite e rejeitam campos desconhecidos; autenticação possui limites por janela; mutações autenticadas exigem evidência de origem; respostas recebem cabeçalhos de proteção; cadastro e recuperação reduzem enumeração de contas; e o timestamp do início do treino não é renderizado antes da hidratação do cliente.

Também foi corrigida a regra de meta de prova: datas passadas são rejeitadas, a comparação usa o fuso local de forma consistente e a fase específica só é elegível na janela próxima ao evento, preservando a progressão regular para eventos distantes. A data e a distância ficam no snapshot do plano. O resumo semanal usa trava advisory transacional para impedir execução concorrente do timer e de uma execução manual. A tela de novidades só grava a confirmação depois de o usuário dispensá-la, e `APP_VERSION`/`UPDATE_NOTES` foram atualizados para `0.12.0`.

Validação local desta fatia: `go test -count=1 ./...`, `go vet ./...`, `npm run build`, lint direcionado dos componentes alterados e `npm audit --omit=dev --audit-level=high` passaram; este último reportou zero vulnerabilidades no grafo de produção. O lint geral ainda possui pendências anteriores fora desta fatia. Os avisos restantes do grafo de desenvolvimento não têm correção automática disponível e não entram na imagem/runtime de produção. `govulncheck` não está instalado neste ambiente, portanto não foi usado como evidência desta rodada.

As correções desta auditoria e o pin do digest do Tunnel foram commitados em `e803d46` e `61d7939`. O deploy oficial foi concluído na VPS em 11 de setembro de 2026 após backup, build, manutenção e validação dos serviços. A produção está no commit `61d7939`/versão `0.12.0`; o compose fixa `cloudflare/cloudflared@sha256:e39ee8…`, correspondente ao `cloudflared 2026.7.3` ARM64 validado na VPS.

## Feedback de produto e recebimento dos relatos

O fluxo inicial foi desenhado para a divulgação do MVP em grupos de ciclismo: cada pessoa cria uma conta, abre a aba `Feedback` e envia uma categoria, uma nota e um relato livre. O backend exige autenticação, valida tamanho e categoria e armazena o registro em `user_feedback` ligado ao usuário; ele não altera o plano nem dispara IA.

Nesta primeira etapa, os relatos continuam centralizados no banco e não geram um e-mail individual. O recebimento escolhido é um resumo semanal por e-mail, enviado pelo job separado ao endereço administrativo `FEEDBACK_DIGEST_TO`; os relatos são marcados como enviados para não reaparecerem no próximo resumo. O timer systemd está ativo na VPS e o primeiro ciclo será observado quanto a entrega e utilidade. Uma tela administrativa protegida ou exportação controlada pode ser avaliada depois, se o volume justificar. Não expor o banco nem liberar uma listagem administrativa ao usuário final faz parte do escopo de segurança.

## Próximas etapas do produto

O MVP de ciclismo está concluído. As próximas atividades são de operação e evolução controlada, não de implementação obrigatória para considerar esta versão pronta:

1. Observar feedback real e o resumo semanal do Resend, acompanhando entrega e utilidade sem transformar um caso isolado em autorização de carga.
2. Manter cópias externas dos backups e concluir o hardening da VPS sem interromper Tailscale, Cloudflare ou os demais aplicativos.
3. Reduzir gradualmente o débito do `npm run lint`; os testes Go, `go vet` e o build do frontend permanecem aprovados, mas o lint geral ainda possui pendências antigas.
4. Continuar a revisão científica e a coleta longitudinal. `rules-v1` continua como única autoridade; os shadows não devem ganhar autoridade sem calibração, efeito longitudinal e revisão adequada.
5. Avaliar integrações externas, como Strava, somente depois de definir escopo, consentimento, custos e segurança dos tokens.

Corrida e musculação estão fora deste produto e não devem reaparecer como tarefas do Cadência. O histórico abaixo preserva as decisões e validações das fatias anteriores; ele não representa pendências atuais.

### Vigésima primeira fatia — contexto detalhado de segurança do perfil — 13 de setembro de 2026

O formulário de limitações da etapa 2 do perfil agora permite registrar, opcionalmente, localização, intensidade percebida de 1 a 10, movimento agravante e data de início. O backend normaliza espaços, limita o tamanho dos textos, rejeita intensidade fora da faixa e não aceita data futura. A tela informa que esses dados não são diagnóstico e mantém a recomendação de orientação profissional quando aplicável.

A migração `000022_limitation_context` adiciona os campos ao PostgreSQL sem apagar registros existentes. A consulta do planejamento continua lendo somente o tipo da limitação e a recomendação de liberação profissional; portanto, os detalhes não entram como diagnóstico, não alteram a duração ou o RPE e não transferem autoridade ao `rules-v2`. O `rules-v1` e os planos existentes permanecem inalterados.

A versão local passou para `0.21.0` e a nota foi registrada em `frontend/lib/release.ts`, para aparecer na tela de novidades após a atualização. A migração `000022` foi aplicada no PostgreSQL local e a validação manual confirmou persistência e remoção da limitação. `go test -count=1 ./...`, `go vet ./...`, `npm run build` e `git diff --check` passaram. O `npm run lint` geral continua apontando pendências anteriores em arquivos do projeto, além dos avisos do próprio `page.tsx`; elas não foram ampliadas para uma limpeza fora do escopo. A implementação foi registrada no commit `9514a00`; ainda falta backup, aplicação da migração em produção e autorização própria para publicação. A produção permanece em `0.20.0` e `000021`.

### Continuidade — erros de autenticação separados de falhas da API — versão local `0.22.0`

O cliente HTTP agora preserva o status da resposta em `ApiError` e traduz falhas de conexão. Nas telas autenticadas, somente `401 Unauthorized` redireciona para `/entrar`; falhas `500`, indisponibilidade da API ou problemas temporários de banco permanecem na rota solicitada e exibem uma mensagem com opção de tentar novamente. Isso evita mascarar falhas operacionais como sessão expirada.

A mudança é somente frontend, não altera autenticação, banco, prescrição ou produção. A nota `0.22.0` foi adicionada à tela de novidades. `npm run build` e `git diff --check` passaram; o lint direcionado continua apontando regras antigas já existentes no `api.ts` e em páginas do frontend. A validação manual confirmou a permanência da rota com a API desligada e o redirecionamento quando a sessão realmente expira. O commit `77e57ec` foi registrado e está no remoto; produção permanece em `0.20.0` até publicação própria.

### Continuidade — integridade do registro pós-treino mais clara — versão local `0.23.0`

O gate `data-integrity-v1` agora identifica também uma combinação operacionalmente incompatível entre distância e duração, sem transformar o corte em limite fisiológico ou prescrição de velocidade. O registro original continua salvo, mas permanece fora da observação histórica quando há inconsistência.

Após concluir um treino, a tela passa a informar quando o registro foi preservado para revisão e não será usado no histórico observado. A nota `0.23.0` foi adicionada à tela de novidades. A mudança não altera autenticação, banco, catálogo, `rules-v1` ou infraestrutura. `go test -count=1 ./...`, `go vet ./...`, `npm run build` e `git diff --check` passaram. A validação manual confirmou `distance_duration_incompatible`, `eligible_for_history: false`, `progression_eligible: false` e `used_for_prescription: false` após concluir um treino com 3 minutos e 55 km. O commit `74f9493` foi registrado; produção permanece em `0.20.0` até deploy autorizado.

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

### Continuidade — correção segura de métricas do pedal — versão local `0.24.0`

Sessões concluídas marcadas por `data-integrity-v1` como `incomplete` ou `inconsistent` agora podem receber uma correção controlada somente nas métricas opcionais do pedal: distância, elevação, potência média e frequência cardíaca média. A rota autenticada `POST /v1/workouts/{workoutID}/correct` exige que o treino esteja concluído e inelegível; duração, RPE, feedback, plano ativo e prescrição ficam fora do escopo da operação.

Cada correção preserva os valores anteriores em `workouts.explanation.data_integrity_corrections`, reavalia o gate de integridade dentro da mesma transação e só permite que a sessão volte à observação histórica quando os dados corrigidos forem coerentes. A interface explica que campos vazios removem a métrica e informa que o plano não será recalculado. `rules-v1` continua prescritivo e `used_for_prescription` permanece falso.

Validação concluída localmente: `go test -count=1 ./...`, `go vet ./...`, `npm run build` e `git diff --check` passaram. O build manteve apenas os avisos preexistentes do Vite sobre importação JSON nativa e classificação de rotas dinâmicas. No navegador, uma sessão com 3 minutos e 55 km foi corrigida para `0,3 km`; a mensagem de sucesso apareceu, o aviso desapareceu após `F5`, duração e RPE permaneceram inalterados e a consulta PostgreSQL confirmou `status: valid`, `distance_km: 0.3` e o valor original `55` no histórico de correção. A API precisou ser reiniciada para carregar a rota nova; sem a sessão, a rota respondeu `401`, confirmando que o endpoint estava registrado. Não houve migração, deploy ou alteração de infraestrutura; a produção permanece em `0.20.0`.

### Continuidade — contexto estruturado do feedback pós-treino — versão local `0.25.0`

O formulário de conclusão de um pedal agora permite registrar, de forma opcional e controlada, satisfação da sessão em escala de 1 a 5, terreno (`flat`, `rolling`, `hilly`, `mixed`, `technical` ou `indoor`) e condições externas (`normal`, `heat`, `cold`, `wind`, `rain`, `poor_visibility` ou `other`). O backend valida as faixas e listas permitidas, a migração `000023_feedback_context` adiciona os campos sem reescrever feedbacks antigos e o histórico de `/atividades` passa a exibir os valores disponíveis.

Os três sinais também aparecem no `post-workout-context-v2` e no `decision_audit.data_used` somente como observação. Eles não alteram o `rules-v1`, não liberam progressão e não entram como autorização de carga; a prescrição continua isolada até haver calibração e evidência longitudinal suficientes. A documentação OpenAPI e a nota `0.25.0` foram atualizadas.

A validação local passou com `go test -count=1 ./...` usando cache local do Go, `go vet ./...`, `npm run build`, o teste SQL transacional da migração e `git diff --check`. No navegador, o formulário exibiu os três controles, uma sessão foi concluída com satisfação `5/5`, terreno ondulado e vento, o plano mostrou os dados e `/atividades` confirmou a persistência. A consulta PostgreSQL confirmou `post-workout-context-v2`, estado `observed` e os três campos em `decision_audit.data_used`. Não houve deploy; a produção permanece em `0.20.0` e `000021`.

### Continuidade — consistência do contexto estruturado no shadow — versão local `planned-vs-actual-v2`

A revisão após a `0.25.0` encontrou uma diferença de cobertura: satisfação, terreno e condições externas já eram armazenados, exibidos e considerados pelo `post-workout-context-v2` e pela auditoria, mas não eram validados pelo gate de integridade nem propagados para `planned_vs_actual`. A cadeia observacional agora usa os mesmos valores e as mesmas listas controladas nos três pontos.

Campos opcionais ausentes ficam explícitos em `missing_data`; valores inválidos tornam a comparação e a integridade `not_evaluated`/`inconsistent`, conforme o componente, sem apagar o feedback. A conclusão e a correção de métricas preservam esse contexto durante a reavaliação. Não houve migração, mudança visual, atualização de release, deploy ou alteração do `rules-v1`; os campos de autoridade continuam falsos.

### Continuidade — cobertura longitudinal do feedback estruturado — versão local `training-history-v5`

As consultas de histórico agora contam, nas janelas de 7/28/42 dias e nos seis períodos semanais, as sessões elegíveis com satisfação, terreno e condições externas válidos. A cobertura é separada por campo para distinguir uma sessão que informou apenas parte do contexto; ausências ficam explícitas no `missing_data` do snapshot e da comparação.

O snapshot passou a `training-history-v5` e o comparador a `period-comparison-v3`. A mudança é somente observacional: não interpreta tendência, tolerância ou efeito longitudinal, não alimenta progressão e mantém `rules-v1` como autoridade. `go test -count=1 ./...`, `go vet ./...`, `npm run build` e `git diff --check` passaram; não houve migração, mudança visual, release, deploy ou alteração de infraestrutura.

### Continuidade — catálogo com pedal longo explícito — versão local `0.26.0` validada

O maior slot de disponibilidade do ciclo agora é apresentado como `Pedal longo` e usa a chave estável `long_endurance`. A sessão mantém o mesmo RPE-alvo 5, duração limitada pelo nível, disponibilidade e multiplicador, estrutura contínua e todas as proteções existentes; a alteração diferencia o catálogo sem aumentar carga.

Os shadows de periodização e seleção de estímulos reconhecem a nova chave como endurance longo. A nota local `0.26.0` foi adicionada, sem migração, deploy ou alteração de infraestrutura. A validação automatizada e a conferência manual do novo nome passaram.

### Deploy da versão `0.27.0` — 14 de setembro de 2026

O perfil agora diferencia **Não informar**, **Estou treinando regularmente** e **Estou retornando após uma pausa**. Somente a última opção ativa o protocolo `return_after_break`, apresentado como **Retorno gradual**; a regra limita as sessões a 45 minutos e RPE 3,5 e substitui qualidade e maior volume durante a retomada. Semanas preenchidas, sozinhas, não reduzem o plano.

As proteções de limitação, dor e recuperação continuam prioritárias; os shadows reconhecem a necessidade de retorno sem ganhar autoridade prescritiva. A nota `0.27.0` foi ajustada e a versão foi publicada sem migração nova ou alteração de infraestrutura. O backup `cadencia-20260914T102918Z.dump` foi criado e verificado; API, frontend, PostgreSQL e Tunnel ficaram saudáveis, `/ready` respondeu corretamente, os dois domínios públicos retornaram HTTP 200 e o HTML público contém `0.27.0`. A validação manual do menu e dos dois comportamentos passou. A release do GitHub ainda não foi criada.

### Estado atual do checkout — contexto de equipamento e sinais de segurança — base local `0.29.0`

- O feedback pós-treino agora aceita `equipment_used` opcional, limitado a 120 caracteres. O valor é persistido, aparece no resultado e em `/atividades` e entra nas leituras observacionais sem alterar carga ou prescrição.
- O perfil agora aceita sintomas de alerta durante/depois do treino e uma restrição médica atual. A API limita os sintomas a cinco opções controladas, rejeita duplicidades e mantém o contexto sem tratá-lo como diagnóstico.
- O planejamento registra `safety_context` no snapshot e explica quando há restrição médica. A proteção de limitação existente continua vencendo qualidade, taper e progressão; além disso, o início revalida uma limitação criada depois do plano e bloqueia sessões acima de RPE 4, sem permitir bypass; `rules-v1` permanece prescritivo.
- A validação de conclusão passou a rejeitar RPE/distância não finitos e contexto de equipamento acima do limite. As migrações `000024_equipment_feedback` e `000025_limitation_safety_signals` são aditivas.

Validação concluída no checkout local: `go test -count=1 ./...`, `go vet ./...`, `npm run build`, `git diff --check` e os testes SQL transacionais `database/tests/000024_feedback_equipment.sql` e `database/tests/000025_limitation_safety_signals.sql` passaram. A validação manual no navegador também passou, incluindo os ajustes responsivos; a entrega foi registrada nos commits `8ff96e7` e `1da6d91`. O `npm run lint` geral continua falhando por débitos anteriores em componentes e páginas amplas, inclusive regras antigas de acessibilidade e navegação; isso não foi ampliado nem tratado como parte desta fatia. O deploy ainda está pendente, e a produção continua em `0.27.0`/migração `000023`.

### Continuidade — metadados científicos e templates auditáveis — migração local `000026`

A base científica passa a registrar população estudada, objetivo, estímulo, benefícios esperados, limitações, riscos, contraindicações, confiança, data de revisão e regras relacionadas. A migração `000026_scientific_source_metadata` preserva as fontes existentes, usa `not_calibrated` quando a confiança ainda não foi revisada e impede que ausência de calibração seja apresentada como certeza.

Os protocolos já selecionáveis passam a anexar ao treino metadados operacionais versionados: objetivo fisiológico e prático, indicação, contraindicação, nível, pré-requisitos, orientação opcional por frequência cardíaca/potência/cadência, interrupção, progressão e regressão. Duração, RPE, aquecimento, parte principal, recuperações e desaquecimento continuam no próprio treino estruturado. Esta fatia melhora os tópicos 3 e 4 sem criar novo estímulo, alterar carga, mudar o `rules-v1` ou exigir mudança visual.

A migração foi aplicada e registrada somente no PostgreSQL local. O teste SQL transacional confirmou os 25 registros com metadados e rejeitou confiança fora da lista controlada. `go test -count=1 ./...`, `go vet ./...`, `npm run build` e `git diff --check` passaram. O lint geral continua com os débitos já registrados; nenhum erro aponta para os arquivos alterados nesta fatia.

### Continuidade — gates de adaptação e recuperação pós-prova — candidata local `0.30.0`

O campo `readiness_assessment.state` agora separa o contexto atual — dados insuficientes, cautela, recuperação, retorno, baixa consistência, preparação específica para evento ou estabilidade observada — da experiência declarada. A classificação permanece observacional e não autoriza progressão sozinha.

A migração `000027_adaptation_integrity_gate` protege a adaptação ativa do `rules-v1`: feedback parcial, duração/RPE/fadiga ausentes ou inválidos e sinais protetivos nos 14 dias anteriores não alteram treinos futuros. Quando a adaptação é permitida, o treino alterado recebe `status: adapted`, preservando o ciclo de estados já suportado pela API.

O protocolo `post_event_recovery` usa a janela de até sete dias após o evento, limita a sessão a 45 minutos e RPE 3,5 e registra as fontes `post-competition-recovery-2019` e `recovery-umbrella-2024` na migração `000028`. A evidência é heterogênea; o protocolo é uma proteção operacional, não tratamento nem dose universal. Retorno, dor, limitação e recuperação insuficiente prevalecem.

A migração `000029_average_cadence_metric` registra `average_cadence_rpm` entre 1 e 300 como métrica opcional. O valor é exibido no resumo e no histórico, validado no banco, API e integridade, e entra somente como campo observado ou ausente em `planned-vs-actual-v3`; não cria meta de cadência, não interpreta desempenho e não altera o `rules-v1`.

Quando o histórico de 28 dias registra treino perdido ou vencido, o `rules-v1` ativo adia a sessão de qualidade e explica a decisão em cada sessão; o shadow continua auditando a necessidade e não ganha autoridade adicional.

Os fixtures SQL `000005`, `000020`, `000027`, `000028` e `000029` foram executados localmente em transações revertidas. `go test -count=1 ./...`, `go vet ./...`, `npm run build`, o lint específico dos arquivos alterados e `git diff --check` passaram. O lint geral mantém apenas débitos anteriores fora desta fatia. Esta fatia ainda não foi commitada, publicada ou aplicada na produção; a produção continua em `0.27.0` e schema `000023`.

### Auditoria real do roadmap antes da candidata `0.30.0`

Implementação agora fechada ou coberta por base testável: escopo exclusivo de ciclismo (tópico 14), parte do catálogo e elegibilidade (4), prontidão observacional (1), regras versionadas em shadow (2), diferenciação por situação de treino (5), integridade e correção auditável (11), coleta estruturada de feedback (12), auditabilidade (13) e regressões automatizadas principais (15).

Ainda não é correto marcar como validados em campo: adaptação em ciclo fechado com autoridade (6), calibração de carga/progressão e efeito da prescrição (7), periodização individual de longo prazo (8), seleção plenamente orientada pelo efeito (9), calibração de segurança clínica (10), novos protocolos fora do catálogo elegível (3 e 4) e critérios que exigem dados reais (16). A matriz de aceitação automatizável do tópico 15 está fechada; as condições restantes dependem de revisão científica específica, volume longitudinal e decisão explícita para ativar qualquer autoridade nova.

### Fechamento verificável do roadmap — candidata local `0.30.0`

A matriz [`roadmap-acceptance.md`](roadmap-acceptance.md) passa a relacionar cada um dos 16 tópicos de `melhorias.md` ao código, contrato e testes que o cobrem. Ela diferencia o que está implementado no checkout das condições externas que nenhum código pode substituir: dados longitudinais, revisão científica/clínica e autorização para alterar a autoridade de prescrição.

Cada sessão nova recebe `workout-decision-audit-v1` dentro de `explanation`: regras avaliadas e aplicadas, dados usados e ausentes, restrições, alternativas descartadas e condições de mudança. A tela do plano apresenta esse contexto de forma recolhível em **Ver detalhes da decisão**. O campo não é um shadow: descreve o `rules-v1` que efetivamente gerou a sessão, com confiança explicitamente `rule_based_not_calibrated`.

O contrato OpenAPI e o tipo do frontend foram alinhados a `planned-vs-actual-v3`; a versão anterior de ambos ainda declarava `v2` mesmo depois de o backend passar a observar a cadência média. A entrega continua local, sem commit, deploy ou migração em produção.

### Vigésima segunda fatia — piloto local de limiar controlado — versão `0.28.0` validada localmente

O perfil agora oferece a preferência opcional **Limiar**. O motor pode apresentar **Limiar controlado** somente para estrada ou indoor, nível avançado, objetivo de performance/prova, avaliação submáxima apta, oito semanas de treino recente, três pedais semanais, pelo menos 60 minutos disponíveis e fase compatível com o evento. A sessão usa três blocos de 8 minutos em RPE 7,5 com 4 minutos leves, sem potência universal ou estimativa automática de limiar.

O protocolo permanece no `rules-v1` local como piloto explícito e não altera o `rules-v2`, os shadows, a carga por feedback ou as proteções de dor e recuperação. Não há migração. `go test -count=1 ./...`, `go vet ./...`, `npm run build` e `git diff --check` passaram, e a validação manual confirmou a seleção elegível e a não seleção nos bloqueios. A funcionalidade foi registrada no commit `d6e36ec`; o commit `011b204` removeu os artefatos de cache gerados e adicionou `.gocache/` ao `.gitignore`. Produção continua na `0.27.0`, sem deploy ou release desta fatia.
