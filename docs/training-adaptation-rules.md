# Adaptação do plano após o treino

Última revisão: 12 de setembro de 2026.

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

- **Semana de recuperação:** a quarta semana não recebe sessão de qualidade; o pedal mais longo permanece como endurance e as demais sessões usam recuperação ativa.
- **Indoor + nível intermediário:** sessão de cadência técnica com RPE moderado.
- **Terreno com subidas + nível intermediário ou avançado:** subidas controladas, com recuperação leve.
- **Nível avançado + medidor de potência e FTP informado:** sweet spot orientado pelo dado disponível, sem converter automaticamente o FTP em metas rígidas de watts nesta versão.
- **Nível avançado + meta de prova:** ritmo de prova controlado, sustentável e sem tentar reproduzir a prova completa.
- **Demais perfis:** mantém o tempo controlado; avançados sem contexto específico recebem sweet spot progressivo.
- **MTB XCO avançado elegível:** contexto `mtb_xco` explícito, objetivo de performance/prova, avaliação submáxima apta, pelo menos 75 minutos disponíveis e semana de construção liberam o piloto de intervalos aeróbicos XCO. Ele usa cinco blocos de 4 minutos com 4 minutos leves, RPE 7, sem sprint máximo ou técnica de trilha.
- **Estrada avançada elegível em ciclo alternado:** contexto `road` explícito, pelo menos oito semanas e três pedais semanais recentes, objetivo de performance/prova, avaliação submáxima apta, pelo menos 75 minutos disponíveis e semana de construção liberam o piloto de intervalos intensos de estrada. Ele usa até cinco blocos de 8 minutos com 4 minutos leves, RPE 8, sem meta rígida de potência; a carga é uma adaptação conservadora e não reproduz os blocos estudados.
- **Preferência VO₂max em estrada avançada elegível:** contexto `road` explícito, preferência `vo2max`, pelo menos oito semanas e três pedais semanais recentes, objetivo de performance/prova, avaliação submáxima apta, pelo menos 60 minutos disponíveis e semana de construção liberam o piloto de intervalos VO₂max de estrada. Ele usa quatro blocos de 4 minutos com 4 minutos leves, RPE 8, sem sprint máximo, cadência obrigatória ou meta fixa de potência; a estrutura adapta estudos em ciclistas bem treinados e não transforma VO₂max em um valor inferido pelo aplicativo.

Quando o ciclista informa preferências de sessão, elas orientam a escolha da sessão de qualidade dentro das mesmas proteções: cadência é elegível para intermediários e avançados; subidas exigem terreno com subidas; sweet spot exige nível avançado e, para potência, FTP informado; intervalos continuam exigindo avaliação submáxima apta, objetivo compatível e semana de construção; VO₂max exige adicionalmente estrada, nível avançado e histórico mínimo. Se todas as opções forem marcadas, o motor interpreta isso como abertura a qualquer protocolo e mantém a seleção contextual padrão. Giro/base e recuperação permanecem preferências registradas, sem transformar todos os dias em sessões de qualidade.

Se houver limitação ativa, a sessão específica é substituída pelo giro leve protegido. Iniciantes não recebem essas sessões de qualidade específicas ainda. Sprints máximos e estímulos de pista não fazem parte do Cadência; não devem ser adicionados como modalidade, preferência ou protocolo.

## Estrutura operacional das sessões

As sessões geradas agora carregam uma sequência de etapas acionáveis em `workout.structure.steps`, além dos campos antigos de aquecimento, parte principal e desaquecimento. Cada etapa informa ordem, tipo, duração, RPE-alvo e uma instrução curta para execução.

Uma sessão de subida, por exemplo, pode apresentar quatro esforços separados por recuperações leves, em vez de apenas o texto "4 blocos sustentados em subida". A soma das etapas é igual à duração prescrita e continua limitada à disponibilidade do dia. Em uma adaptação de duração, a interface ajusta proporcionalmente a exibição das etapas até que uma futura versão do motor passe a recalcular os blocos no próprio banco.

## Taper pré-prova orientado por evento (`taper-v1`)

O motor pode reduzir o volume do plano quando o atleta declara um evento futuro e atende aos gates conservadores do piloto. A decisão é registrada em `prescription_snapshot.event_taper`; não é uma nova modalidade, não aumenta a carga antes da prova e não substitui as proteções de segurança do `rules-v1`.

- **Elegibilidade do plano:** evento futuro válido entre 7 e 21 dias da geração, nível avançado, avaliação submáxima apta, pelo menos 8 semanas de treino e 3 pedais semanais informados.
- **Bloqueios:** limitação ativa, dor ou necessidade recente de recuperação mantêm o taper inativo e deixam as proteções existentes vencerem.
- **Aplicação:** sessões agendadas depois da data de geração e entre 1 e 14 dias antes do evento têm a duração multiplicada por `0,5`, respeitando o mínimo de 20 minutos. A frequência e o RPE-alvo são preservados.
- **Exceções:** o dia do evento não recebe taper e a quarta semana continua sob as regras de recuperação; ela não recebe sessões de qualidade por causa da meta de prova.
- **Auditoria:** a avaliação informa `version`, `status`, `applied`, `used_for_prescription`, distância para o evento, regras, motivos e evidências. As sessões afetadas recebem `event_taper_applied: true` e as referências do taper.

Esta é a primeira implementação local do piloto. As fontes `taper-meta-2023`, `taper-cyclist-2025` e `taper-overreach-cyclists-2023` são registradas na migração `000017`, aplicada apenas no banco de desenvolvimento nesta etapa. A produção ainda não recebeu a migração, o backend, a versão `0.13.0` nem a nota de novidades. O piloto VO₂max local é uma etapa posterior e separada, registrada na migração `000018`.

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
- `xco_aerobic_intervals`: cinco blocos de 4 minutos com 4 minutos leves, piloto publicado restrito a XCO avançado elegível; não inclui sprint máximo, descida ou técnica de trilha.
- `road_high_intensity_intervals`: até cinco blocos de 8 minutos com 4 minutos leves, piloto local restrito a estrada avançada elegível em ciclo alternado; não reproduz bloco concentrado, sprint máximo ou meta rígida de potência.
- `road_vo2_intervals`: quatro blocos de 4 minutos com 4 minutos leves, piloto local restrito a estrada avançada elegível com preferência explícita `vo2max`; não prescreve sprint, cadência ou potência fixa.
- `protected_recovery`: giro leve protegido quando existe uma limitação ativa.

As referências associadas sustentam princípios como progressão gradual, monitoramento de carga e uso contextual de intervalos. O ensaio de HIT em mountain bikers treinados e a revisão sistemática contemporânea de XCO orientam o piloto, mas não validam a mesma carga para todas as pessoas; os parâmetros continuam sujeitos às regras de segurança do produto.

### Piloto local em validação — intervalos curtos autorregulados

Este desenho preliminar foi implementado localmente como `short_self_regulated_intervals`, com seleção pelo `rules-v1` somente quando todos os gates abaixo passam. O objetivo continua sendo avaliar um estímulo curto de qualidade sem transformar o resultado dos estudos em uma dose universal e sem liberar sprint máximo; produção não recebeu o piloto.

**Escopo do piloto local:** estrada ou indoor, com preferência explícita do atleta; nível avançado; objetivo `performance` ou `event`; avaliação submáxima apta; pelo menos oito semanas de treino regular e três pedais semanais; no mínimo 50 minutos disponíveis; semana de construção; fase compatível com o evento; nenhuma limitação, dor ou proteção de recuperação ativa; e no máximo uma sessão de qualidade no ciclo semanal. XCO, XCM e gravel permanecem contextos distintos sem receber este piloto; downhill/enduro e pista sprint/BMX não fazem parte do produto.

**Estrutura conservadora do piloto:** 10 minutos de aquecimento em RPE 3–4; seis repetições de 1 minuto autorreguladas em RPE 7–8, cada uma seguida de pelo menos 1 minuto leve em RPE 2–3; e 10 minutos de desaquecimento. O atleta controla o esforço pela percepção e pela técnica, começa as primeiras repetições na parte baixa da faixa e não busca falha, sprint máximo, cadência obrigatória, potência fixa ou frequência cardíaca tratada como equivalente a VO₂max. A recuperação pode ser estendida; não há compensação por repetição interrompida.

**Travas e comportamento esperado:** qualquer dor, tontura, mal-estar, falta de ar incomum, perda de controle técnico ou esforço que deixe de ser autorregulável interrompe a parte intensa e encaminha para recuperação/avaliação profissional quando necessário. Sinais de sono, estresse, fadiga, dor pós-treino, limitação, semana de recuperação, taper ou feedback protetivo vencem a preferência e selecionam o giro leve protegido. O protocolo não pode ser escolhido em dias consecutivos de qualidade, não pode criar sessão extra e não pode aumentar carga automaticamente após um feedback fácil.

**Progressão e auditoria:** a primeira versão deve manter seis repetições fixas, sem aumentar simultaneamente duração e esforço. Uma eventual progressão posterior exigirá dados completos de execução, feedback sem sinal protetivo e revisão separada; não será inferida apenas por aderência, por ausência de dor ou por um valor de RPE isolado. A prescrição deve registrar versão, elegibilidade, preferência, estrutura, fontes, limites aplicados e motivo de não seleção quando algum gate falhar.

**Matriz mínima de validação:** um caso elegível deve selecionar somente o candidato; preferência ausente, modalidade não suportada, nível não avançado, objetivo incompatível, avaliação inapta, histórico menor que oito semanas, menos de três pedais semanais, disponibilidade abaixo do mínimo, semana de recuperação, taper/evento incompatível, dor/limitação, sinal de recuperação, sessão de qualidade já ocupando a semana e preferência concorrente devem impedir a seleção. Os testes devem provar que o candidato não altera sessões concluídas, não cria treinos extras, não ultrapassa minutos disponíveis, não substitui o `rules-v1`, mantém o fallback protegido e deixa o snapshot auditável. A implementação local foi adicionada com testes direcionados de seleção, não seleção, preferência e proteção; a suíte Go, o `go vet`, o build e a migração local passaram. A validação manual também confirmou o fluxo completo no navegador: a preferência foi salva e o plano regenerado apresentou **Intervalos curtos autorregulados**. A publicação continua condicionada à revisão do catálogo, backup e autorização explícita.

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

### Registro explícito de treino não realizado

Um treino `planned` ou `adapted` cuja data já passou pode ser encerrado pelo atleta como **Não realizei**. O backend altera o status para `skipped` somente depois de confirmar que o treino pertence ao plano ativo e que a data é anterior à data atual do banco. Nenhuma sessão é criada para representar uma atividade que não aconteceu.

O registro fecha a oportunidade de aderência e deixa de aparecer como pendência vencida, mas não é usado sozinho para inferir destreinamento, baixa tolerância ou necessidade de compensação. Não há reagendamento automático, sessão substituta, aumento de carga ou redução prescritiva nesta fatia. A próxima geração de plano continua sujeita às lacunas, aos sinais de dor/recuperação e às regras conservadoras já documentadas.

## Leitura observacional de prontidão (`readiness-v1`)

Implementação publicada no deploy de 11 de setembro de 2026, a partir do commit `61d7939`. Ao gerar um rascunho, o backend registra `prescription_snapshot.readiness_assessment` separadamente de `engine_version: rules-v1`. É uma descrição versionada dos dados disponíveis, não uma nova prescrição, diagnóstico ou leitura do estado de hoje. O horário UTC da classificação fica em `assessed_at`; ela não é recalculada ao abrir um plano antigo.

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

Implementação publicada no deploy de 11 de setembro de 2026, a partir do commit `61d7939`. Novos rascunhos congelam `prescription_snapshot.training_history` em modo `observation`; planos existentes não são recalculados. O campo é adjacente a `readiness_assessment` e mantém `used_for_prescription: false`, portanto não altera duração, RPE, estrutura ou escolha de sessão do `rules-v1`.

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

### Distribuição observacional dos estímulos (`training-history-v4` / `stimulus-distribution-v1`)

A fatia seguinte acrescenta ao snapshot de novos rascunhos uma leitura da distribuição dos estímulos exigentes realizados nos seis períodos semanais. O `period_comparison` passa a `period-comparison-v2` e cada período registra `quality_sessions`, `quality_sessions_with_load`, `quality_performed_minutes`, `quality_density_percent` e `high_intensity_sessions`. A sessão é considerada de qualidade quando o RPE-alvo persistido é pelo menos 6,0; alta intensidade usa RPE-alvo de pelo menos 7,0. Esses cortes são convenções operacionais do produto para agrupar sessões, não zonas fisiológicas nem regras universais.

O bloco `stimulus_distribution` resume a contagem de sessões de qualidade em 7, 14, 28 e 42 dias, o total de sessões de alta intensidade em 42 dias, a quantidade de pares de sessões de qualidade em dias consecutivos, o menor intervalo entre elas e a data da sessão mais recente. As datas de proximidade são normalizadas para o dia UTC de `completed_at`, porque o Cadência ainda não possui fuso horário individual. Sessões explicitamente inelegíveis pelo `data-integrity-v1` e sessões futuras não entram nessa leitura; duração ou RPE realizados ausentes geram lacunas de cobertura sem apagar o registro original.

O resultado permanece `mode: observation`, `used_for_prescription: false` e não altera o `rules-v1`, a seleção de estímulo, a duração, o RPE ou o trigger pós-feedback. Proximidade entre sessões não é interpretada como excesso de carga nem como diagnóstico. Antes de qualquer uso prescritivo, será necessária cobertura real, comparação com recuperação e feedback, revisão dos critérios e validação longitudinal.

Os testes unitários cobrem contagem por período, densidade, alta intensidade, recência, sessões em dias consecutivos, menor intervalo, datas ausentes e isolamento da prescrição. A fixture PostgreSQL verifica as novas colunas da consulta, o corte temporal, o filtro de integridade e o isolamento do atleta em transação somente leitura. Não há migração, mudança visual ou nota de versão nesta fatia backend-only.

### Avaliação shadow do motor (`rules-v2`)

O `rules-v2` começou em paralelo, sem substituir o `rules-v1`. Durante a geração de um novo rascunho, `prescription_snapshot.rules_v2_shadow` avalia três gates determinísticos: integridade do período, sinais protetivos e evidência mínima para progressão. O resultado é congelado no snapshot com `mode: shadow` e escopo `plan_generation_only`. Essa avaliação foi incluída no commit `61d7939` e está publicada na produção, mas continua não autoritativa.

Os estados possíveis são:

- `protective_signal` / `prefer_recovery`: há limitação ativa, dor, fadiga elevada ou necessidade de recuperação observada; a resposta candidata é protetiva, mas não é aplicada pelo shadow;
- `observation_only` / `maintain_observed`: existem dois períodos recentes com sessão, carga session-RPE e feedback completos, além de um check-in de recuperação completo nos últimos 14 dias, sem sinal protetivo;
- `not_evaluated`: faltam períodos, cobertura mínima ou integridade dos dados para comparar a resposta.

Mesmo no segundo estado, não há autorização para progressão. A avaliação registra regras adiadas, motivos, lacunas e inconsistências, mantendo `progression_eligible: false`, `applied: false` e `used_for_prescription: false`. Ela não calcula destreinamento, tolerância, mudança fisiológica, ACWR ou resposta fora do Cadência. A conferência manual da API e a matriz controlada confirmaram os estados, as barreiras e a invariância dos treinos do `rules-v1`; um teste regressivo adicional mantém essa garantia localmente.

### Primeiro desenho de adaptação pós-treino (`rules-v2-adaptation-v1`)

O `rules-v1` já possui uma adaptação pós-feedback ativa no trigger `feedback_adapts_future_workouts`. A sexta fatia acrescenta uma avaliação paralela em Go; a sétima a executa na mesma transação do fluxo de conclusão e a grava somente em `workouts.explanation.adaptation_shadow` do treino concluído. Ela não altera duração, RPE, estímulo ou status e não substitui o comportamento já testado.

O avaliador recebe o RPE-alvo, o feedback validado da sessão e os períodos observados. Ele mantém quatro gates explícitos:

- feedback e alvo precisam ser válidos;
- dor, esforço muito alto, fadiga máxima ou sinal protetivo recente bloqueiam progressão e produzem `prefer_recovery`;
- resposta dentro do esperado produz `maintain_observed`;
- uma resposta claramente fácil só pode produzir `progress_duration_5pct` como candidata se houver seis períodos íntegros, dois períodos recentes com sessões, carga session-RPE e feedback completos, além de recuperação completa registrada nesses períodos. Sem isso, produz `defer_progression`.

Mesmo com evidência suficiente, a candidata permanece em `mode: shadow`, com `progression_eligible: false`, `applied: false` e `used_for_prescription: false`. O módulo não infere destreinamento, tolerância, mudança fisiológica, efeito da prescrição ou dados de atividades fora do Cadência. Se a consulta dos períodos falhar, um savepoint impede que a observação bloqueie o feedback: o resultado fica `not_evaluated` e recebe `history_query_failed`. Os testes cobrem resposta fácil isolada, evidência completa, dor, período inconsistente e a invariância de não produzir candidata com observação inconsistente. A conferência manual de `explanation.adaptation_shadow` no `GET /v1/plans/current` foi concluída; a validação seguinte é a do bloco de comparação planejado versus realizado.

### Integridade observacional da sessão (`data-integrity-v1`)

Antes de uma futura adaptação prescritiva, a sessão concluída recebe uma leitura separada de integridade. O gate classifica os dados como `valid`, `incomplete` ou `inconsistent` e grava o resultado em `workouts.explanation.data_integrity`, sem rejeitar, corrigir ou sobrescrever o registro original.

São verificados os dados mínimos de duração positiva, RPE realizado, feedback e fadiga, além das faixas das métricas opcionais. Também são registradas combinações incompatíveis, como duração zero com distância ou elevação. Um valor ausente é lacuna; um valor impossível ou incompatível é inconsistência. Sessões incompletas ou inconsistentes não entram como elegíveis no shadow do pós-treino, e a decisão continua sem aplicação.

As consultas que montam o resumo observado, as janelas de histórico, os seis períodos comparativos e os agregados da tela de Evolução respeitam `workouts.explanation.data_integrity.eligible_for_history`. Um registro explicitamente inelegível deixa de alimentar minutos, carga session-RPE, dor, fadiga, RPE acima do alvo e recência observados. Treinos legados sem esse bloco continuam legíveis; a aderência planejada permanece separada da qualidade dos dados realizados.

O resumo observado de 28 dias também usa o intervalo fechado entre `now() - interval '28 days'` e `now()`. Assim, uma sessão com `completed_at` futuro não influencia minutos, médias, dor, fadiga ou cobertura do contexto de prontidão. A mesma referência temporal já é usada nas janelas cumulativas e nos períodos não sobrepostos; o filtro não apaga nem reclassifica o registro original.

Os agregados da tela de Evolução seguem a mesma proteção: sessões concluídas ou canceladas no futuro não entram no resumo total, na semana, nas sessões recentes ou nos pontos de recuperação. O filtro temporal também impede que um check-in com `recorded_on` futuro seja apresentado como observação atual. Esses filtros não removem registros e não transformam a Evolução em fonte prescritiva.

Esta fatia não altera o trigger `feedback_adapts_future_workouts` nem a prescrição `rules-v1`. Portanto, a barreira ativa contra progressão de registros inconsistentes continua sendo uma decisão posterior, depois de validar a observação e sua integração com o histórico.

### Observação de tolerância à carga (`load-tolerance-v1`)

Esta fatia acrescenta uma leitura observacional dentro de `adaptation_shadow`. Ela avalia os dois períodos semanais mais recentes, não sobrepostos, somente quando cada um possui sessões realizadas com carga por session-RPE, feedback completo e ao menos um check-in de recuperação completo. A leitura registra os períodos usados, as lacunas e as inconsistências, sem calcular ACWR, inferir tolerância fisiológica ou concluir destreinamento.

Dor, fadiga alta, necessidade recente de recuperação e esforço percebido pelo menos dois pontos acima do alvo produzem `protective_signal`/`prefer_recovery`. Sem esses sinais e com os dois períodos completos, o resultado é `observation_only`/`maintain_observed`: isso descreve suporte observacional para manter a carga sob análise, não uma autorização para aumentá-la. A avaliação também fica com `progression_eligible: false`, `applied: false` e `used_for_prescription: false`.

O resultado protetivo de `load-tolerance-v1` participa também do gate do `rules-v2-adaptation-v1`. Assim, um sinal recente de esforço acima do alvo (`recent_above_target_rpe`), além de dor, fadiga alta ou recuperação necessária, mantém a candidata em `protective_signal`/`prefer_recovery`, mesmo que a sessão atual isoladamente pareça fácil. Essa integração continua em shadow e não concede autoridade prescritiva ao bloco.

Para uma candidata de progressão, o shadow também exige que `load-tolerance-v1` esteja em `observation_only`: cada um dos dois períodos precisa ter sessões com carga por session-RPE, feedback completo e ao menos um check-in de recuperação completo. Um check-in presente apenas no período anterior não cobre a lacuna do período recente; nesse caso, a candidata fica em `not_evaluated`/`defer_progression`.

A aderência planejada também funciona como gate separado para a progressão shadow. Se qualquer um dos dois períodos de evidência tiver sessão prevista perdida (`missed_sessions`) ou treino em andamento vencido (`overdue_in_progress_sessions`), a candidata fica em `not_evaluated`/`defer_progression` com `low_adherence`. Cancelamentos explícitos (`cancelled_sessions`) continuam registrados, mas não são convertidos sozinhos em baixa aderência.

O `rules-v1` continua sendo a única fonte prescritiva. A implementação não altera o trigger pós-feedback, não modifica sessões futuras, não cria migração e não muda a interface; por isso, não exige nova nota de versão nesta fatia. A evidência de session-RPE orienta o método de registro, mas os critérios de cobertura e os estados são barreiras prudentes do produto, não limiares fisiológicos universais.

### Comparação observacional entre planejado e realizado (`planned-vs-actual-v1`)

Esta fatia registra, dentro de `workouts.explanation.adaptation_shadow`, a diferença descritiva entre a sessão planejada e a sessão concluída. São comparados a duração planejada e realizada, o RPE-alvo e o RPE realizado, além da cobertura do feedback, do contexto de conclusão e das métricas opcionais disponíveis no encerramento.

O resultado pode ficar `observed` quando os dados mínimos da comparação estão válidos ou `not_evaluated` quando há duração, RPE, feedback ou contexto de conclusão inválidos/ausentes. `completion_status` informa `complete` ou `partial`; neste último caso, `partial_reason` identifica um dos motivos controlados. Cadência, sono, estresse e recuperação continuam explícitos em `not_evaluated`; métricas opcionais ausentes ficam em `missing_data` sem invalidar a comparação principal.

As diferenças de duração e RPE, assim como o status de conclusão e seu motivo, são apenas registros de execução. O bloco mantém `progression_eligible: false` e `used_for_prescription: false`, não interpreta tolerância fisiológica nem aplica limiares de progressão. A validação manual local confirmou uma sessão de 3 minutos realizados de 35 planejados como `observed`, sem alteração da próxima sessão. Esta documentação antecede a validação manual da nova interface; a implementação local usa a migração `000020` e a versão `0.17.0`, sem deploy.

### Contexto de conclusão parcial (`completion_status`)

Ao concluir um treino, o atleta informa se realizou a sessão completa ou apenas parte dela. Para uma conclusão parcial, o motivo é obrigatório e fica restrito a `time_available_changed`, `fatigue_or_recovery`, `pain_or_discomfort`, `equipment_or_conditions` ou `other`. Registros antigos recebem `complete` por padrão na migração `000020`.

O contexto é informativo e não diagnostica a causa da interrupção. Uma sessão parcial pode entrar na observação quando seus dados são coerentes, mas não é usada como evidência de tolerância ao treino completo e não libera progressão. O trigger do `rules-v1` mantém proteções para dor, fadiga e esforço alto; somente a ramificação de progressão exige `completion_status = 'complete'`.

Na avaliação `rules-v2-adaptation-v1`, a conclusão parcial agora é um gate explícito do shadow. Depois que sinais protetivos de dor, fadiga ou esforço alto são priorizados, uma sessão `partial` recebe `status: not_evaluated`, `candidate_response: defer_progression` e o motivo `partial_completion`. O `adaptation-audit-v1` também registra `partial_completion` como restrição; isso torna visível que a sessão foi observada, mas não serve como evidência de tolerância ao treino inteiro.

### Contexto adicional pós-treino (`000021`)

O feedback pode registrar também `recovery_after` e `repeat_confidence`, ambos em escala de 1 a 5. O primeiro descreve a recuperação percebida após a sessão; o segundo registra a confiança do atleta para repetir aquele treino. Os campos são opcionais no armazenamento para preservar feedbacks antigos e são preenchidos pelo formulário atual com uma resposta neutra inicial.

Esses sinais são observacionais nesta versão: não são diagnóstico, não substituem o check-in diário, não autorizam progressão e não alteram o `rules-v1` ou o trigger pós-feedback. Quando presentes, aparecem no resumo da sessão, no histórico e em `planned-vs-actual-v1`; valores fora da faixa são rejeitados na API, no banco e na integridade observacional. A interpretação futura exige cobertura suficiente, comparação com a carga realizada e revisão específica antes de qualquer uso prescritivo.

Na fatia técnica seguinte, esses mesmos sinais passaram a ser classificados também pelo bloco `post-workout-context-v1`, aninhado em `workouts.explanation.adaptation_shadow`. O bloco registra cobertura completa, parcial, ausente ou inválida por meio de `observed_fields`, `missing_data`, `data_issues` e motivos explicáveis. Mesmo quando os dois valores estão presentes, `candidate_response: maintain_observed`, `progression_eligible: false` e `used_for_prescription: false` deixam explícito que o resultado é somente registro; não há limiar fisiológico, tendência longitudinal ou efeito prescritivo sendo inferido.

### Auditoria da decisão em shadow

O bloco `adaptation-audit-v1`, também aninhado em `adaptation_shadow`, registra a proveniência da avaliação: dados considerados, lacunas, restrições aplicadas, caminhos de adaptação não selecionados e condições informativas para uma futura revisão. `confidence` permanece `not_calibrated`, porque ainda não existe calibração estatística ou validação de efeito que justifique um nível de confiança.

Essa auditoria não transforma condições futuras em gatilhos automáticos. Ela mantém `used_for_prescription: false`, não altera o `rules-v1`, não substitui o check-in diário e não interpreta o efeito fisiológico de uma sessão. Os caminhos listados como não selecionados são explicativos; qualquer mudança prescritiva exigirá cobertura, comparação, calibração e revisão específicas.

A proveniência também consolida as lacunas de `post_workout_context`, `load_tolerance` e `planned_vs_actual` quando esses blocos estão disponíveis. Campos básicos só entram em `data_used` quando o feedback e o RPE passaram pela validação mínima; dados inválidos não são apresentados como se tivessem sustentado a decisão. A auditoria continua descritiva e não autoritativa.

Na revisão de precedência, sinais protetivos continuam vencendo uma conclusão parcial. Sem proteção, a progressão permanece adiada quando `load-tolerance-v1` não consegue classificar os dois períodos exigidos; nesse caso, `decision_audit.constraints_applied` registra `load_tolerance_gate` para deixar explícito que a lacuna foi aplicada como barreira. A candidata continua em shadow e não é autorização de carga.

No fluxo transacional, o `decision_audit` é reconstruído após a anexação de `planned_vs_actual` e após o tratamento de `history_query_failed`. Assim, as lacunas da comparação chegam à auditoria consolidada e um histórico indisponível não aparece em `data_used`; o `history_query_gate` deixa a falha explícita sem transformar a observação em bloqueio do feedback.

O contrato HTTP autenticado foi coberto para garantir que essa auditoria final chega ao `GET /v1/plans/current` sem perder `planned_vs_actual`, lacunas ou a barreira `used_for_prescription: false`.

Uma matriz final de invariantes cobre os estados protetivo, parcial, incompleto, inconsistente, baixa aderência, integridade inválida e candidato completo. Em todos eles, a avaliação permanece em shadow e `progression_eligible`, `applied` e `used_for_prescription` continuam falsos.

### Rotação segura e recuperação ativa

O catálogo geral possui o protocolo `active_recovery`, apresentado ao atleta como **Recuperação ativa**. O motor o seleciona somente para uma sessão de base na quarta semana do ciclo. A sessão mantém o multiplicador de recuperação já existente, usa alvo RPE 3,5 e uma instrução de pedal leve e contínuo; não representa uma prescrição universal de minutos ou intensidade.

Essa variação reduz a repetição nominal sem criar uma nova modalidade ou alterar os protocolos específicos de estrada e XCO. A evidência `acsm-1998` sustenta apenas o princípio de progressão gradual e controle de carga, não os minutos dessa sessão. Limitação ativa, dor relatada e sinais recentes que exigem recuperação continuam vencendo a escolha e substituindo-a por `Giro leve protegido`. A seleção é determinística e não depende da IA, do `rules-v2` shadow ou de dados ausentes.

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

### Estado publicado do catálogo ampliado — 12 de setembro de 2026

Os pilotos `taper-v1`, `road_vo2_intervals` e `short_self_regulated_intervals` foram publicados após validação local, backup e aplicação ordenada das migrações. Eles continuam limitados pelos gates documentados e pelo `rules-v1`. Sprint/pista/BMX e downhill/enduro permanecem fora do produto e não são candidatos de prescrição.
