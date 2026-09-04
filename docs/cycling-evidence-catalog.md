# Mapa de evidências do catálogo de ciclismo

Última revisão: 4 de setembro de 2026.

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

**Uso previsto no Cadência:** um futuro protocolo de esforços variáveis, inicialmente experimental e restrito a atletas com contexto XCO explícito, nível suficiente, avaliação apta e ausência de sinais de recuperação insuficiente.

**Limite:** a distribuição observada em uma prova não será convertida diretamente em séries universais. A parte técnica permanece instrução de habilidade e não deve ser simulada automaticamente pelo aplicativo.

### XCM e gravel

As provas combinam distância, elevação, terreno variável e oportunidades limitadas de alimentação. A evidência específica para prescrição de sessões ainda é menor que a de estrada e XCO.

**Uso previsto no Cadência:** reutilizar endurance e progressão de volume já validados, acrescentando o contexto de terreno somente quando houver dados suficientes.

**Limite:** não criar um protocolo “gravel” apenas por trocar o nome de uma sessão de estrada. Nutrição e hidratação devem permanecer orientações separadas, sem metas automáticas baseadas em um único estudo de campo.

### Downhill e enduro

O desempenho depende de técnica, controle corporal, força isométrica e tolerância à fadiga de pegada, além do condicionamento. Estudos recentes também mostram risco relevante de lesão em treinamento e prova.

**Uso previsto no Cadência:** não liberar descidas, saltos ou treinos técnicos como prescrição automática na primeira expansão.

**Limite:** qualquer futuro módulo deverá separar condicionamento físico, habilidade técnica, proteção e avaliação profissional. O motor de endurance atual não é suficiente para prescrever DH/enduro com segurança.

### Pista sprint e BMX

São modalidades com exigências anaeróbicas, neuromusculares e de força muito diferentes das sessões de endurance do MVP. A literatura de velocistas de pista descreve grande concentração de carga nas zonas mais intensas, mas isso não sustenta reutilizar o catálogo atual.

**Uso previsto no Cadência:** manter fora da primeira expansão e tratar como futuro produto específico, com avaliação e regras próprias.

## Protocolos candidatos

| Candidato | Modalidade | Situação da evidência | Liberação inicial |
| --- | --- | --- | --- |
| Endurance de estrada | `road` | Base de endurance e distribuição de intensidade apoiadas por revisões | Todos os níveis, com progressão conservadora |
| Intervalos moderados | `road` | Ensaios recentes em ciclistas bem treinados | Piloto local: intermediário/avançado, objetivo compatível, avaliação apta, 60 min disponíveis e sem sinais de recuperação insuficiente |
| Intervalos intensos controlados | `road` | Revisões e ensaios, com maior exigência fisiológica | Avançado, avaliação apta e objetivo compatível |
| Esforços variáveis XCO | `mtb_xco` | Forte evidência de demanda; evidência de intervenção ainda limitada | Futuro piloto restrito, não padrão geral |
| Endurance gravel/XCM | `gravel`, `mtb_xcm` | Evidência direta de prescrição ainda insuficiente | Usar somente base/endurance contextual |
| Força complementar | `road`, `mtb_xco` | Meta-análise recente favorável, mas com baixa certeza | Módulo opcional e separado do treino de bike |
| Downhill/enduro técnico | `dh_enduro` | Evidência de risco, não de protocolo automatizado seguro | Bloqueado nesta fase |
| Sprint de pista/BMX | `track_sprint` | Modalidade distinta do MVP | Bloqueado nesta fase |

As situações acima são decisões de produto provisórias. Antes de transformar qualquer candidato em regra do `rules-v1`, seus parâmetros, população-alvo e critérios de interrupção devem ser revisados por profissional habilitado.

### Primeiro piloto local: intervalos moderados de estrada

O primeiro protocolo específico implementado localmente é `road_moderate_intervals`, apresentado como **Intervalos moderados de estrada**. Ele usa três blocos de 10 minutos com três minutos de recuperação leve, alvo RPE 6 e uma sessão de qualidade por semana. Essa dose é uma adaptação conservadora do contexto dos estudos, não a reprodução do bloco de seis sessões em sete dias.

O motor só o seleciona quando a disciplina é explicitamente `road`, o atleta é intermediário ou avançado, a avaliação submáxima está apta, o objetivo é performance ou evento, há pelo menos 60 minutos disponíveis e a preferência está vazia ou indica intervalos. Dor, limitação ou recuperação insuficiente substituem o protocolo por uma sessão protegida.

## Alterações de modelo necessárias antes do catálogo

O contexto agora guarda `bike_type`, `terrain` e uma disciplina explícita, opcional e validada. A disciplina não é inferida pelo tipo de bicicleta: XCO, gravel ou pista só podem ser usados quando o atleta os informa diretamente.

Valores planejados para `cycling_context.discipline`:

- `general`: ciclismo sem disciplina informada;
- `road`: estrada;
- `mtb_xco`: MTB cross-country olímpico;
- `mtb_xcm`: MTB maratona;
- `gravel`: gravel;
- `indoor`: indoor/rolo;
- `dh_enduro`: downhill/enduro;
- `track_sprint`: pista sprint/BMX.

Quando o campo estiver vazio ou for `general`, somente protocolos gerais e já existentes poderão ser selecionados. Perfis antigos não devem ser migrados automaticamente para uma modalidade específica.

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
- **`gravel-field-2024`** — *Fluid Intake and Hydration Responses to Mass Participation Gravel Cycling*. 2024. Estudo de campo sobre gravel; informa contexto de distância e hidratação, não valida sozinho um protocolo de treino. https://pubmed.ncbi.nlm.nih.gov/39807388/
- **`dh-injury-2024`** — Fallon et al. *Downhill race for a rainbow jersey: the epidemiology of injuries in downhill mountain biking at the 2023 UCI cycling world championships*. 2024. Estudo observacional de lesões; usado como trava de segurança, não como prescrição. https://pubmed.ncbi.nlm.nih.gov/39411021/
- **`mtb-crash-mechanisms-2025`** — Bonte et al. *Injury Mechanisms in Mountain Biking: A Systematic Video Analysis of 534 Cases*. 2025. Estudo de mecanismos de queda; reforça que habilidade técnica e prevenção não devem ser reduzidas a carga aeróbica. https://pubmed.ncbi.nlm.nih.gov/40534393/
- **`track-sprint-load-2023`** — *Training load and intensity distribution for sprinting among world-class track cyclists*. 2023. Descrição de treinamento de velocistas de pista; modalidade fora do escopo do primeiro catálogo ampliado. https://pubmed.ncbi.nlm.nih.gov/36961508/

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
