# Adaptação do plano após o treino

Última revisão: 2 de setembro de 2026.

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

## IA explicativa opcional

O endpoint de explicação envia ao modelo apenas o nome, objetivo, duração, RPE-alvo, regras e escopo de evidência do treino. O modelo deve explicar a decisão em duas ou três frases; não recebe autorização para criar etapas, alterar carga, inventar referências ou interpretar sintomas. A integração usa Ollama local com limites de tempo, saída e concorrência e pode usar a rota protegida do Worker como fallback. Enquanto `AI_ENABLED=false`, ou quando os provedores estiverem indisponíveis, a API devolve o resumo validado pelo `rules-v1`.

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

### Etapa futura: ampliar evidências específicas de ciclismo

As referências de ciclismo sobre periodização, cadência, testes submáximos e distribuição de intensidade foram mapeadas para uma próxima etapa. Elas ainda não estão cadastradas no banco nem associadas aos protocolos; a implementação deverá ocorrer depois de consolidar o uso do histórico observado pelo motor e revisar os parâmetros com profissional habilitado.

- Galán-Rioja et al. (2023): https://pubmed.ncbi.nlm.nih.gov/36640771/
- Mater, Clos e Lepers (2021): https://pubmed.ncbi.nlm.nih.gov/34360206/
- Capostagno, Lambert e Lamberts (2016): https://pubmed.ncbi.nlm.nih.gov/27701968/
- Seiler (2010): https://pubmed.ncbi.nlm.nih.gov/20861519/
