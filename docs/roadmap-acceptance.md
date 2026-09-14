# Matriz histórica de aceitação técnica

Atualizada em 14 de setembro de 2026 para a versão publicada `0.30.0`.

Este documento preserva a matriz verificável que orientou a fase técnica anterior. O documento único de planejamento vigente é [`planejamento.md`](../planejamento.md). Esta matriz separa três situações que não podem ser confundidas:

- **implementado e verificado**: comportamento presente no checkout e coberto por código ou teste;
- **protegido, ainda não calibrado**: o app aplica a proteção conservadora, mas não afirma validade longitudinal ou clínica;
- **condição externa de saída**: não existe código honesto que possa substituir dados reais, revisão científica específica ou autorização de publicação.

`rules-v1` continua sendo o único motor prescritivo. `rules-v2` e os snapshots `shadow` são registros observacionais e não podem ganhar autoridade apenas para encerrar uma tarefa documental.

## 1. Experiência e prontidão atual

**Implementado e verificado.** A experiência continua no perfil, enquanto `readiness_assessment.state` registra dados insuficientes, cautela, recuperação, retorno, baixa consistência, preparação de evento ou estabilidade observada. O retorno após pausa limita a sessão a 45 minutos/RPE 3,5.

Evidência de teste: `readiness_test.go`, `TestBuildPlanUsesGradualReturnForLowRecentRegularity` e `TestBuildPlanDoesNotTreatRegularTrainingAsReturn`.

Condição externa: calibrar a relação entre esses estados e resultados individuais exige histórico real longitudinal; o estado não é diagnóstico nem libera progressão sozinho.

## 2. Motor de regras versionado

**Implementado e verificado.** `rules-v1` é determinístico e ativo; `rules-v2`, tolerância de carga, periodização, seleção de estímulo e adaptação registram versão, motivo, dados ausentes e `used_for_prescription: false` quando estão em shadow.

Evidência de teste: `rules_v2_test.go`, `rules_v2_adaptation_invariant_test.go` e `planning_coherence_shadow_test.go`.

Condição externa: qualquer transferência de autoridade para `rules-v2` requer comparação controlada com `rules-v1`, dados completos e revisão explícita da mudança de prescrição.

## 3. Base científica

**Implementado e verificado para as fontes integradas.** A migração `000026_scientific_source_metadata` guarda população, objetivo, estímulo, benefícios, limitações, riscos, contraindicações, confiança, data de revisão e regras relacionadas. O catálogo mantém links para as fontes e seus limites de transferência.

Evidência de teste: `database/tests/000026_scientific_source_metadata.sql` e metadados em `protocol_metadata.go`.

Condição externa: a revisão de confiança de cada nova fonte é trabalho científico contínuo; `not_calibrated` não pode ser convertido em certeza por código.

## 4. Biblioteca de treinos de ciclismo

**Implementado e verificado para os protocolos elegíveis.** Os protocolos ativos possuem estrutura, RPE, aquecimento, bloco principal, recuperação, desaquecimento, indicação, contraindicação, pré-requisitos, critérios de parada/progressão/regressão e fontes relacionadas. Incluem recuperação ativa, endurance, pedal longo, base, cadência/técnica, tempo, sweet spot, limiar, VO₂max, intervalos controlados, subidas, retorno, taper e recuperação pós-prova.

Evidência de teste: `TestProtocolMetadataIsCompleteForSelectableProtocols`, `post_event_recovery_test.go`, `taper_test.go` e `service_test.go`.

Decisões explícitas de escopo: descanso completo é um dia sem sessão planejada, e não uma sessão concluível artificialmente; estímulos máximos neuromusculares de sprint/pista/BMX e downhill/enduro estão fora do produto; resistência específica com força/carga mensurada permanece bloqueada até ter critério de medição e segurança próprio. Essas decisões estão em [`cycling-evidence-catalog.md`](cycling-evidence-catalog.md).

## 5. Diferenças por nível

**Implementado e verificado.** Iniciantes não recebem sessão de qualidade; intermediários evoluem por sessões controladas; protocolos avançados exigem gates adicionais. A experiência nunca vence retorno, dor, limitação, baixa recuperação ou baixa aderência.

Evidência de teste: `TestBuildPlanDoesNotAssignQualitySessionToBeginner`, cenários por nível em `readiness_test.go` e gates de protocolos avançados em `service_test.go`.

## 6. Adaptação após a sessão

**Protegido e parcialmente automático.** O `rules-v1` reduz carga ou faz progressão pequena somente com feedback válido. A migração `000027_adaptation_integrity_gate` bloqueia adaptação diante de sessão parcial, duração/RPE/fadiga inválidos ou sinal protetivo recente e preserva o motivo no treino adaptado.

Evidência de teste: `adaptation_test.go`, `database/tests/000005_adaptive_training_feedback.sql` e `database/tests/000027_adaptation_integrity_gate.sql`.

Condição externa: trocar estímulo, adiar, cancelar ou ampliar autoridade de adaptação exige efeito longitudinal demonstrado. O app não deve inventar essas ações por um único registro.

## 7. Carga e progressão

**Implementado como monitoramento conservador.** Carga por session-RPE, períodos de 7/28/42 dias, aderência e distribuição são calculados; sinais protetivos, baixa aderência e dados incompletos impedem progressão. ACWR não é aplicado automaticamente.

Evidência de teste: `history_test.go`, `load_tolerance_test.go`, `stimulus_distribution_gate_test.go` e `TestBuildPlanDoesNotAssignQualitySessionAfterLowAdherence`.

Condição externa: não existe progressão calibrada sem cobertura suficiente de carga, feedback e recuperação de usuários reais.

## 8. Periodização

**Implementado no ciclo atual e protegido.** O ciclo de quatro semanas progride volume de forma gradual, inclui semana de recuperação, limita qualidade nela, suporta taper e recuperação pós-prova. A auditoria `periodization-shadow-v1` verifica intensidade, densidade e conflitos sem substituir o motor ativo.

Evidência de teste: `periodization_shadow_test.go`, `taper_test.go`, `post_event_recovery_test.go` e `TestBuildPlanUsesActiveRecoveryInRecoveryWeek`.

Condição externa: fases de pico, transição e planejamento longo precisam de objetivo, calendário e efeito real observados; não serão simuladas como periodização individual validada.

## 9. Seleção por necessidade

**Implementado com precedência de segurança e auditoria.** O `rules-v1` já troca qualidade por proteção em retorno, dor, recuperação, limitação e baixa aderência. `stimulus-selection-shadow-v1` registra necessidade, família esperada e estímulos escolhidos, permitindo comparar o comportamento sem alterar a prescrição.

Evidência de teste: `stimulus_selection_shadow_test.go`, `planning_coherence_shadow_test.go` e testes de proteção em `service_test.go`.

Condição externa: a seleção plenamente adaptativa depende da validação do efeito das alternativas sobre adesão e recuperação; enquanto isso, a escolha permanece conservadora e explicável.

## 10. Segurança

**Implementado como coleta e proteção, não como diagnóstico.** Limitações, localização, intensidade, movimento agravante, início, sintomas de alerta, restrição médica e recomendação de avaliação profissional são armazenados. Limitação ativa bloqueia início de sessão acima de RPE 4, sem bypass, e a geração protege o plano.

Evidência de teste: `TestWorkoutRequiresSafetyBlockOnlyForIntenseSessionWithLimitation`, `TestGenerateCapsIntensityWhenLimitationExists` e `database/tests/000025_limitation_safety_signals.sql`.

Condição externa: a classificação clínica de sintomas, diagnóstico e liberação para esforço exigem profissional habilitado; o app não tentará automatizá-los.

## 11. Integridade e correção de dados

**Implementado e verificado.** O gate rejeita métricas fora de faixa, duração/distância incompatíveis, feedback parcial sem motivo e valores não finitos. Registros inelegíveis ficam fora da observação; métricas opcionais podem ser corrigidas com histórico preservado e sem recalcular a prescrição.

Evidência de teste: `data_integrity_test.go`, `execution_comparison_test.go`, `TestCorrectWorkoutValidatesAndDelegates` e fixtures `000024`–`000029`.

## 12. Feedback pós-treino

**Implementado e verificado.** A sessão registra conclusão, motivo parcial, RPE, dificuldade, fadiga, dor, recuperação, confiança, satisfação, terreno, condições externas, equipamento e métricas opcionais. O `rules-v1` usa sinais válidos de proteção; os demais dados permanecem explícitos em observação até calibração.

Evidência de teste: `post_workout_context_test.go`, `execution_comparison_test.go`, `adaptation_audit_test.go` e `planning_serialization_test.go`.

## 13. Decisões auditáveis

**Implementado e verificado.** Cada sessão gerada recebe `workout-decision-audit-v1`, com regras avaliadas/aplicadas, dados usados/ausentes, restrições, alternativas descartadas, condições de mudança e confiança. A tela do plano mostra essas informações em “Ver detalhes da decisão”, além das regras e fontes.

Evidência de teste: `TestGenerateAttachesDecisionAuditToEveryWorkout`, `workout_decision_audit_test.go` e contrato em `api/openapi.yaml`.

## 14. Escopo de ciclismo

**Implementado e verificado.** O domínio aceita somente ciclismo de estrada, MTB XCO/XCM, gravel e indoor. Corrida, musculação, sprint/pista/BMX e downhill/enduro não recebem telas, perfis, regras ou protocolos.

Evidência de teste: validações de disciplina e catálogo em `service_test.go`; decisão registrada no catálogo científico.

## 15. Testes obrigatórios

**Implementado para os cenários automatizáveis.** A suíte cobre iniciante/intermediário/avançado, retorno, recuperação, dor, baixa aderência, evento, ausência de sensores, dados inconsistentes, parcial, aumento/manutenção/redução, semana de recuperação, densidade de qualidade, fadiga, progressão fácil e regressão por esforço alto. A camada SQL testa geração/adaptação transacional; as suítes Go testam serialização e preservação de histórico.

Comandos de verificação da candidata: `go test -count=1 ./...`, `go vet ./...`, fixtures SQL transacionais `000005`, `000020`, `000024`–`000029`, lint direcionado dos arquivos alterados, `npm run build` e `git diff --check`.

Limite conhecido: o lint geral possui dívida histórica fora desta entrega; o lint direcionado dos arquivos modificados deve permanecer limpo.

## 16. Critérios operacionais de aceitação

**Atendidos e verificados em produção.** O Cadência preserva o PostgreSQL privado, não cria modalidades fora do escopo, mantém o motor explicável e conservador e foi publicado na versão `0.30.0`.

O deploy foi concluído com backup verificável, aplicação ordenada das migrações `000024`–`000029`, confirmação de `cadencia_schema_migrations`, leitura autenticada de `/v1/plans/current`, validação da interface e healthchecks. A regra de compatibilidade de schema permanece obrigatória nos próximos deploys; healthcheck isolado não é suficiente.

## Fechamento técnico desta fase

Não há tarefa de implementação do MVP pendente. As atividades restantes são manutenção operacional, observação com dados reais e revisão científica/clínica contínua; elas não são itens de código esquecidos nem reabrem o escopo de corrida ou musculação.
