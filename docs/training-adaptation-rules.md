# Adaptação do plano após o treino

Última revisão: 4 de setembro de 2026.

## Objetivo

O aplicativo usa o feedback enviado ao concluir uma sessão para fazer ajustes pequenos nas próximas sessões ainda planejadas. A adaptação acontece na mesma transação que salva o feedback no PostgreSQL: se qualquer parte falhar, nenhuma alteração parcial é mantida.

O mecanismo é uma regra conservadora de produto, não um diagnóstico nem uma prescrição médica individualizada. RPE, dificuldade e fadiga são sinais subjetivos e devem ser interpretados junto com o contexto do atleta.

## Regras da versão 1

- **Dor relatada:** reduz em 20% a duração das duas próximas sessões e limita o RPE-alvo a 3. A interface recebe também um aviso para interromper a atividade se a dor reaparecer e buscar avaliação profissional se ela for persistente ou intensa.
- **Esforço ou fadiga muito altos:** quando o RPE real é pelo menos 9, a fadiga é 5 ou a dificuldade é `very_hard`, reduz em 20% a duração das duas próximas sessões, diminui o RPE-alvo em 1 e o limita a 4.
- **Carga acima do esperado:** quando o RPE real supera o alvo em pelo menos 2, a fadiga é 4 ou a dificuldade é `hard`, reduz em 10% a duração da próxima sessão e diminui o RPE-alvo em 1.
- **Resposta claramente fácil:** quando o RPE real fica pelo menos 2 pontos abaixo do alvo, a fadiga é no máximo 2 e a dificuldade é `easy` ou `very_easy`, aumenta somente 5% da duração da próxima sessão. A intensidade não aumenta.
- **Resposta dentro do esperado:** mantém o plano inalterado.

## Seleção de sessões específicas

Na geração do plano, o motor mantém a frequência, os limites de duração, a semana de recuperação e as proteções de segurança já definidas. O contexto opcional do ciclista serve somente para escolher a sessão de qualidade mais adequada:

- **Indoor + nível intermediário:** sessão de cadência técnica com RPE moderado.
- **Terreno com subidas + nível intermediário ou avançado:** subidas controladas, com recuperação leve.
- **Nível avançado + medidor de potência e FTP informado:** sweet spot orientado pelo dado disponível, sem converter automaticamente o FTP em metas rígidas de watts nesta versão.
- **Nível avançado + meta de prova:** ritmo de prova controlado, sustentável e sem tentar reproduzir a prova completa.
- **Demais perfis:** mantém o tempo controlado; avançados sem contexto específico recebem sweet spot progressivo.

Quando o ciclista informa preferências de sessão, elas orientam a escolha da sessão de qualidade dentro das mesmas proteções: cadência é elegível para intermediários e avançados; subidas exigem terreno com subidas; sweet spot exige nível avançado e, para potência, FTP informado; intervalos continuam exigindo avaliação submáxima apta, objetivo compatível e semana de construção. Se todas as opções forem marcadas, o motor interpreta isso como abertura a qualquer protocolo e mantém a seleção contextual padrão. Giro/base e recuperação permanecem preferências registradas, sem transformar todos os dias em sessões de qualidade.

Se houver limitação ativa, a sessão específica é substituída pelo giro leve protegido. Iniciantes não recebem essas sessões de qualidade específicas ainda. Intervalos de alta intensidade ("tiros") continuam fora do motor até existir uma avaliação inicial de capacidade e regras próprias de progressão.

## Estrutura operacional das sessões

As sessões geradas agora carregam uma sequência de etapas acionáveis em `workout.structure.steps`, além dos campos antigos de aquecimento, parte principal e desaquecimento. Cada etapa informa ordem, tipo, duração, RPE-alvo e uma instrução curta para execução.

Uma sessão de subida, por exemplo, pode apresentar quatro esforços separados por recuperações leves, em vez de apenas o texto "4 blocos sustentados em subida". A soma das etapas é igual à duração prescrita e continua limitada à disponibilidade do dia. Em uma adaptação de duração, a interface ajusta proporcionalmente a exibição das etapas até que uma futura versão do motor passe a recalcular os blocos no próprio banco.

## Biblioteca de protocolos

O motor mantém uma biblioteca explícita de protocolos em código. Cada protocolo possui uma chave estável, um formato de blocos, uma instrução de execução e referências de evidência. A duração final ainda é calculada pelo motor conforme a semana, o nível, os minutos disponíveis e as proteções de segurança.

- `base_endurance`: giro de base contínuo, RPE leve a moderado.
- `continuous_endurance`: endurance contínuo no maior período disponível, sem blocos intensos.
- `controlled_tempo`: três blocos de ritmo controlado com recuperação leve.
- `controlled_hills`: até quatro blocos sustentados em subida, com recuperação leve.
- `technical_cadence`: blocos curtos de cadência técnica, sem elevar excessivamente o esforço.
- `power_sweet_spot` e `progressive_sweet_spot`: blocos sustentáveis; o FTP contextualiza a sessão, mas não cria metas rígidas automaticamente.
- `controlled_event_pace`: blocos sustentáveis orientados à meta de prova, sem simular a prova completa.
- `controlled_intervals`: quatro blocos de 4 minutos com 3 minutos leves, liberados somente para o perfil avançado elegível.
- `protected_recovery`: giro leve protegido quando existe uma limitação ativa.

As referências associadas sustentam princípios como progressão gradual, monitoramento de carga e uso contextual de intervalos. Elas não devem ser interpretadas como validação de um número universal de minutos para todas as pessoas; os parâmetros continuam sujeitos às regras de segurança do produto.

## Avaliação inicial submáxima

A rota `/avaliacao` apresenta um pedal de referência opcional: aquecimento leve, até 20 minutos de esforço controlado próximo de RPE 5 e desaquecimento. O atleta registra duração, RPE percebido e dor. Não há teste máximo, estimativa de VO₂max nem diagnóstico.

Uma referência com pelo menos 18 minutos, RPE até 6 e sem dor fica marcada como apta. Para atletas avançados, com objetivo `performance` ou `event`, sem limitação ativa e ao menos 50 minutos disponíveis, ela libera somente duas sessões de **intervalos controlados** nas semanas de construção: 4 blocos de 4 minutos em RPE 7, com 3 minutos leves entre blocos. O plano não usa sprints, não ultrapassa o RPE avançado já existente e não adiciona sessões extras. Dor, tontura, falta de ar incomum, mal-estar ou outro sintoma preocupante são motivos para interromper a atividade e buscar orientação profissional quando necessário.

As alterações nunca ultrapassam os minutos disponíveis cadastrados para o dia. Sessões concluídas, puladas ou já modificadas por um feedback anterior não são recalculadas.

## Check-in diário antes do treino

A rota `/recuperacao` registra duração e qualidade do sono, estresse e fadiga percebida. Os limiares são regras conservadoras desta versão do produto, não critérios clínicos:

- menos de 6 horas de sono ou qualidade do sono 1 ou 2 formam um sinal de sono; estresse 4 ou 5 e fadiga 4 ou 5 são os outros sinais de atenção;
- um único sinal de atenção reduz em 10% a duração da próxima sessão futura e limita o RPE-alvo a 5;
- dois ou mais sinais de atenção, ou fadiga 5, reduzem em 20% a duração e limitam o RPE-alvo a 4;
- sem sinal de atenção, o plano é mantido; o check-in nunca aumenta intensidade ou duração por si só.

Somente a próxima sessão `planned` ou `adapted` do plano ativo, na data do check-in ou depois dela, pode ser alterada. A decisão é transacional e fica em `workouts.explanation.pre_session_recovery`, incluindo a data, o nível de prontidão e o histórico de datas já aplicadas. Esse histórico impede que salvar novamente o mesmo check-in reduza repetidamente qualquer sessão. A duração mínima após o ajuste é 20 minutos.

Depois que uma redução é aplicada, editar o check-in não aumenta novamente a sessão de forma automática. Essa escolha evita que uma correção de formulário seja interpretada como autorização para progredir carga; um ajuste manual futuro deverá ser uma ação explícita e auditável.

## Histórico observado na geração do ciclo

Ao gerar um novo rascunho, o motor agrega os últimos 28 dias de sessões concluídas e check-ins de recuperação. São considerados minutos realizados, RPE médio, fadiga média, dor relatada, quantidade de check-ins e fadiga média informada nos check-ins. O resumo é salvo no `prescription_snapshot.observed_training` para permitir auditoria da decisão.

O uso é deliberadamente conservador: dor relatada protege todas as sessões do novo ciclo com um giro leve; fadiga média igual ou superior a 4, ou fadiga média dos check-ins igual ou superior a 4, protege a sessão de qualidade. A proteção limita o alvo a RPE 3,5, reduz a duração e mantém o treino dentro da disponibilidade cadastrada. O histórico não aumenta intensidade, não substitui a avaliação submáxima e não representa diagnóstico clínico.

## Leitura observacional de prontidão (`readiness-v1`)

Implementação local de 5 de setembro de 2026, ainda não publicada. Ao gerar um rascunho, o backend registra `prescription_snapshot.readiness_assessment` separadamente de `engine_version: rules-v1`. É uma descrição versionada dos dados disponíveis, não uma nova prescrição, diagnóstico ou leitura do estado de hoje. O horário UTC da classificação fica em `assessed_at`; ela não é recalculada ao abrir um plano antigo.

Ordem determinística das decisões:

1. Limitação ativa, dor relatada ou fadiga média pós-treino/check-in entre 4 e 5: `recovery_needed`, mesmo quando há outras lacunas. Reutiliza os sinais e limiares de proteção do motor existente; não implica que a dor histórica persiste hoje.
2. Janela diferente de 28 dias, agregados inconsistentes, cobertura desconhecida ou nenhum registro utilizável: `insufficient_data`.
3. Algum registro utilizável, mas faltam sessões, check-ins ou campos: `caution` por dados parciais, não diagnóstico de recuperação ruim.
4. Sem os alertas acima e com cobertura dos campos observados: `stable`, entendido estritamente como **sem alertas nos agregados disponíveis**. Não significa que médias capturem picos de fadiga, sono ruim recente ou tolerância a uma carga maior.

Cobertura (`data_coverage`): sessões com duração positiva; RPE entre 1 e 10; feedback com fadiga entre 1 e 5 e resposta de dor; sessões que reúnem todos esses campos; check-ins com fadiga entre 1 e 5. Zero minuto, RPE zero e valores ausentes não contam como registro completo. São verificações de integridade do produto, não novos limiares fisiológicos. O número de check-ins completos aqui se refere apenas à fadiga; sono e estresse não estão sendo agregados nesta leitura.

`missing_data` registra lacunas concretas, enquanto `not_evaluated` declara os fatores ainda fora do cálculo: aderência, tolerância à carga, destreinamento, tendências 7/28/42 dias, sono/estresse/fadiga recentes, recência da avaliação e variações dentro da janela. Experiência e avaliação submáxima continuam no contexto do motor antigo, mas não são atalhos para esta classificação. Poucos registros podem ser uma conta nova ou pedais fora do app: por isso não se emite `low_consistency` nem `ready_for_progression`.

`progression_eligible: false` significa que **este classificador não autoriza progressão**. Não interfere na elegibilidade nem nos treinos do `rules-v1`. Nenhuma duração, intensidade, bloco, proteção, adaptação por feedback ou check-in foi alterada. Os agregados legados foram mantidos, inclusive sua definição temporal; recência, registros futuros e adequação das janelas serão tratados na próxima fatia antes de usar esta leitura para prescrever.

Testes automatizados em `backend/internal/planning/readiness_test.go`: contas novas, cobertura parcial/desconhecida, somente check-ins, duração zero, valores impossíveis/NaN/infinito, prioridades de dor/fadiga, limites existentes, independência de experiência/avaliação, determinismo, cópia dos dados, serialização JSON e invariância do plano quando muda apenas a cobertura. `go test -count=1 ./...` e `go vet ./...` passaram. O teste `pwsh -NoProfile -File scripts/test-readiness-queries.ps1` extrai as consultas do repositório e as executa no PostgreSQL local, com dados fictícios em CTEs e transação somente leitura: passaram histórico completo, campos incompletos, ausência de histórico e somente check-ins. Também verifica exclusão de sessões canceladas, antigas e de outro atleta, sem consultar ou alterar registros reais.

Essa conferência foi concluída pelo proprietário: após gerar o rascunho, atualizar a tela e consultar `GET /v1/plans/current`, permaneceram iguais o ID do plano, `assessed_at`, a classificação, os motivos, as lacunas e os treinos. A prontidão observacional está persistindo no snapshot como previsto.

## Histórico de aderência e carga em 7/28/42 dias (`training-history-v1` e `v2`)

Implementação local de 5 de setembro de 2026, ainda não publicada. Novos rascunhos congelam `prescription_snapshot.training_history` em modo `observation`; planos existentes não são recalculados. O campo é adjacente a `readiness_assessment` e mantém `used_for_prescription: false`, portanto não altera duração, RPE, estrutura ou escolha de sessão do `rules-v1`.

Cada janela cumulativa de 7, 28 e 42 dias possui dois eixos:

- aderência pelo `scheduled_on`: considera planos ativos ou concluídos e somente sessões cuja oportunidade já foi fechada. Uma sessão de hoje entra se estiver concluída ou cancelada; `planned`/`adapted` de hoje não prejudica a taxa. Em datas anteriores, `planned`/`adapted` contam como pendência vencida e `in_progress` fica identificado como em andamento vencido;
- carga realizada pelo `completed_at`: considera sessões concluídas do atleta entre o início da janela móvel e o relógio atual do banco. Sessão futura é excluída. Duração nula/zero ou RPE fora de 1 a 10 impedem apenas o cálculo de carga daquela sessão e aumentam `sessions_without_session_rpe_load`.

A taxa de conclusão é `scheduled_completed_sessions / expected_sessions × 100`; quando não existe sessão esperada, retorna `null`, não zero. `session_rpe_load` é a soma de `duração em minutos × RPE real`, em unidades arbitrárias. A base científica é o uso de session-RPE para monitoramento de carga interna (`foster-2001`, `haddad-2017`, `impellizzeri-2020`), com as limitações já documentadas: essa medida não representa sozinha resposta fisiológica, risco, recuperação ou prontidão para progredir.

As janelas se sobrepõem e não são usadas como ACWR. Não existe nesta versão limiar automático para boa/baixa aderência, queda importante de volume, destreinamento ou tolerância. Pedais externos, atividades não iniciadas pelo Cadência e motivos da não conclusão não estão disponíveis; por isso a medição descreve o uso registrado no app e não deve ser apresentada como julgamento do atleta. A data planejada segue `CURRENT_DATE` do PostgreSQL e a janela realizada segue `now()` do banco; um fuso por atleta ainda não existe e deverá ser definido antes de decisões sensíveis à virada do dia.

Testes: `backend/internal/planning/history_test.go` cobre ordenação, taxas, ausência de denominador, cobertura incompleta, entradas inconsistentes, determinismo, serialização e invariância da prescrição. `scripts/test-training-history-query.ps1` extrai a consulta usada pelo repositório e a executa com fixtures sintéticas em transação somente leitura. A suíte Go sem cache passou antes do último ajuste defensivo; depois dele, os pacotes alterados passaram novamente. Uma repetição de `internal/httpapi`, não alterado nesta fatia, foi impedida pelo Controle de Aplicativos do Windows ao abrir o executável temporário, embora o pacote já tivesse passado na suíte anterior. `go vet`, build do frontend e validação do OpenAPI passaram.

A persistência foi confirmada manualmente no `GET /v1/plans/current`, com as três janelas, `data_issues` vazio e modo observacional preservado. Na conta usada, cinco sessões concluídas tinham duração zero e não havia sessão planejada já fechada nas janelas. O resultado correto foi taxa `null`, carga zero acompanhada de lacunas de cobertura e nenhuma mudança visual ou de prescrição.

### Qualidade temporal e sinais protetivos (`training-history-v2`)

A terceira fatia local mantém as janelas e acrescenta fatos necessários para uma interpretação futura segura:

- recência da última sessão concluída, da última carga session-RPE válida e do último check-in;
- contagem e exclusão de sessões concluídas no futuro e check-ins com data futura;
- registros de feedback, cobertura completa dos campos, dor, fadiga de 4 a 5 e RPE realizado pelo menos dois pontos acima do planejado;
- check-ins totais/completos, presença de sinais que já geram cautela e presença da combinação que já gera necessidade de recuperação.

Essas contagens reutilizam somente definições já existentes no produto. Não constituem uma classificação de tolerância e não alteram o plano. Feedback ausente e feedback incompleto são lacunas diferentes; dor é preservada mesmo se fadiga estiver ausente. `not_evaluated` lista tolerância à carga, destreinamento, mudança de condicionamento, atividades externas, fuso do atleta e progressão, enquanto `app_recording_gap_interpretation` declara que o intervalo significa somente ausência de atividade registrada no Cadência.

Essa cautela é deliberada. O session-RPE é uma medida útil e validada para monitorar carga interna, mas fatores contextuais afetam a percepção e a interpretação deve considerar outros sinais. A relação aguda:crônica não é usada porque há críticas conceituais e estatísticas fortes contra transformá-la em recomendação de carga ou risco. Estudos de destreinamento avaliam cessação ou redução controlada de treino; não sustentam concluir perda de condicionamento apenas porque uma atividade não foi registrada no aplicativo. Referências: Haddad et al. (2017), https://pubmed.ncbi.nlm.nih.gov/29163016/; Impellizzeri et al. (2020), https://pubmed.ncbi.nlm.nih.gov/32502973/; Zheng et al. (2022), https://pubmed.ncbi.nlm.nih.gov/36017396/; Rietjens et al. (2001), https://pubmed.ncbi.nlm.nih.gov/11726481/; Maldonado-Martín et al. (2017), https://pubmed.ncbi.nlm.nih.gov/27476326/.

Testes automatizados cobrem recência, cópia defensiva, metadados temporais divergentes, registros futuros, cobertura parcial e invariância da prescrição. A consulta real foi executada no PostgreSQL local com CTEs sintéticas e transação somente leitura; `go test -count=1 ./...`, `go vet ./...`, build do frontend e validação estrutural do OpenAPI passaram. A conferência manual de um novo rascunho confirmou `training-history-v2`, as três janelas, recência, cobertura e sinais protetivos com `data_issues: []` e `used_for_prescription: false`. Os cinco registros sem duração continuaram sem carga calculada, como esperado. Snapshots `training-history-v1` existentes não são atualizados retroativamente.

### Comparação observacional entre períodos (`training-history-v3`)

Novos rascunhos também registram `period_comparison`, uma série de seis blocos semanais não sobrepostos dentro dos últimos 42 dias: `last_7d`, `days_8_14`, `days_15_21`, `days_22_28`, `days_29_35` e `days_36_42`. O objetivo é permitir uma leitura posterior da distribuição dos registros sem misturar a semana atual com as anteriores.

Cada período repete medições brutas de aderência, sessões realizadas, minutos, cobertura de session-RPE, carga, feedback, dor, fadiga, RPE acima do alvo e check-ins de recuperação. A consulta mantém duas bases temporais já documentadas: sessões realizadas/carga por intervalos de `completed_at` no relógio do PostgreSQL; aderência/recuperação por datas relativas a `CURRENT_DATE`. Por isso, os campos não devem ser comparados como se viessem de um fuso do atleta.

O contrato é `period-comparison-v1`, em `mode: observation`, com `used_for_prescription: false`. Não há tendência calculada, razão aguda:crônica, inferência de destreinamento, limiar de tolerância, progressão ou regressão. Dados ausentes continuam sendo lacunas; dados inconsistentes entram em `data_issues`. O item `period_trend_for_prescription` permanece explicitamente não avaliado, e `rules-v1` não lê essa estrutura.

Os testes unitários verificam ordenação, seis períodos, taxas, separação das medições, invariância da prescrição e rejeição de período que não tenha sete dias. `scripts/test-training-history-query.ps1` executa a consulta real com fixtures sintéticas em transação somente leitura e verifica os seis blocos, inclusive as fronteiras temporais e o isolamento do atleta. A validação manual via API foi concluída em um novo rascunho antes do commit `64e554d`.

### Avaliação shadow do motor (`rules-v2`)

O `rules-v2` começou em paralelo, sem substituir o `rules-v1`. Durante a geração de um novo rascunho, `prescription_snapshot.rules_v2_shadow` avalia três gates determinísticos: integridade do período, sinais protetivos e evidência mínima para progressão. O resultado é congelado no snapshot com `mode: shadow` e escopo `plan_generation_only`. Essa avaliação foi versionada no commit `64e554d`; não está publicada na produção.

Os estados possíveis são:

- `protective_signal` / `prefer_recovery`: há limitação ativa, dor, fadiga elevada ou necessidade de recuperação observada; a resposta candidata é protetiva, mas não é aplicada pelo shadow;
- `observation_only` / `maintain_observed`: existem dois períodos recentes com sessão, carga session-RPE e feedback completos, além de um check-in de recuperação completo nos últimos 14 dias, sem sinal protetivo;
- `not_evaluated`: faltam períodos, cobertura mínima ou integridade dos dados para comparar a resposta.

Mesmo no segundo estado, não há autorização para progressão. A avaliação registra regras adiadas, motivos, lacunas e inconsistências, mantendo `progression_eligible: false`, `applied: false` e `used_for_prescription: false`. Ela não calcula destreinamento, tolerância, mudança fisiológica, ACWR ou resposta fora do Cadência. A conferência manual da API e a matriz controlada confirmaram os estados, as barreiras e a invariância dos treinos do `rules-v1`; um teste regressivo adicional mantém essa garantia localmente.

## IA explicativa opcional

O endpoint de explicação envia ao modelo apenas o nome, objetivo, duração, RPE-alvo, regras e escopo de evidência do treino. O modelo deve explicar a decisão em duas ou três frases; não recebe autorização para criar etapas, alterar carga, inventar referências ou interpretar sintomas. A integração usa Ollama local com limites de tempo, saída e concorrência e pode usar a rota protegida do Worker como fallback (Groq `openai/gpt-oss-20b`). Enquanto `AI_ENABLED=false`, ou quando os provedores estiverem indisponíveis, a API devolve o resumo validado pelo `rules-v1`.

## Transparência e segurança

Cada treino alterado guarda em `workouts.explanation.adaptation`:

- tipo e motivo da adaptação;
- treino que originou a decisão;
- duração e RPE-alvo anteriores;
- aviso de segurança, quando aplicável.

Os percentuais e limiares acima são escolhas prudentes desta versão do produto. Eles foram informados pelo uso consolidado do session-RPE para monitorar carga interna e pelo princípio de progressão gradual; não devem ser apresentados como valores universais comprovados ou substituir acompanhamento profissional.

## Referências primárias e revisões

- Foster et al. (2001), *A new approach to monitoring exercise training*: https://pubmed.ncbi.nlm.nih.gov/11708692/
- Haddad et al. (2017), revisão sistemática sobre validade do session-RPE: https://pubmed.ncbi.nlm.nih.gov/29163016/
- Impellizzeri et al. (2020), revisão de 25 anos do session-RPE: https://pubmed.ncbi.nlm.nih.gov/33508782/
- Bourdon et al. (2017), consenso sobre monitoramento de carga: https://pubmed.ncbi.nlm.nih.gov/28253038/
- ACSM (1998), progressão gradual do exercício aeróbico: https://pubmed.ncbi.nlm.nih.gov/9624661/
- Rosenblat, Perrotta e Thomas (2020), revisão e meta-análise sobre intervalos intensos versus sprints: https://pubmed.ncbi.nlm.nih.gov/32034701/

### Estado da expansão de evidências específicas de ciclismo

As referências de ciclismo sobre periodização, cadência, testes submáximos e distribuição de intensidade foram mapeadas e estão descritas em [`cycling-evidence-catalog.md`](cycling-evidence-catalog.md). A migração `000015` e o primeiro protocolo `road_moderate_intervals` estão publicados na produção após revisão de elegibilidade, limites de transferência da evidência e parâmetros. A ampliação deverá continuar com revisão própria e profissional habilitado quando necessário.

- Galán-Rioja et al. (2023): https://pubmed.ncbi.nlm.nih.gov/36640771/
- Mater, Clos e Lepers (2021): https://pubmed.ncbi.nlm.nih.gov/34360206/
- Capostagno, Lambert e Lamberts (2016): https://pubmed.ncbi.nlm.nih.gov/27701968/
- Seiler (2010): https://pubmed.ncbi.nlm.nih.gov/20861519/

### Estado operacional para a próxima revisão

O histórico observado, o fluxo de feedback, o check-in de recuperação e o fallback da IA explicativa já estão disponíveis em produção. Os fluxos funcionais e a latência, os limites e o fallback do Worker já foram testados. Relatos reais e o primeiro resumo semanal do Resend serão avaliados quando disponíveis, em paralelo às melhorias; não são pré-requisitos para desenvolver e testar a evolução do motor. Alterações de prescrição continuam exigindo critérios próprios, testes e revisão dos limites.

### Próxima evolução planejada

O roadmap atual prioriza prontidão, evolução versionada das regras, adaptação em ciclo fechado, integridade dos dados, segurança, feedback e auditabilidade. O `rules-v1` deve permanecer disponível durante a validação de qualquer evolução. O escopo desta fase é exclusivamente ciclismo; corrida e força não entram no catálogo atual. Toda entrega com mudança visível deve atualizar `APP_VERSION` e `UPDATE_NOTES` para informar o usuário na tela de novidades.
