# Avaliação: o pedal de referência

A aba `/avaliacao` guarda um **pedal de referência**: de 15 a 30 minutos contínuos em Z2 (20 é o ideal), feito descansado e sem esforço máximo. Não é exame médico nem teste de limite.

## O que o resultado faz

1. **Apto a progredir.** A regra é a mesma desde a primeira versão: sem dor relatada, pelo menos 18 minutos e esforço até RPE 6 (zona Z3). O "apto" é só um sim ou não e vale o último resultado. Ele libera os treinos de qualidade (intervalados, limiar, VO2) quando o perfil é avançado e a meta é desempenho ou prova, entra no polimento (taper) pré-prova e nas auditorias do motor. Para os outros perfis o plano não muda.
2. **Evolução.** Com números do pedal, o app calcula e compara uma avaliação com a anterior. Os números **nunca** mudam o "apto".
3. **Quando refazer.** A tela sugere refazer a partir de 4 semanas (`REASSESS_AFTER_DAYS = 28` em `frontend/lib/assessment.ts`) e diz ao atleta "4 a 6 semanas", nas mesmas condições.

## Números opcionais

| Campo | Limites | Para que serve |
| --- | --- | --- |
| FC média | 30 a 250 bpm | eficiência |
| Potência média | 0 a 2000 W | eficiência (com FC) |
| Distância | 0 a 500 km, no máximo 80 km/h para a duração | eficiência (com FC, sem potência) |
| FC média da 1ª e da 2ª metade | 30 a 250 bpm, sempre as duas | deriva |

Os valores derivados são calculados na leitura (`Assessment.Derive` em `backend/internal/athlete/assessment.go`) e não são gravados:

- **Eficiência aeróbica:** `potência ÷ FC` (W por bpm) quando há potência; senão `velocidade ÷ FC × 100` (km/h a cada 100 bpm). A FC é a média informada, ou a média das duas metades. Só se compara com uma avaliação do **mesmo tipo**. Mais alto é melhor condição.
- **Deriva de FC:** `(FC 2ª metade − FC 1ª metade) ÷ FC 1ª metade × 100`. Até 3% é estável, até 5% é o esperado em Z2, acima disso o ritmo pode ter ficado forte para a zona, ou houve calor, pouca água ou cansaço. É informativo.

Os limiares da deriva e a leitura da eficiência são critérios do Cadência, não resultado de estudo; a base científica do pedal em si segue sendo Dunbar, Kalinski e Robertson (1996).

## Preencher com um pedal importado

Quando há atividades importadas (arquivo `.fit` ou `.gpx`) de 15 a 30 minutos, a tela oferece preencher duração, FC, potência e distância a partir de uma delas. As duas metades da FC não vêm do arquivo (a leitura por trechos é a Fase C das zonas) e ficam para o atleta.

## API e banco

- `POST /v1/assessments/submaximal`: aceita `actual_zone` (1 a 5; vence o RPE) ou `actual_rpe` (clientes antigos) e os números opcionais. A zona vira o RPE da tabela de zonas, como nos treinos.
- `GET /v1/assessments`: últimas 10, da mais recente para a mais antiga, com `efficiency` e `heart_rate_drift_percent`.
- `GET /v1/assessments/current`: a mais recente (continua existindo).
- Migração `000037_assessment_metrics`: cinco colunas opcionais em `cycling_assessments` e a regra de que as duas metades andam juntas. Avaliações antigas ficam sem números.

## Fora desta etapa

Testes guiados de FC máxima, limiar de FC (LTHR) e FTP (Fase B das zonas). São esforços fortes: pedem avisos de segurança e uma decisão de produto antes.
