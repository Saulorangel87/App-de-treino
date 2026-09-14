# Mapa de evidências do catálogo de ciclismo

Última revisão: 12 de setembro de 2026.

## Objetivo

Este documento registra a ponte entre a literatura científica e a futura biblioteca de protocolos do Cadência. Ele não é uma prescrição médica e não transforma resultados de estudos em metas universais.

O princípio de implementação é:

```text
demanda da modalidade
        +
evidência de eficácia do estímulo
        +
nível, histórico, disponibilidade e recuperação do atleta
        =
protocolo elegível e limitado
```

Um estudo que descreve a intensidade de uma prova não valida automaticamente um treino com os mesmos números. Estudos de demanda servem para orientar a especificidade; ensaios e revisões de intervenção servem para sustentar a escolha do treinamento.

## Síntese por modalidade

### Estrada

Existe a base mais aproveitável para a primeira expansão. A literatura recente sugere que diferentes distribuições de intensidade podem melhorar o desempenho, sem uma superioridade universal do modelo polarizado para todos os desfechos. Dois estudos de 2025 encontraram melhora com blocos de intervalos moderados e intensos em ciclistas bem treinados, mas seus participantes não representam automaticamente iniciantes ou ciclistas recreacionais.

**Uso previsto no Cadência:** endurance contínuo, tempo controlado e intervalos progressivos com liberação condicionada à avaliação, nível e recuperação.

**Limite:** os parâmetros dos estudos não devem ser copiados como receita fixa. O motor deve continuar usando RPE e duração disponível quando não houver potência ou frequência cardíaca confiável.

### MTB XCO

O XCO contemporâneo combina alta capacidade aeróbica, esforços curtos acima da potência aeróbica máxima, largadas rápidas, subidas e trechos técnicos. A revisão sistemática de 2026 reuniu 53 estudos e destaca a natureza intermitente da modalidade, mas também informa que avaliações diretas de desempenho das intervenções ainda são escassas.

**Uso nesta fatia:** piloto local de intervalos aeróbicos controlados, restrito a atletas avançados com contexto XCO explícito, objetivo de performance ou prova, avaliação submáxima apta, disponibilidade suficiente e ausência de sinais de recuperação insuficiente.

**Limite:** a distribuição observada em uma prova não será convertida diretamente em séries universais. A parte técnica permanece instrução de habilidade e não deve ser simulada automaticamente pelo aplicativo.

### XCM e gravel

As provas combinam distância, elevação, terreno variável e oportunidades limitadas de alimentação. A evidência específica para prescrição de sessões ainda é menor que a de estrada e XCO.

**Uso previsto no Cadência:** reutilizar endurance e progressão de volume já validados, acrescentando o contexto de terreno somente quando houver dados suficientes.

**Limite:** não criar um protocolo “gravel” apenas por trocar o nome de uma sessão de estrada. Nutrição e hidratação devem permanecer orientações separadas, sem metas automáticas baseadas em um único estudo de campo.

### Downhill e enduro — fora do produto

O desempenho depende de técnica, controle corporal, força isométrica e tolerância à fadiga de pegada, além do condicionamento. Estudos recentes também mostram risco relevante de lesão em treinamento e prova.

**Decisão de escopo:** não implementar downhill/enduro no Cadência. Descidas, saltos e treinos técnicos não serão oferecidos como prescrição automática ou modalidade do app.

**Registro:** as fontes dessa modalidade podem permanecer no catálogo apenas como referência de segurança e justificativa de exclusão; não autorizam criar um módulo futuro dentro deste produto.

### Pista sprint e BMX — fora do produto

São modalidades com exigências anaeróbicas, neuromusculares e de força muito diferentes das sessões de endurance do MVP. A literatura de velocistas de pista descreve grande concentração de carga nas zonas mais intensas, mas isso não sustenta reutilizar o catálogo atual.

**Decisão de escopo:** não implementar sprint/pista/BMX no Cadência. Essas modalidades não serão oferecidas como opção de perfil, preferência ou prescrição no app.

## Protocolos candidatos

| Candidato | Modalidade | Situação da evidência | Liberação inicial |
| --- | --- | --- | --- |
| Endurance de estrada | `road` | Base de endurance e distribuição de intensidade apoiadas por revisões | Todos os níveis, com progressão conservadora |
| Pedal longo | `road`, `gravel`, `mtb_xcm`, `indoor` | Extensão de endurance com progressão gradual; não define uma dose universal | Todos os níveis, no maior slot disponível e respeitando disponibilidade e recuperação |
| Intervalos moderados | `road` | Ensaios recentes em ciclistas bem treinados | Piloto local: intermediário/avançado, objetivo compatível, avaliação apta, 60 min disponíveis e sem sinais de recuperação insuficiente |
| Intervalos intensos de estrada | `road` | Ensaio recente comparando blocos moderados e intensos em ciclistas bem treinados, com limite de transferência explícito | Piloto local: avançado, 8 semanas e 3 pedais/semana recentes, avaliação apta, objetivo compatível, pelo menos 75 min, ciclo alternado e sem sinais protetivos |
| Intervalos VO₂max de estrada | `road` | Ensaios recentes em ciclistas bem treinados; transferência limitada | Piloto local: avançado, preferência explícita, histórico mínimo, avaliação apta, pelo menos 60 min e sem sinais protetivos |
| Intervalos curtos autorregulados | `road`, `indoor` | Ensaio recente e estudo em ciclistas treinados, com protocolos e populações heterogêneos | Piloto local: avançado, preferência explícita, histórico mínimo, avaliação apta, pelo menos 50 min e sem sinais protetivos |
| Intervalos aeróbicos XCO | `mtb_xco` | Demanda bem descrita; intervenção direta favorável ao HIT, mas em população treinada | Piloto local restrito, sem sprint máximo ou técnica de trilha |
| Resistência específica na bicicleta | `road`, `indoor` | Ensaio randomizado de 2025 favorável, mas dependente de calibração de força e esforços máximos | Candidato bloqueado: exige medição de força/carga, população avançada e módulo de segurança próprio |
| Endurance gravel/XCM | `gravel`, `mtb_xcm` | Evidência direta de prescrição ainda insuficiente | Usar somente base/endurance contextual |
| Força complementar | `road`, `mtb_xco` | Meta-análise recente favorável, mas com baixa certeza | Módulo opcional e separado do treino de bike |
| Downhill/enduro técnico | — | Evidência de risco, sem protocolo automatizado seguro | Fora do produto |
| Sprint de pista/BMX | — | Modalidade distinta do MVP | Fora do produto |

As situações acima combinam protocolos ativos, candidatos sob revisão e exclusões permanentes. Antes de transformar um candidato permitido em regra do `rules-v1`, seus parâmetros, população-alvo e critérios de interrupção devem ser revisados por profissional habilitado.

### Primeiro piloto publicado: intervalos moderados de estrada

O primeiro protocolo específico implementado e publicado é `road_moderate_intervals`, apresentado como **Intervalos moderados de estrada**. Ele usa três blocos de 10 minutos com três minutos de recuperação leve, alvo RPE 6 e uma sessão de qualidade por semana. Essa dose é uma adaptação conservadora do contexto dos estudos, não a reprodução do bloco de seis sessões em sete dias.

O motor só o seleciona quando a disciplina é explicitamente `road`, o atleta é intermediário ou avançado, a avaliação submáxima está apta, o objetivo é performance ou evento, há pelo menos 60 minutos disponíveis e a preferência está vazia ou indica intervalos. Dor, limitação ou recuperação insuficiente substituem o protocolo por uma sessão protegida.

### Terceiro piloto local: intervalos intensos de estrada

O terceiro protocolo específico é `road_high_intensity_intervals`, apresentado como **Intervalos intensos de estrada**. Ele usa até cinco blocos de oito minutos com quatro minutos de recuperação leve, alvo RPE 8 e uma única sessão de qualidade na semana. A estrutura pode reduzir o número de blocos quando a duração da semana exigir; não há meta obrigatória de potência ou frequência cardíaca.

O motor só o seleciona para atleta avançado com disciplina `road`, pelo menos oito semanas de treino recente e três pedais semanais informados, objetivo `performance` ou `event`, avaliação submáxima apta, pelo menos 75 minutos disponíveis, preferência vazia ou por intervalos e ciclo alternado (`rotation_index` ímpar). A liberação ocorre apenas nas semanas de construção. A sessão é substituída por proteção quando há limitação, dor ou recuperação insuficiente; iniciantes, intermediários e atletas retornando após baixa consistência continuam fora deste piloto.

O estudo de Rønnestad et al. (2025) comparou blocos de intervalos moderados e intensos em 22 ciclistas bem treinados; ambos melhoraram alguns indicadores, com respostas dependentes da intensidade. O Cadência não reproduz as cinco sessões em seis dias, o RPE do estudo ou a carga concentrada: usa somente uma adaptação conservadora para testar a elegibilidade do estímulo no produto. A revisão de Rosenblat et al. (2020) informa a escolha de intervalos em vez de sprints máximos, sem validar a dose desta implementação.

### Segundo piloto publicado: intervalos aeróbicos XCO

O segundo protocolo específico publicado como piloto é `xco_aerobic_intervals`, apresentado como **Intervalos aeróbicos XCO**. Ele usa cinco blocos de quatro minutos com quatro minutos de recuperação leve, alvo RPE 7 e uma única sessão de qualidade no ciclo. A estrutura é uma adaptação conservadora do HIT estudado em mountain bikers treinados; não reproduz a frequência, a progressão de seis semanas ou a carga do ensaio e não inclui sprint máximo.

O motor só o seleciona quando a disciplina é explicitamente `mtb_xco`, o atleta é avançado, a avaliação submáxima está apta, o objetivo é performance ou evento, há pelo menos 75 minutos disponíveis, a semana não é de recuperação e a preferência está vazia ou indica intervalos. Limitação, dor, recuperação insuficiente, dados ausentes ou inconsistentes e qualquer outro perfil substituem ou impedem o piloto. A sessão não prescreve descidas, saltos, técnica de trilha, força complementar ou metas rígidas de potência.

O estudo randomizado de Inoue et al. encontrou melhora do desempenho de MTB após seis semanas de HIT ou SIT, com vantagem provável do HIT; a revisão sistemática contemporânea de XCO de 2026 confirma a combinação de alta demanda aeróbica e esforços intermitentes, mas ressalta que avaliações diretas de desempenho das intervenções ainda são escassas. Por isso, a implementação permanece um piloto restrito ao perfil elegível e não altera o protocolo ativo para perfis gerais.

### Protocolo geral diferenciado: pedal longo

O slot de maior disponibilidade do ciclo passa a ser apresentado como **Pedal longo**, com a chave `long_endurance`. Ele representa uma sessão contínua de endurance com volume relativamente maior dentro da rotina do próprio atleta; não é uma meta fixa de distância, velocidade ou duração universal.

O protocolo reutiliza a mesma base segura do endurance contínuo: duração derivada do nível e do maior período disponível, RPE-alvo 5, aquecimento, parte principal contínua e desaquecimento. Limitação, dor, recuperação insuficiente, disponibilidade menor e os demais gates do `rules-v1` continuam prevalecendo. A referência `acsm-1998` sustenta somente progressão gradual e controle de carga; o nome não libera carga adicional.

## Modelo e critérios de integração do catálogo

O contexto agora guarda `bike_type`, `terrain` e uma disciplina explícita, opcional e validada. A disciplina não é inferida pelo tipo de bicicleta: XCO, XCM, gravel e os demais contextos permitidos só podem ser usados quando o atleta os informa diretamente. As migrações `000015` e `000016` registram as fontes do catálogo inicial e do piloto XCO na produção. Os identificadores legados `dh_enduro` e `track_sprint` não são modalidades válidas do app e são rejeitados ao salvar o perfil. Cada protocolo permitido continua dependendo de revisão de elegibilidade, segurança e transferência da evidência antes de ser publicado.

Valores planejados para `cycling_context.discipline`:

- `general`: ciclismo sem disciplina informada;
- `road`: estrada;
- `mtb_xco`: MTB cross-country olímpico;
- `mtb_xcm`: MTB maratona;
- `gravel`: gravel;
- `indoor`: indoor/rolo;
- `dh_enduro` e `track_sprint` são identificadores legados rejeitados e não fazem parte dos valores planejados.

Quando o campo estiver vazio ou for `general`, somente protocolos gerais e já existentes poderão ser selecionados. Perfis antigos não devem ser migrados automaticamente para uma modalidade específica.

### Protocolo geral implementado localmente — recuperação ativa

O protocolo `active_recovery` é uma variação geral de baixa carga para a semana de recuperação. Ele mantém o atleta em movimento com esforço leve e volume reduzido, sem séries intensas, sprint ou alvo obrigatório de potência. A implementação usa `acsm-1998` somente para o princípio de progressão gradual e controle de carga; a referência não define a duração usada pelo Cadência.

No motor, a sessão aparece apenas em uma posição de base da quarta semana. Limitações, dor e recuperação insuficiente continuam bloqueando a variação e selecionando `Giro leve protegido`. Como não depende de modalidade específica, ela não libera protocolos de estrada, XCO, gravel, XCM, downhill, enduro ou pista sprint/BMX.

Cada novo protocolo também deverá declarar, em código:

- chave estável;
- modalidade elegível;
- nível mínimo;
- necessidade de avaliação apta;
- sensores necessários ou alternativa por RPE;
- evidências associadas;
- população e limite de transferência da evidência;
- progressão e semana de recuperação;
- travas por limitação, dor e recuperação insuficiente.

## Registro inicial de fontes

- **`road-intensity-2024`** — Oliveira, Boppre e Fonseca. *Comparison of Polarized Versus Other Types of Endurance Training Intensity Distribution on Athletes' Endurance Performance: A Systematic Review with Meta-analysis*. 2024. Revisão de 17 estudos; apoia comparação de distribuição de intensidade, sem superioridade universal para tempo de prova. https://pubmed.ncbi.nlm.nih.gov/38717713/
- **`road-mit-block-2025`** — *A Moderate-Intensity Interval Training Block Improves Endurance Performance in Well-Trained Cyclists*. 2025. Ensaio em ciclistas bem treinados; informa um bloco moderado, não uma regra para iniciantes. https://pubmed.ncbi.nlm.nih.gov/40101160/
- **`road-block-comparison-2025`** — *Block Training With Moderate- or High-Intensity Intervals Both Improve Endurance Performance in Well-Trained Cyclists*. 2025. Compara blocos moderados e intensos; os efeitos dependem do desfecho e da população treinada. https://pubmed.ncbi.nlm.nih.gov/41169000/
- **`road-strength-2026`** — Llanos-Lagos et al. *Heavy strength training effects on physiological determinants of endurance cyclist performance: a systematic review with meta-analysis*. 2026. Relata efeitos favoráveis em eficiência, potência anaeróbica e desempenho, mas com baixa certeza para definir a implementação ótima. https://pubmed.ncbi.nlm.nih.gov/40632222/
- **`xco-physiology-2026`** — Protzen et al. *The Physiology of Contemporary Olympic Cross-Country Mountain Biking: A Systematic Review*. 2026. Revisão de 53 estudos sobre o XCO contemporâneo; sustenta a especificidade intermitente, não uma receita fixa de séries. https://pubmed.ncbi.nlm.nih.gov/41739301/
- **`xco-power-distribution-2021`** — *Aerobic and Anaerobic Power Distribution During Cross-Country Mountain Bike Racing*. 2021. Estudo de demanda de prova; descreve esforços curtos e repetidos acima da potência aeróbica máxima. https://pubmed.ncbi.nlm.nih.gov/33848975/
- **`xco-pacing-2021`** — *Exercise Intensity and Pacing Pattern During a Cross-Country Olympic Mountain Bike Race*. 2021. Estudo de intensidade e pacing em prova XCO; usado somente para especificidade da modalidade. https://pubmed.ncbi.nlm.nih.gov/34349670/
- **`xco-hit-2016`** — Inoue et al. *Effects of Sprint versus High-Intensity Aerobic Interval Training on Cross-Country Mountain Biking Performance: A Randomized Controlled Trial*. 2016. Ensaio randomizado com 16 mountain bikers treinados; compara HIT e SIT por seis semanas e informa a escolha de um piloto aeróbico XCO conservador. https://pubmed.ncbi.nlm.nih.gov/26789124/
- **`gravel-field-2024`** — *Fluid Intake and Hydration Responses to Mass Participation Gravel Cycling*. 2024. Estudo de campo sobre gravel; informa contexto de distância e hidratação, não valida sozinho um protocolo de treino. https://pubmed.ncbi.nlm.nih.gov/39807388/
- **`dh-injury-2024`** — Fallon et al. *Downhill race for a rainbow jersey: the epidemiology of injuries in downhill mountain biking at the 2023 UCI cycling world championships*. 2024. Estudo observacional de lesões; usado como trava de segurança, não como prescrição. https://pubmed.ncbi.nlm.nih.gov/39411021/
- **`mtb-crash-mechanisms-2025`** — Bonte et al. *Injury Mechanisms in Mountain Biking: A Systematic Video Analysis of 534 Cases*. 2025. Estudo de mecanismos de queda; reforça que habilidade técnica e prevenção não devem ser reduzidas a carga aeróbica. https://pubmed.ncbi.nlm.nih.gov/40534393/
- **`track-sprint-load-2023`** — *Training load and intensity distribution for sprinting among world-class track cyclists*. 2023. Descrição de treinamento de velocistas de pista; modalidade fora do escopo do primeiro catálogo ampliado. https://pubmed.ncbi.nlm.nih.gov/36961508/
- **`road-vo2-intervention-2024`** — Hebisz e Hebisz. *Greater improvement in aerobic capacity after a polarized training program including cycling interval training at low cadence (50-70 RPM) than freely chosen cadence (above 80 RPM)*. 2024. Ensaio com ciclistas mulheres bem treinadas; informa blocos de 4 minutos em alta intensidade, mas não sustenta copiar carga, cadência ou potência para todos os perfis. https://pubmed.ncbi.nlm.nih.gov/39536034/
- **`road-vo2-response-2024`** — Odden et al. *The higher the fraction of maximal oxygen uptake is during interval training, the greater is the cycling performance gain*. 2024. Intervenção observacional em ciclistas bem treinados; associa maior fração de VO₂max durante intervalos a ganhos de desempenho, sem definir dose universal. https://pubmed.ncbi.nlm.nih.gov/39385317/
- **`road-onbike-strength-2025`** — Barranco-Gil et al. *Off- and On-Bike Resistance Training in Cyclists: A Randomized Controlled Trial*. 2025. Ensaio com 37 ciclistas bem treinados durante 10 semanas; o grupo na bicicleta usou esforços máximos contra resistência muito alta e cadência muito baixa, com carga calibrada por força máxima. Houve melhora de força, potência e estrutura muscular, mas não de VO₂max; o resultado não sustenta uma sessão baseada somente em RPE nem uma liberação para perfis gerais. https://pubmed.ncbi.nlm.nih.gov/39231694/

## Critérios para os próximos protocolos

Antes de adicionar um protocolo ao motor, ele deverá passar por esta lista:

1. A modalidade e o objetivo estão explícitos?
2. Existe fonte adequada para o formato do estímulo?
3. Está claro se a fonte é de demanda, intervenção, revisão ou segurança?
4. A população estudada é compatível com o nível liberado?
5. A sessão continua executável apenas com RPE quando o atleta não tem sensor?
6. Existem limites de duração, progressão, recuperação e interrupção?
7. Dor, limitação ou recuperação ruim bloqueiam a sessão?
8. Há teste automatizado para garantir que o protocolo não seja escolhido fora do contexto?

Se uma resposta for “não”, a sessão permanece documentada como candidata e não entra no `rules-v1`.

## Avaliação dos próximos candidatos — 11 de setembro de 2026

### Candidato recomendado: taper pré-prova

O próximo protocolo a ser especificado é um taper pré-prova orientado por evento. A evidência é mais transferível para o produto do que uma sessão exclusiva de gravel/XCM: uma meta-análise de esportes de endurance encontrou melhora de desempenho após taper, com redução progressiva do volume e manutenção da intensidade e da frequência em estratégias de até 21 dias; isso é uma faixa de evidência, não uma dose universal para cada atleta. [Wang et al., 2023](https://journals.plos.org/plosone/article?id=10.1371/journal.pone.0282838)

Um estudo de 2025 em ciclistas bem treinados observou que uma redução de aproximadamente 50% do volume por duas semanas, mantendo a intensidade, preservou a maior parte das adaptações de desempenho. Os próprios autores ressaltam que a resposta depende do volume inicial, da intensidade e do período de treinamento, e que o desempenho submáximo pode ser afetado. [Lange et al., 2025](https://doi.org/10.14814/phy2.70302)

O estudo não sustenta criar uma sobrecarga automática antes da redução: em ciclistas treinados, duas semanas de intensificação elevaram a carga e pioraram temporariamente desempenho, humor e equilíbrio recuperação-estresse; após o taper, as medidas retornaram à linha de base, sem benefício acima dela. [Effect of intensified training on cognitive function, psychological state & performance in trained cyclists](https://pubmed.ncbi.nlm.nih.gov/35771645/)

**Limites para o Cadência:**

- a especificação deverá atuar no plano e na distribuição de volume, não apenas renomear uma sessão;
- não haverá bloco automático de sobrecarga antes do taper;
- intensidade preservada não significa sprint, teste máximo ou meta rígida de potência;
- o taper deverá depender de evento futuro válido, janela temporal explícita, dados mínimos e ausência de dor, limitação ou necessidade de recuperação;
- perfis sem evidência suficiente continuarão em sessões regulares ou protegidas;
- o motor prescritivo `rules-v1` e suas travas continuam sendo a autoridade durante a validação.

### Candidatos adiados

Gravel e XCM permanecem como contexto de endurance, sem protocolo próprio nesta rodada. O estudo de campo disponível para gravel descreve hidratação e perda de massa em uma prova, mas não valida uma sessão de treinamento; portanto, não sustenta transformar terreno ou distância em séries automáticas. [Fluid Intake and Hydration Responses to Mass Participation Gravel Cycling](https://pubmed.ncbi.nlm.nih.gov/39807388/)

No XCM, os indicadores aeróbicos e intermitentes se relacionam ao desempenho de prova, mas o estudo é de demanda/predição e não um ensaio de prescrição. Ele pode orientar especificidade futura, não definir uma dose para o catálogo. [Predictive ability of a comprehensive incremental test in mountain bike marathon](https://pubmed.ncbi.nlm.nih.gov/29387445/)

Força complementar, calor e restrição de fluxo sanguíneo não entram no motor geral nesta etapa. Sprint/pista/BMX e downhill/enduro estão fora do produto, não são módulos futuros do Cadência e não devem ser reintroduzidos como opções de modalidade.

### Implementação local do taper — 11 de setembro de 2026

O primeiro candidato orientado por evento foi implementado localmente como uma alteração de volume do plano, sem criar uma modalidade nova. A avaliação fica no snapshot em `prescription_snapshot.event_taper`, com versão `taper-v1`, escopo `event_based_plan_volume` e modo prescritivo somente quando todos os gates passam.

- exige evento futuro válido entre 7 e 21 dias da geração, nível avançado, avaliação submáxima apta, pelo menos 8 semanas de treino e 3 pedais semanais informados;
- não se aplica com limitação ativa, dor ou necessidade recente de recuperação;
- reduz em 50% a duração das sessões elegíveis, preservando a frequência e o RPE-alvo, com mínimo de 20 minutos;
- atua somente em sessões anteriores ao evento e dentro de 14 dias da prova; não altera o dia do evento nem a quarta semana de recuperação;
- mantém `rules-v1` como motor prescritivo e não cria sobrecarga automática antes do taper;
- registra `taper-cyclist-2025` e `taper-meta-2023` nas sessões afetadas; a migração `000017` também registra `taper-overreach-cyclists-2023` como limite contra intensificação automática.

Os testes cobrem evento próximo e distante, evento passado, nível/histórico insuficiente, proteção por dor, semana de recuperação e determinismo. A migração `000017` foi aplicada somente no PostgreSQL de desenvolvimento local; a produção permanece na `000016`. A versão local visível passou para `0.13.0` e a tela de novidades foi atualizada, mas a validação do proprietário e qualquer publicação ainda estão pendentes.

### Novo piloto local: intervalos VO₂max de estrada — 12 de setembro de 2026

O protocolo `road_vo2_intervals`, apresentado como **Intervalos VO₂max de estrada**, foi adicionado somente ao checkout local como uma nova opção explícita de preferência. Ele usa quatro blocos de 4 minutos em RPE 8, com 4 minutos leves entre os blocos, aquecimento e desaquecimento; a estrutura não prescreve sprint máximo, cadência baixa, potência fixa ou frequência cardíaca como substituto de VO₂max.

O motor só o seleciona quando a disciplina informada é `road`, o atleta é avançado, a avaliação submáxima está apta, o objetivo é `performance` ou `event`, há pelo menos oito semanas de treino recente, três pedais semanais informados, 60 minutos disponíveis, semana de construção e preferência explícita `vo2max`. Limitação, dor, recuperação insuficiente, evento fora da fase específica, níveis menores ou outra modalidade impedem a escolha; as proteções do `rules-v1` continuam prioritárias.

A sessão é uma adaptação conservadora dos blocos de 4 minutos observados em ciclistas mulheres bem treinadas e da associação entre maior fração de VO₂max durante intervalos e ganhos de desempenho. A evidência orienta o formato e a população do piloto, mas não valida a dose para todos os atletas. A migração `000018` registra as duas fontes somente no banco local; produção continua em `000016`, sem deploy ou mudança de infraestrutura. A suíte automatizada, o build e a validação ponta a ponta passaram: a preferência foi salva no navegador e um plano local apresentou **Intervalos VO₂max de estrada**. O piloto continua fora da produção até revisão, backup e autorização explícita.

### Piloto local em validação: intervalos curtos autorregulados — 12 de setembro de 2026

O piloto local avalia um protocolo de intervalos curtos autorregulados para ciclismo de estrada ou indoor, separado do piloto de VO₂max e sem prescrever sprint máximo. Um ensaio randomizado de 2025 comparou seis semanas de `4–8 × 30 s` com 120 segundos de recuperação e `6–10 × 1 min` com 1 minuto de recuperação, três vezes por semana, em 82 adultos anteriormente inativos; os dois formatos melhoraram o VO₂peak de forma semelhante. [Hesketh et al., 2025](https://www.frontiersin.org/journals/physiology/articles/10.3389/fphys.2025.1484722/full)

Em ciclistas de elite, um estudo anterior encontrou vantagem de um formato de 30 segundos sobre blocos de 5 minutos em alguns indicadores de desempenho após três semanas, mas a amostra era pequena, altamente treinada e o protocolo usava esforço intenso repetido. [Rønnestad et al., 2020](https://pubmed.ncbi.nlm.nih.gov/31977120/)

Uma revisão com meta-análise publicada em 2025 reuniu HIIT, SIT e repeated-sprint training em atletas de perfis diversos e encontrou melhora de VO₂max, mas também heterogeneidade de população, modalidade, duração e recuperação. Ela deve orientar a comparação dos formatos, não definir a prescrição do Cadência. [Revisão sistemática e meta-análise de 2025](https://pubmed.ncbi.nlm.nih.gov/40605061/)

**Decisão de publicação:** manter a versão curta e autorregulada com RPE e controle técnico, sem liberar `all-out` para perfis gerais. A especificação foi transformada em implementação local com modalidade explícita, nível e histórico mínimos, avaliação apta, disponibilidade, seis blocos máximos, recuperação, uma única sessão de qualidade, travas por dor/fadiga/recuperação e testes de não seleção. A validação automatizada e manual local passou: a preferência foi salva e o plano regenerado apresentou **Intervalos curtos autorregulados**. O piloto está no `rules-v1` local, recebeu a migração `000019` e a versão `0.15.0`, mas ainda não foi publicado.

A especificação preliminar foi implementada localmente como `short_self_regulated_intervals`, com seis repetições de 1 minuto em RPE 7–8, recuperação leve autorregulada, aquecimento/desaquecimento, nível avançado, estrada ou indoor e gates de segurança. Ela continua sendo uma hipótese operacional, não uma conclusão científica nem uma prescrição liberada em produção. A migração `000019` registra as fontes do piloto, e a suíte automatizada, o `go vet`, o build e a aplicação local passaram. A validação manual confirmou a persistência da preferência e a apresentação do protocolo no plano regenerado; a nota `V0.15.0` também aparece no navegador. O piloto continua fora da produção até revisão e autorização explícita.

### Próximo candidato em avaliação: resistência específica na bicicleta — 12 de setembro de 2026

O ensaio randomizado de Barranco-Gil et al. (2025) comparou, em 37 ciclistas bem treinados, dez semanas de treinamento de resistência fora da bicicleta com esforços na própria bicicleta. O protocolo na bicicleta usou resistência muito alta, cadência muito baixa e carga calibrada por aproximadamente 70% da força dinâmica máxima, em duas sessões semanais. Os grupos de resistência melhoraram força, potência e alguns marcadores estruturais; não houve melhora de VO₂max.

Esse resultado é relevante para investigar um estímulo de resistência específico do ciclismo, mas não autoriza copiar esforços máximos para o público do Cadência. O produto não possui medição de força dinâmica máxima, calibração de carga ou controle suficiente para verificar a execução; RPE isolado não é equivalente à carga estudada. A amostra também era de ciclistas bem treinados, o que limita a transferência para iniciantes, intermediários e atletas retornando após pausa.

**Decisão nesta etapa:** registrar `road-onbike-strength-2025` como evidência e manter o candidato fora do `rules-v1`, sem preferência, protocolo, migração ou nota de versão. Para uma futura implementação serão necessários pré-requisitos de medição, critérios de elegibilidade avançada, limites de resistência/cadência, proteção para dor e uma matriz de não seleção. Gravel/XCM continuam contextos de endurance sem protocolo próprio; sprint de pista/BMX e downhill/enduro estão fora do produto.

### Publicação do catálogo ampliado — 12 de setembro de 2026

Após a validação local, os pilotos `taper-v1`, `road_vo2_intervals` e `short_self_regulated_intervals` foram publicados na produção com as migrações `000017`, `000018` e `000019`. A decisão de escopo também foi publicada: sprint/pista/BMX e downhill/enduro não são modalidades do Cadência e não devem ser reintroduzidos no perfil, nas preferências ou no motor. A versão comunicada ao usuário passou para `0.16.0` e foi registrada na release [v0.16.0](https://github.com/Saulorangel87/App-de-treino/releases/tag/v0.16.0).
