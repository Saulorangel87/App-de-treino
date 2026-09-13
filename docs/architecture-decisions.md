# Decisões de arquitetura

Última revisão: 13 de setembro de 2026.

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

O contexto de ciclismo permanece em JSONB para evoluir sem migrações a cada pergunta opcional. Atualmente inclui horas semanais, pedais por semana, distância semanal recente, semanas de regularidade, maior distância e pedal, preferências de sessão, equipamento, terreno e sensores. Além dele, o motor consulta um resumo agregado dos últimos 28 dias de sessões concluídas e check-ins de recuperação. Esses dados são preservados no snapshot do plano; sinais de dor, fadiga elevada ou recuperação insuficiente apenas protegem a sessão de forma conservadora, sem criar metas rígidas ou diagnósticos. O catálogo valida localmente pilotos de intervalos aeróbicos XCO, VO₂max de estrada e intervalos curtos autorregulados com disciplina/preferência explícitas, evidências próprias e limites restritos. Sprint/pista/BMX e downhill/enduro não fazem parte do produto: foram removidos do perfil e deixaram de ser aceitos pela API. Novas fórmulas de carga só serão ativadas após revisão e testes específicos.

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

As migrações `000001` a `000022` estão versionadas no checkout; a produção está aplicada até `000021`. A `000013` cria os relatos de feedback, a `000014` adiciona o controle de envio do resumo semanal, a `000015` registra fontes do catálogo inicial, a `000016` registra a fonte do piloto XCO, a `000017` registra as fontes do taper pré-prova, a `000018` registra as fontes do piloto VO₂max de estrada, a `000019` registra as fontes do piloto de intervalos curtos, a `000020` adiciona o contexto de conclusão parcial, a `000021` adiciona o contexto pós-treino e a `000022` adiciona contexto opcional de segurança às limitações. As migrações `000017`, `000018` e `000019` foram aplicadas pelo perfil `maintenance` no deploy do commit `6fdbe45`, após backup verificável; `000020` e `000021` foram aplicadas no deploy da versão `0.20.0`. Antes de qualquer nova mudança estrutural em produção, deve existir backup verificável e a migração deve ser executada pelo perfil `maintenance`.

## ADR-006 — Feedback de produto

**Status:** Implementada e publicada; migrações `000013` e `000014` aplicadas na produção.

O feedback solicitado aos primeiros ciclistas é separado do feedback pós-treino. A rota autenticada `POST /v1/feedback` aceita somente uma categoria (`experience`, `bug` ou `suggestion`), uma nota de 1 a 5 e uma mensagem entre 10 e 2000 caracteres. O registro é vinculado ao usuário na tabela `user_feedback`, sem armazenar um e-mail duplicado ou permitir conteúdo anônimo nesta primeira versão.

A tela `/feedback` é acessível pelo menu principal no desktop e por um atalho no cabeçalho em telas menores. O envio não chama a IA e não altera o plano. Os relatos são consolidados no PostgreSQL e não geram e-mail individual; um job separado reúne até 50 relatos pendentes em um único resumo semanal via Resend, destinado somente ao endereço administrativo configurado em `FEEDBACK_DIGEST_TO`.

## ADR-007 — Resumo semanal de feedback

**Status:** Implementada e ativa na produção.

O resumo é executado fora do processo HTTP, como um comando de curta duração (`cadencia-feedback-digest`) acionado semanalmente por systemd. O comando consulta no máximo 50 relatos ainda não enviados, escapa o conteúdo no HTML, envia uma mensagem HTML e texto pelo remetente Resend já verificado e só grava `digest_sent_at` depois de o envio retornar sucesso. Uma trava advisory transacional no PostgreSQL serializa uma execução do timer e uma execução manual concorrente. Se `FEEDBACK_DIGEST_TO` estiver vazio, a execução termina sem consultar o banco nem enviar mensagens. O serviço acessa o PostgreSQL pela rede Docker interna e o Resend por uma rede de saída dedicada; não publica portas e não fica residente, reduzindo consumo na VPS. A migração e o timer já foram aplicados; o primeiro envio será acompanhado em operação. Como não existe transação distribuída entre PostgreSQL e o provedor de e-mail, uma queda exatamente após o envio e antes do commit ainda pode exigir conferência manual de duplicidade.

O timer está agendado para segunda-feira às 11:00 UTC (08:00 no horário de São Paulo) e possui `Persistent=true` para executar após uma indisponibilidade. O endereço administrativo é único nesta primeira versão; a quantidade de mensagens permanece previsível e deve ser acompanhada junto das demais aplicações que usam a mesma conta Resend. Um painel administrativo protegido continua como possível evolução, não como dependência do MVP.

## ADR-008 — Evolução pós-MVP do motor

**Status:** Aceita como direção planejada; ainda não implementada como substituição do `rules-v1`.

A próxima fase seguirá o roadmap de `melhorias.md` com foco exclusivo em ciclismo: classificação explícita de prontidão, regras versionadas, adaptação em ciclo fechado, progressão/carga, integridade dos dados, segurança, feedback e auditabilidade. O `rules-v1` continuará preservado até que uma evolução paralela esteja testada, comparável e auditável. Dados ausentes ou inconsistentes não devem ser tratados como autorização para aumentar carga. A IA continua explicativa e subordinada às regras validadas; não prescreve, inventa evidências ou diagnostica.

Primeira fatia local, em 5 de setembro de 2026: `readiness-v1` é uma função determinística em `backend/internal/planning/readiness.go`, salva no snapshot do plano em modo `observation`. Não substitui `rules-v1` nem o nível de prontidão calculado pelo check-in diário. Usa somente limitações ativas, os agregados de 28 dias já existentes e contagens explícitas de cobertura dos campos. Não depende do nível de experiência, de aprovação antiga na avaliação ou de volume autodeclarado para concluir que há prontidão. Não há migração nem recálculo retroativo de planos. A ausência de sessões não permite inferir baixa consistência; a progressão permanece não avaliada até haver aderência, tolerância e qualidade temporal dos dados. O recebimento de feedback real e do Resend é acompanhamento paralelo, não pré-requisito de desenvolvimento.

Segunda fatia no commit local `a4e5f9f`, ainda sem publicação: `training-history-v1` mede janelas cumulativas de 7, 28 e 42 dias em modo observacional. A aderência usa a data planejada e apenas sessões já fechadas: treinos de hoje entram depois de concluídos ou cancelados; datas anteriores ainda `planned`/`adapted` são pendências vencidas, e `in_progress` vencido fica separado. A carga realizada usa `completed_at`, duração positiva e RPE real entre 1 e 10; o produto registra `duração × RPE` em unidades arbitrárias e explicita quantas sessões não têm dados suficientes. Planos `draft` ou `cancelled` não compõem o denominador de aderência. Sessões concluídas do atleta continuam compondo a carga realizada independentemente do plano de origem. As duas bases temporais, as lacunas e as referências ficam no snapshot. As janelas não formam ACWR, não geram limiares de aderência e não participam da prescrição nesta fatia.

Terceira fatia local, ainda sem commit ou publicação: `training-history-v2` acrescenta recência, exclusão explícita de registros futuros, cobertura de feedback/check-in e contagens dos sinais protetivos já usados pelo motor. O snapshot diferencia feedback inexistente de feedback incompleto e preserva dor mesmo quando falta outro campo. Ausência de registros no Cadência é rotulada apenas como lacuna de atividade registrada; destreinamento, perda de condicionamento, tolerância e progressão permanecem em `not_evaluated`. A decisão evita extrapolar estudos de cessação total para redução de treino ou falta de sincronização do atleta. `used_for_prescription` permanece `false`, e planos antigos `training-history-v1` continuam legíveis sem recálculo.

Quarta fatia no commit local `810183c`, ainda sem publicação: `training-history-v3` acrescenta `period_comparison` com seis blocos semanais não sobrepostos, do mais recente ao mais antigo. Cada bloco repete medições brutas de aderência, sessões, carga session-RPE, feedback, sinais protetivos e recuperação; não calcula tendência, ACWR ou qualquer razão que autorize carga. Sessões e carga usam intervalos de `completed_at` no relógio do banco, enquanto aderência e recuperação usam datas de `CURRENT_DATE`, mantendo as bases explícitas. O contrato registra `period-comparison-v1`, lacunas e inconsistências, mantém `used_for_prescription: false` e deixa a tendência para prescrição em `not_evaluated`. Snapshots antigos continuam legíveis sem recálculo. A conferência manual via API foi concluída.

Quinta fatia nos commits locais `64e554d` e `2359c3f`, ainda sem publicação: `rules-v2` é executado em modo `shadow` durante a geração do plano. Ele avalia integridade dos períodos, sinais protetivos e evidência mínima de dois períodos recentes, registrando regras avaliadas, regras adiadas, motivos, lacunas e uma resposta candidata. Os estados são `protective_signal`, `observation_only` e `not_evaluated`; nenhum deles altera a prescrição. `engine_version` continua `rules-v1`, e `progression_eligible`, `applied` e `used_for_prescription` permanecem `false`. A conferência manual no `GET /v1/plans/current` e a matriz controlada confirmaram esse isolamento. A decisão permite comparar o comportamento antes de qualquer migração para um motor prescritivo.

Sexta fatia no commit local `de23add`, ainda sem publicação: `rules-v2-adaptation-v1` avalia o feedback pós-treino em modo `shadow`. Ela mantém proteção para dor/esforço alto, transforma uma resposta fácil isolada em `defer_progression` e só registra `progress_duration_5pct` como candidata quando há dois períodos recentes íntegros com carga, feedback e recuperação. Todos os resultados mantêm `progression_eligible`, `applied` e `used_for_prescription` como `false`; o trigger e o comportamento prescritivo do `rules-v1` permanecem inalterados.

Sétima fatia no commit local `b6ea8bd`, ainda sem publicação: o repositório executa essa avaliação dentro da mesma transação que conclui a sessão e grava o resultado somente em `workouts.explanation.adaptation_shadow`. A consulta dos períodos usa um savepoint para que uma falha produza `not_evaluated`/`history_query_failed` sem impedir o feedback principal. A tipagem do frontend e o contrato OpenAPI foram atualizados para leitura do campo, sem mudança visual ou de versão do produto. A conferência manual confirmou a resposta protetiva e as barreiras de não aplicação.

Oitava fatia no commit local `9034287`, ainda sem publicação: uma matriz regressiva compara o resultado do `rules-v1` com o candidato do `rules-v2-adaptation-v1` em cenários protetivos, neutros, de progressão com evidência suficiente, de progressão sem evidência e de histórico inconsistente. A matriz exige coincidência nas proteções e mantém o shadow não autoritativo em todos os casos. Ela não altera carga, banco, trigger ou interface.

Nona fatia no commit local `34efbe2`, ainda sem publicação: `data-integrity-v1` classifica sessões concluídas como válidas, incompletas ou inconsistentes. O gate separa campos ausentes de valores incompatíveis, preserva o registro original e é anexado à explicação do treino. O shadow não produz candidato quando a sessão atual não é elegível para histórico; o `rules-v1` permanece inalterado até uma decisão específica sobre uma barreira prescritiva.

Décima fatia publicada no commit `9d8c624`: o piloto `xco_aerobic_intervals` adiciona uma sessão aeróbica específica para XCO avançado elegível, com disciplina explícita, avaliação submáxima apta, objetivo compatível, disponibilidade mínima e proteções de recuperação. A migração `000016` registra o ensaio de HIT em mountain bikers que sustenta o formato, enquanto a revisão contemporânea de XCO delimita a transferência da evidência. O piloto não inclui sprint máximo, técnica de trilha, descida, salto ou metas rígidas de potência. Como a escolha é visível, a versão `0.8.0` do frontend e a nota de novidades foram publicadas após validação e autorização.

Décima primeira fatia no commit local `ac8ed3b`, ainda sem publicação: o catálogo geral ganhou `active_recovery`, uma variação determinística para uma sessão de base na quarta semana do ciclo. Ela usa esforço leve, volume reduzido pelo multiplicador de recuperação e a referência geral `acsm-1998`; não é um protocolo específico de modalidade nem uma dose universal derivada de estudo. Limitação, dor e recuperação insuficiente continuam substituindo a variação por `Giro leve protegido`, e o motor `rules-v1` não passa a usar dados shadow para prescrever.

Décima segunda fatia local, ainda sem publicação: o ciclo permite registrar explicitamente um treino passado como não realizado. A transição aceita somente `planned`/`adapted` do plano ativo cuja data seja anterior à data do banco e grava `skipped` sem criar uma sessão. A decisão é factual e não dispara reagendamento, compensação, ajuste de carga ou inferência de destreinamento; o cancelamento de uma sessão já iniciada continua no endpoint próprio.

Décima terceira fatia local, registrada no commit `4312fa9` e ainda sem publicação: o catálogo ganhou `road_high_intensity_intervals`, um piloto de intervalos intensos para estrada avançada em ciclo alternado. Ele exige oito semanas e três pedais semanais recentes, avaliação submáxima apta, objetivo compatível, pelo menos 75 minutos e semana de construção; usa até cinco blocos de 8 minutos com 4 minutos leves e RPE 8. A estrutura é uma adaptação conservadora dos estudos em ciclistas bem treinados, não reproduz bloco concentrado nem cria metas universais de potência. As travas do `rules-v1` para dor, limitação, recuperação e baixa consistência continuam prioritárias.

Correção local seguinte, ainda sem commit: a quarta semana agora bloqueia a classificação `quality` antes de construir a sessão. O pedal longo permanece endurance e os outros slots são `Recuperação ativa`, evitando que a meta de prova ou outro protocolo de qualidade atravesse a semana de recuperação. A versão visível foi atualizada para `0.11.1` e a suíte Go/`go vet` passou; nenhum deploy ou alteração de infraestrutura foi feito.

Décima quarta fatia local, no commit `0eb34d6` e ainda sem publicação: `taper-v1` aplica uma redução de 50% na duração das sessões anteriores a um evento próximo, sem mudar frequência ou RPE. O piloto exige evento entre 7 e 21 dias, atleta avançado, avaliação submáxima apta, base mínima de oito semanas e três pedais semanais, e cede a limitação, dor, recuperação insuficiente e à semana de recuperação. A migração `000017` registra as fontes no banco local; produção permanece em `000016` até validação final, backup e autorização.

A validação local foi concluída em 12 de setembro de 2026. No navegador, uma conta avançada com evento em `20/09/2026` gerou `status: eligible`, `applied: true`, `used_for_prescription: true`, oito dias até a prova e multiplicador `0.5`; cinco sessões pré-prova receberam `event_taper_applied`, com frequência e RPE preservados. Os testes automatizados confirmaram que evento fora da janela e sinais de dor/recuperação mantêm `status: not_applicable` e não aplicam taper. O rascunho não foi aceito, a produção continua em `0.12.0`/`000016` e a publicação segue condicionada à ampliação do catálogo e autorização explícita.

Décima quinta fatia local: o catálogo ganhou `road_vo2_intervals`, com a preferência explícita `vo2max`. O piloto exige estrada, nível avançado, objetivo de performance/evento, avaliação apta, oito semanas e três pedais semanais recentes, 60 minutos disponíveis, semana de construção e ausência de sinais protetivos. Usa quatro blocos de 4 minutos em RPE 8 e quatro minutos leves, sem sprint máximo, cadência ou potência fixa. A migração `000018` registra duas fontes primárias de 2024; a versão local passou para `0.14.0` e a tela de novidades foi atualizada. A suíte automatizada, o build, a migração local e a validação ponta a ponta passaram; o trabalho está no commit `01875c9`. Produção continua em `0.12.0`/`000016`, condicionada a revisão, backup e autorização explícita.

O piloto local de catálogo `short_self_regulated_intervals` avalia intervalos curtos autorregulados para estrada ou indoor. A literatura de 2025 em adultos anteriormente inativos e os estudos em ciclistas treinados sustentam investigar o formato, mas não autorizam uma dose universal nem sprint máximo para todos os perfis. A implementação propõe seis repetições de 1 minuto em RPE 7–8, recuperação leve autorregulada, nível avançado, histórico mínimo e gates de proteção; a validação automatizada e manual local foi concluída, incluindo salvamento da preferência e apresentação do protocolo no plano regenerado. Produção permanece sem o protocolo e sem a migração `000019`.

A pesquisa seguinte avaliou resistência específica na bicicleta como possível estímulo futuro. O ensaio randomizado de 2025 utilizou ciclistas bem treinados, resistência muito alta, cadência muito baixa e calibração por força dinâmica máxima; por isso, a evidência não pode ser convertida para RPE isolado no Cadência. O candidato permanece documentado, sem protocolo no `rules-v1`, até que existam medição, elegibilidade e controles de segurança apropriados. Gravel/XCM continuam contextos de endurance sem protocolo próprio; sprint de pista/BMX e downhill/enduro estão fora do produto.

Fatia técnica local implementada: `load-tolerance-v1` observa os dois períodos semanais mais recentes com carga session-RPE, feedback completo e recuperação registrada. Dor, fadiga alta, recuperação necessária ou esforço atual pelo menos dois pontos acima do alvo mantêm uma resposta protetiva; evidência completa sem esses sinais produz apenas `observation_only`. A leitura fica aninhada em `workouts.explanation.adaptation_shadow`, com `progression_eligible`, `applied` e `used_for_prescription` sempre falsos. Ela não calcula ACWR, não infere tolerância fisiológica e não altera `rules-v1`, o trigger, migrações, interface ou produção.

Fatia técnica seguinte, local e ainda não publicada: `planned-vs-actual-v1` registra a comparação descritiva entre duração e RPE planejados e realizados, junto da cobertura de feedback e métricas opcionais disponíveis. A leitura é anexada ao mesmo `adaptation_shadow`, mantém `progression_eligible: false` e `used_for_prescription: false`, explicita campos ausentes e não interpreta tolerância, progressão ou efeito fisiológico. A validação manual ponta a ponta foi concluída localmente com uma sessão de 3 minutos realizados de 35 planejados, sem alteração da próxima sessão; não há migração, alteração visual, atualização de versão ou deploy.

## ADR-010 — Modalidades fora do escopo do Cadência

**Status:** Aceita e aplicada localmente.

O Cadência não inclui sprint/pista/BMX nem downhill/enduro. Essas modalidades não devem aparecer como opção de perfil, preferência, protocolo ou objetivo de prescrição. A API rejeita os identificadores legados `track_sprint` e `dh_enduro`, e o frontend trata registros antigos como disciplina não informada para impedir que liberem qualquer treino específico. A decisão é de escopo do produto, não depende da existência de um novo estudo ou módulo futuro.

## ADR-009 — Comunicação de atualizações no produto

**Status:** Aceita e aplicada.

Toda atualização com funcionalidade visível deve atualizar `frontend/lib/release.ts`, incrementando `APP_VERSION` e registrando a mudança em `UPDATE_NOTES`. O componente `UpdateNotice` apresenta as notas no primeiro acesso autenticado após a versão mudar e registra a confirmação por conta, versão e navegador usando armazenamento local. As notas não devem conter segredos. A versão `0.7.0` registrou o catálogo de ciclismo baseado em evidências e a versão `0.8.0` registrou o piloto aeróbico de MTB XCO. A versão `0.12.0` reúne a recuperação ativa, o registro de treino não realizado, o piloto de intervalos intensos para estrada, a correção da semana de recuperação, o planejamento por proximidade do evento, as sessões adaptadas iniciáveis, os reforços de segurança, a proteção do resumo semanal e a correção da confirmação da tela de novidades. Ela foi publicada com o commit `61d7939` e a release `v0.12.0`. A versão `0.16.0` acrescenta o piloto de intervalos curtos autorregulados e registra a exclusão de sprint/pista/BMX e downhill/enduro; ela foi publicada no commit `6fdbe45` e está registrada na release [v0.16.0](https://github.com/Saulorangel87/App-de-treino/releases/tag/v0.16.0). A versão local `0.17.0` acrescenta o contexto de conclusão parcial, mas ainda não está publicada.

Nesta auditoria, a versão `0.12.0` acrescenta planejamento por proximidade do evento, sessões adaptadas iniciáveis, reforços de autenticação e entrada HTTP, serialização do resumo semanal e confirmação da tela de novidades somente após dispensa. A versão foi publicada no commit `61d7939` após validação e autorização explícita.

## ADR-011 — Contexto de conclusão do treino

**Status:** Validada localmente; ainda não publicada.

O encerramento de uma sessão deve registrar explicitamente `completion_status` como `complete` ou `partial`. Para `partial`, `partial_reason` é obrigatório e usa valores controlados. O contexto melhora a interpretação do realizado sem transformar uma sessão interrompida em evidência de tolerância ao treino completo.

A conclusão parcial não libera progressão no trigger do `rules-v1`; dor, fadiga alta e esforço elevado continuam podendo acionar suas reduções protetivas. O mesmo contexto é exposto no plano, no histórico e no bloco `planned-vs-actual-v1`, sempre em modo observacional. A migração `000020` usa `complete` como padrão para registros anteriores e foi aplicada e validada localmente com teste SQL e teste manual; ela não altera a produção sem backup, aplicação pelo perfil `maintenance` e autorização explícita.

## ADR-012 — Sinais adicionais do feedback pós-treino

**Status:** Validada localmente; ainda não publicada.

O feedback pode registrar `recovery_after` e `repeat_confidence` em escala de 1 a 5. Os campos são opcionais no banco para não reescrever históricos anteriores, mas são enviados pelo formulário atual com uma seleção neutra inicial. Eles aparecem no plano, no histórico e na comparação observacional da sessão.

Esses sinais não alteram duração, RPE, estímulo ou status, não liberam progressão e não substituem sono, estresse, fadiga ou o check-in diário. A migração `000021` adiciona apenas colunas e restrições de faixa; a validação manual local foi concluída e a produção permanece sem ela até backup, aplicação pelo perfil `maintenance` e autorização explícita.

Como continuação não visual, a cobertura desses campos é classificada por `post-workout-context-v1` dentro de `workouts.explanation.adaptation_shadow`. A avaliação informa dados observados, ausentes e inválidos, mas não cria limiares ou tendência e mantém `progression_eligible` e `used_for_prescription` como `false`. O `rules-v1`, o trigger e a prescrição permanecem inalterados; testes automatizados cobrem os estados completo, parcial, inválido e o isolamento do shadow.

## ADR-013 — Histórico permanente de novidades

**Status:** Validada localmente; ainda não publicada.

O aviso exibido no primeiro acesso após uma atualização deve comunicar somente as novidades da versão atual, com altura reduzida e acesso ao histórico completo. A rota autenticada `/novidades` reúne todas as notas em `frontend/lib/release.ts`, agrupadas por versão e recolhidas por padrão nas versões antigas. O acesso permanece disponível no menu lateral, no menu móvel e no cabeçalho das telas internas.

Essa separação preserva a exigência de informar mudanças logo após uma atualização sem transformar o modal em uma lista extensa. A confirmação continua sendo armazenada por conta, versão e navegador; abrir o histórico pelo aviso também encerra o aviso atual. A versão local foi atualizada para `0.19.0`, validada visualmente e registrada no commit `8798ea3`; produção permanece em `0.16.0` até backup, deploy e autorização explícita.

## ADR-014 — Auditoria da avaliação shadow de adaptação

**Status:** Implementada localmente; ainda não publicada.

Toda avaliação `rules-v2-adaptation-v1` deve expor sua proveniência em `adaptation-audit-v1`: dados considerados, lacunas, restrições aplicadas, caminhos não selecionados e condições informativas para uma eventual revisão. A confiança fica explicitamente como `not_calibrated` até existir base de dados e calibração suficientes.

O bloco é observacional e não é um novo motor de prescrição. `rules-v1`, o trigger e o comportamento das sessões permanecem inalterados; `used_for_prescription` é sempre `false`. A decisão evita esconder limitações atrás de uma resposta genérica e cria uma superfície estável para auditoria antes de qualquer adaptação em ciclo fechado.

## ADR-015 — Exclusão de sessões inelegíveis das métricas observacionais

**Status:** Implementada localmente; ainda não publicada.

O resultado `data-integrity-v1` é salvo junto do treino concluído em `workouts.explanation.data_integrity`. Quando `eligible_for_history` é explicitamente `false`, a sessão não deve alimentar o resumo observado, as janelas de carga, os períodos usados pelo shadow ou os agregados da tela de Evolução. Isso evita que um registro já classificado como incompleto ou inconsistente altere dor, fadiga, RPE acima do alvo, recência ou carga observados.

A regra é aplicada somente às métricas de sessões realizadas. A aderência planejada continua baseada nos estados dos treinos e não é apagada por uma inconsistência de medição. Registros legados sem `data_integrity` permanecem aceitos para preservar históricos anteriores; a ausência desse bloco não é convertida retroativamente em validade comprovada. Nenhuma prescrição nova é ativada: `rules-v1` continua sendo a única autoridade e não há migração de banco.

## ADR-016 — Limite temporal único para o resumo observado

**Status:** Implementada localmente; ainda não publicada.

O resumo observado de 28 dias usado na construção do contexto de prontidão deve considerar somente sessões com `completed_at` entre `now() - interval '28 days'` e `now()`. O limite superior impede que registros futuros, criados por correção de relógio, fixture ou inconsistência operacional, contaminem minutos, RPE, fadiga, dor ou cobertura.

Essa consulta passa a seguir a mesma referência temporal já usada nas janelas cumulativas, nos seis períodos não sobrepostos e na qualidade temporal do histórico. O filtro não altera a aderência planejada, não apaga registros e não reclassifica sessões; apenas mantém a observação coerente. `rules-v1` continua como autoridade prescritiva, sem ativação de regra nova, migração ou mudança visual.

## ADR-017 — Limite temporal dos agregados de Evolução

**Status:** Implementada localmente; ainda não publicada.

Os agregados observacionais da Evolução devem ignorar eventos futuros: sessões concluídas usam `completed_at <= now()`, sessões canceladas usam `cancelled_at <= now()` e check-ins usam `recorded_on <= CURRENT_DATE`. A barreira é aplicada ao resumo total, à série semanal, às sessões recentes e aos pontos de recuperação.

Isso evita que correções de relógio, fixtures ou inconsistências operacionais apresentem atividade futura como histórico já realizado. A separação de atleta e a exclusão por `eligible_for_history: false` permanecem ativas. A decisão não apaga dados, não altera a aderência planejada e não autoriza qualquer prescrição nova; `rules-v1` continua como único motor ativo.

## ADR-018 — Gate de tolerância integrado ao shadow de adaptação

**Status:** Implementada localmente; ainda não publicada.

Quando `load-tolerance-v1` classifica um sinal protetivo recente, o `rules-v2-adaptation-v1` deve bloquear sua própria candidata de progressão e registrar `prefer_recovery`. A resposta fácil da sessão atual não pode neutralizar um esforço recente acima do alvo, dor, fadiga alta ou necessidade de recuperação observada.

Essa decisão fecha a integração entre o avaliador de tolerância e o shadow sem ativar prescrição: `rules-v1` e o trigger continuam responsáveis pela adaptação real, enquanto `progression_eligible`, `applied` e `used_for_prescription` permanecem falsos. O motivo específico e a regra avaliada ficam disponíveis para auditoria e regressão.

## ADR-019 — Conclusão parcial bloqueia progressão observacional

**Status:** Implementada localmente; ainda não publicada.

Uma sessão marcada como `partial` não representa tolerância ao treino completo. O `rules-v2-adaptation-v1` deve priorizar sinais protetivos de dor, fadiga ou esforço alto e, quando eles não existirem, registrar `not_evaluated`/`defer_progression` com o motivo `partial_completion`.

A decisão torna explícita a coerência entre o contexto de conclusão, a integridade e os gates de carga sem interpretar o motivo da interrupção como diagnóstico. O `rules-v1`, o trigger SQL e a prescrição ativa permanecem inalterados; o resultado continua com `progression_eligible: false`, `applied: false` e `used_for_prescription: false`. A auditoria também identifica o `load_tolerance_gate` quando a tolerância observada está protetiva.

## ADR-020 — Proveniência da auditoria deve refletir dados válidos

**Status:** Implementada localmente; ainda não publicada.

O `adaptation-audit-v1` deve distinguir dados realmente considerados de campos inválidos ou ausentes. As lacunas dos blocos observacionais aninhados são consolidadas na auditoria, enquanto os campos básicos só aparecem em `data_used` quando o feedback e o RPE passam pela validação mínima. Os períodos históricos avaliados são identificados separadamente.

Essa decisão evita uma explicação que pareça mais completa do que os dados permitem. Ela não altera a classificação fisiológica, não ativa qualquer candidato do `rules-v2`, não modifica `rules-v1` ou o trigger e mantém a auditoria sem autoridade prescritiva.

## ADR-021 — Progressão shadow exige recuperação em cada período

**Status:** Implementada localmente; ainda não publicada.

O candidato de progressão do `rules-v2-adaptation-v1` deve respeitar a mesma cobertura de recuperação definida por `load-tolerance-v1`: cada um dos dois períodos recentes precisa ter carga por session-RPE, feedback completo e ao menos um check-in de recuperação completo.

Essa decisão evita que um check-in antigo cubra uma lacuna no período mais recente. Quando a avaliação de tolerância não está completa, a candidata fica adiada e a lacuna é registrada; nenhum resultado observacional altera `rules-v1`, o trigger ou a prescrição ativa.

## ADR-022 — Treino perdido bloqueia progressão shadow

**Status:** Implementada localmente; ainda não publicada.

Uma candidata de progressão não deve ser considerada quando os períodos de evidência contêm sessões previstas perdidas ou treinos em andamento vencidos. O `rules-v2-adaptation-v1` registra `low_adherence` e adia a candidata, mantendo cancelamentos explícitos como estado separado.

A decisão usa apenas estados factuais do histórico planejado; não presume o motivo da ausência, não cria compensação ou reagendamento e não altera a prescrição ativa. `rules-v1`, o trigger e o banco permanecem inalterados.

## ADR-023 — Precedência explícita dos gates de adaptação shadow

**Status:** Implementada localmente; ainda não publicada.

Na avaliação `rules-v2-adaptation-v1`, sinais protetivos devem prevalecer sobre conclusão parcial. Quando não há proteção, uma candidata de progressão só pode ser registrada depois que a tolerância à carga e os dois períodos de evidência estejam completos. Se a tolerância estiver incompleta, o `decision_audit` registra `load_tolerance_gate` junto da candidata adiada.

Essa decisão separa três situações que não devem ser confundidas: proteção observada, evidência insuficiente e candidata observacional. Nenhuma delas altera `rules-v1`, a sessão seguinte, `progression_eligible`, `applied` ou `used_for_prescription`. A regressão combinada permanece como barreira antes de qualquer integração prescritiva.

## ADR-024 — Auditoria finalizada após a transação shadow

**Status:** Implementada localmente; ainda não publicada.

O `decision_audit` deve ser reconstruído no fim do fluxo de conclusão do treino, depois que o repositório anexar `planned_vs_actual` e depois que registrar o resultado de uma consulta histórica protegida por savepoint. Isso garante que a auditoria consolidada represente o mesmo estado que será serializado em `workouts.explanation`.

Uma falha recuperável do histórico não pode ser listada como dado utilizado. Ela deve aparecer como `history_query_failed`, com `history_query_gate`, enquanto o feedback principal continua seguindo o fluxo transacional definido. A regra continua observacional e não altera a prescrição ativa.

## ADR-025 — Periodização observada sem autoridade prescritiva

**Status:** Implementada localmente; ainda não publicada.

Cada rascunho pode carregar uma auditoria `periodization-shadow-v1` da estrutura planejada de quatro semanas. Ela observa a sequência de três semanas de progressão e uma semana de recuperação, o volume planejado, a distribuição de sessões de qualidade, a alta intensidade, a recuperação, os treinos longos, o taper e o espaçamento entre estímulos exigentes.

O bloco verifica incoerências estruturais e preserva lacunas sem escolher, substituir ou ajustar sessões. Ele permanece em `mode: shadow`, com `progression_eligible: false`, `applied: false` e `used_for_prescription: false`; `rules-v1` continua sendo a única autoridade de prescrição. A decisão não cria migração, mudança visual ou nova nota de versão e exige revisão do contrato e validação antes de eventual publicação.

## ADR-026 — Seleção de estímulos observada sem autoridade prescritiva

**Status:** Publicada no commit `c5d8822`, sem mudança de versão.

Cada rascunho pode carregar `stimulus-selection-shadow-v1`, que compara a necessidade inferida do contexto atual com as famílias de estímulos selecionadas pelo `rules-v1`. A leitura considera proteção/recuperação, aderência e base, especificidade de evento e progressão de qualidade, preservando as lacunas quando o contexto não permite uma avaliação segura.

Um desencontro é registrado como observação e não dispara troca de treino, ajuste de duração, mudança de RPE ou progressão. O bloco permanece em `mode: shadow`, com `progression_eligible: false`, `applied: false` e `used_for_prescription: false`. Essa separação permite avaliar a futura seleção inteligente sem substituir o motor determinístico antes de haver calibração e efeito longitudinal demonstrados.

## ADR-027 — Contexto de segurança detalhado sem diagnóstico

**Status:** Implementada localmente; ainda não publicada.

O perfil pode registrar localização, intensidade percebida, movimento agravante e data de início de uma limitação, além dos campos já existentes. Esses dados são opcionais, possuem limites de formato e tamanho e não aceitam data futura. O texto da interface informa que o registro não é diagnóstico e mantém a recomendação de avaliação profissional quando aplicável.

Os detalhes servem para contexto informado pelo atleta e auditabilidade do cadastro. O fluxo de planejamento consulta somente o tipo da limitação e a recomendação de liberação profissional; localização, intensidade e movimento agravante não entram no `prescription_snapshot` nem autorizam interpretação clínica. `rules-v1` continua sendo a única autoridade prescritiva, e os shadows permanecem não autoritativos. A migração `000022` é aditiva e preserva registros anteriores; a versão local `0.21.0` é comunicada pela tela de novidades, foi validada e está no commit `9514a00`, mas ainda não foi publicada.

## Estado de produção

Em 3 de setembro de 2026, o commit `57c241a` foi implantado na VPS Oracle. A imagem da API, do job de digest e do frontend foi reconstruída com Go 1.25; as migrações `000013` e `000014` foram aplicadas após backup preventivo. O serviço Ollama permanece isolado e parado após a medição de capacidade, e a API usa temporariamente o Worker remoto com `AI_ENABLED=true` e `AI_PROVIDER=worker`.

Ainda em 3 de setembro, o commit `33de28a` foi publicado por fast-forward na mesma VPS para corrigir a legibilidade dos períodos e da barra de rolagem dos gráficos da Evolução em telas pequenas. Somente `cadencia-frontend-1` foi reconstruído e recriado; API, PostgreSQL e túnel permaneceram ativos e saudáveis. As rotas públicas principal e `/evolucao` retornaram HTTP 200. O backup preventivo correspondente foi `cadencia-20260904T023630Z.dump` (timestamp em UTC).

Em 4 de setembro de 2026, o commit `5fbc668` foi publicado por fast-forward na mesma VPS. O backup preventivo `cadencia-20260905T003553Z.dump` (UTC) foi criado e verificado, a migração `000015` foi aplicada pelo perfil `maintenance` e as imagens da API e do frontend foram reconstruídas. Os containers de API e frontend foram recriados; PostgreSQL e túnel permaneceram ativos e saudáveis. A API interna respondeu `{"status":"ready"}` e os dois domínios públicos retornaram HTTP 200.

Na sequência, o commit `c768ef7` atualizou somente o frontend para publicar a versão `0.7.0` e a nota do catálogo. A nota apareceu no primeiro acesso autenticado de teste; o fluxo funcional do check-in de recuperação e os testes de latência, limites e fallback do Worker já haviam sido validados.

Em 5 de setembro de 2026, o commit `9d8c624` foi publicado por fast-forward na VPS Oracle. O backup `cadencia-20260905T221541Z.dump` foi criado e verificado, a migração `000016` foi aplicada pelo perfil `maintenance` e as imagens da API e do frontend foram reconstruídas. API e frontend ficaram saudáveis; PostgreSQL e túnel permaneceram ativos. A release `v0.8.0` foi publicada no GitHub.

Em 11 de setembro de 2026, o commit `61d7939` foi publicado por fast-forward na VPS Oracle. O backup preventivo `cadencia-20260911T234851Z.dump` foi criado e verificado; não houve migração nova. As imagens da API e do frontend foram reconstruídas, os serviços de aplicação e o Tunnel foram recriados e o PostgreSQL permaneceu saudável. A API interna respondeu `{"status":"ready"}`, os dois domínios públicos retornaram HTTP 200 e a release `v0.12.0` foi publicada no GitHub.

Na sequência, o commit `53cbadc` foi publicado como atualização somente do frontend para incluir o acesso ao perfil em telas pequenas. A API, o PostgreSQL e o Tunnel não foram recriados. A versão permaneceu `0.12.0`, conforme decisão de não abrir nova nota para essa correção de usabilidade.

Em 12 de setembro de 2026, o commit `6fdbe45` foi publicado por fast-forward na VPS Oracle. O backup preventivo `cadencia-20260912T155746Z.dump` foi criado e verificado; as migrações `000017`, `000018` e `000019` foram aplicadas pelo perfil `maintenance`; e somente API e frontend foram reconstruídos e recriados. PostgreSQL e Tunnel permaneceram ativos. A API respondeu `/health` e `/ready`, os dois domínios oficiais retornaram HTTP 200 e os quatro serviços ficaram saudáveis. A versão do produto passou para `0.16.0` e foi registrada na release [v0.16.0](https://github.com/Saulorangel87/App-de-treino/releases/tag/v0.16.0).

Em 11 de setembro de 2026, a pesquisa do próximo protocolo do catálogo recomendou especificar um taper pré-prova orientado por evento. A implementação local agora atua no plano com `taper-v1`: evento futuro entre 7 e 21 dias, atleta avançado, avaliação apta, base mínima de treino e ausência de proteção; sessões anteriores ao evento e dentro de 14 dias recebem redução de 50% da duração, sem alterar frequência, RPE ou a semana de recuperação. O `rules-v1` continua como motor prescritivo e não há sobrecarga automática antes do taper. A migração `000017` e a versão local `0.13.0` foram validadas localmente; ainda aguardam ampliação do catálogo e autorização antes de qualquer publicação.

O destino oficial de produção é a composição Docker na VPS Oracle, exposta pelos hostnames `cadencia.devsaulo.com.br` e `cadencia-api.devsaulo.com.br` no Cloudflare Tunnel dedicado. Uma publicação privada acidental no Sites, feita durante uma tentativa de deploy, foi excluída pelo proprietário. O Sites não é um destino autorizado para futuras publicações do Cadência.

Na mesma data, a versão `35f24685` do Worker ajustou o provedor Groq para `max_completion_tokens: 512` e `reasoning_effort: 'low'`. O Worker rejeita respostas com `finish_reason` diferente de `stop`, mantendo o fallback determinístico como proteção contra truncamento. A sessão autenticada confirmou explicações completas para treinos de base e subidas.

Validações realizadas:

- API interna `/health`: `200`.
- API interna `/ready`: `200`.
- API pública e frontend público: `200`.
- API, frontend e PostgreSQL saudáveis; túnel ativo.
- PostgreSQL sem porta publicada pelo Cadência.
- Dependabot do GitHub: 0 alertas abertos e 39 fechados.
- `go test ./...`, build Docker, build do frontend e auditoria de dependências de produção concluídos nesta atualização; `govulncheck` não está instalado no ambiente desta rodada.

## Backups

- `cadencia-backup.timer` está habilitado na VPS e executa diariamente às 03:30 UTC.
- Os dumps ficam em `/var/backups/cadencia`, com retenção de 14 dias.
- O script valida cada arquivo com `pg_restore --list`.
- O backup preventivo do último deploy funcional é `cadencia-20260911T234851Z.dump` (UTC).
- O backup anterior, `cadencia-20260902T104801Z.dump`, permanece registrado e validado.
- O teste completo de restauração foi concluído em 2 de setembro de 2026 com `cadencia-20260902T104801Z.dump`: a restauração em PostgreSQL 17 temporário terminou sem erro, validou 15 tabelas públicas e `cadencia_schema_migrations`, e o container temporário foi removido sem tocar a produção.

## Segurança operacional da VPS

A composição do Cadência está isolada, mas a VPS também hospeda outros aplicativos. A auditoria encontrou portas Docker publicadas para Rotas, Despesas, Estoque, n8n, Immich, Uptime Kuma, Tecboard e Nginx Proxy Manager.

O teste externo confirmou a porta SSH `22` e a porta `2283` do Immich; as demais portas testadas não responderam externamente. A política local de entrada ainda é permissiva (`accept`), não há `ufw` instalado e a cadeia `DOCKER-USER` está vazia.

A auditoria somente leitura de 2 de setembro de 2026 encontrou também Home Assistant na porta `8123` (processo Python em `/config`), `casaos-gateway` na `8888` e a publicação Docker do Immich na `2283`. O Nginx Proxy Manager encaminha `casaos.oraclecloud.com.br` para `100.67.151.30:8888` e `immich.photo.com.br` para `100.67.151.30:2283`, usando Tailscale. O container `cloudflared-tunnel` é separado do túnel dedicado do Cadência e possui o Immich como origem; seus logs indicaram reconexão normal e uma versão desatualizada do cliente. Os testes externos confirmaram inicialmente apenas as portas `22` e `2283` acessíveis. Uma regra persistente na cadeia `DOCKER-USER` passou a bloquear `tcp/2283` somente pela interface pública `enp0s6`; a porta continua acessível por Tailscale e loopback, e o Immich permaneceu respondendo. Em seguida, novas conexões SSH foram bloqueadas na interface pública `enp0s6`, mantendo o acesso administrativo por Tailscale; as regras TCP públicas `22`, `81`, `2283`, `8096` e `8097` foram removidas da Oracle Cloud, restando somente ICMP.

Nenhuma porta de outro aplicativo deve ser bloqueada sem mapear antes seus domínios, túneis, proxies e necessidade de acesso. O próximo hardening deve preservar Tailscale, Cloudflare e os aplicativos existentes.

## Distribuição observacional de estímulos

A leitura de sessões exigentes foi adicionada como `stimulus-distribution-v1`, derivada dos seis períodos semanais já existentes. O agrupamento usa o RPE-alvo persistido (`>= 6,0` para qualidade e `>= 7,0` para alta intensidade), e a proximidade usa dias UTC de `completed_at`. Essa escolha é determinística, reproduzível e explícita sobre a ausência de fuso individual; não pretende representar zonas fisiológicas.

O bloco é calculado somente para observação e mantém `used_for_prescription: false`. Ele não altera o `rules-v1`, não modifica sessões, não cria um limite automático de carga e não transforma sessões em dias consecutivos em diagnóstico. A integração futura exigirá validação de cobertura, recuperação, feedback e efeito longitudinal, com o shadow preservado durante a revisão. Nenhuma migração ou alteração de infraestrutura é necessária porque os campos já existem em `workouts` e `workout_sessions`.

## Próximas decisões

1. Observar os primeiros relatos reais em `/feedback` e confirmar a entregabilidade/utilidade do resumo semanal pelo Resend.
2. Monitorar a explicação autenticada pelo Worker remoto, sua latência, limites e acionamento do fallback determinístico; manter o Ollama parado.
3. Comparar os resultados shadow em cenários controlados, preservando `rules-v1` e sem alterar carga; essa validação já foi concluída no commit `64e554d` e deve ser mantida como regressão.
4. Só então definir adaptação em ciclo fechado, progressão/carga e barreiras de integridade e segurança.
5. Evoluir o catálogo de protocolos de ciclismo com elegibilidade e evidências próprias; o catálogo inicial e o piloto de estrada já estão em produção, enquanto novos protocolos permanecem condicionados à revisão.
6. Definir cópia externa dos backups, monitoramento de falhas e alertas de saúde.
7. Escolher a política de firewall e a lista mínima de portas públicas da VPS, preservando Tailscale, Cloudflare e os demais aplicativos.

## Gate de distribuição dos estímulos em shadow — 13 de setembro de 2026

O Cadência passa a reutilizar `stimulus-distribution-v1` dentro dos avaliadores `rules-v2` e `rules-v2-adaptation-v1`, mas mantém a separação entre medição e prescrição. O gate observa duas ou mais sessões de qualidade em sete dias e sessões de qualidade em dias consecutivos nos 42 dias disponíveis. São critérios operacionais para detectar padrões que merecem revisão, não limiares fisiológicos universais.

Dados incompletos ou inconsistentes mantêm a candidata de progressão não avaliada/adiada e aparecem na auditoria; histórico completo sem estímulos de qualidade não é bloqueado. A distribuição é anexada ao resultado shadow, o gate é auditado e `rules-v1` continua sendo a única fonte capaz de alterar o plano. Não há migração, mudança visual, release ou deploy.

## Coerência integrada dos shadows — decisão publicada

`planning-coherence-shadow-v1` integra, para fins de auditoria, a estrutura de periodização do plano, a distribuição histórica de estímulos e a seleção de famílias de estímulos. O bloco registra estados e sinais resumidos, identifica divergências entre recuperação, fase e estímulo e mantém a decisão explicável sem duplicar os snapshots completos.

Esta integração não possui autoridade prescritiva: `rules-v1` continua gerando o plano, enquanto `progression_eligible`, `applied` e `used_for_prescription` permanecem falsos. A etapa foi validada localmente e publicada no commit `f0fec8b` após backup e validação operacional, sem migração, release ou mudança visual.
