# Proteção do motor: cenários, comportamento atual e alvo

Etapa 1 de 5 da flexibilização da proteção do `rules-v1`. Este documento não altera
nenhum comportamento: registra o que o motor faz hoje (fixado em
`backend/internal/planning/protection_scenarios_test.go`) e propõe o alvo para as
próximas etapas. Os critérios da coluna "Alvo" dependem de aprovação do produto.

## Como a proteção funciona hoje

- `Generate` tira uma foto do histórico dos últimos 28 dias e monta o ciclo de 4
  semanas (11 treinos no cenário de teste) de uma vez.
- `RequiresRecovery()` liga com **dor em qualquer sessão concluída**, fadiga média
  pós-treino ≥ 4 ou fadiga média de check-in ≥ 4.
- Com **dor**, todos os treinos do ciclo viram "Giro leve protegido" (RPE 3,5,
  máximo 45 min, carga ×0,8). Com **fadiga alta**, só os treinos de qualidade são
  trocados; os demais ficam como estão.
- Nada reavalia os treinos futuros depois da geração. A adaptação pós-treino
  (`DecideAdaptation`) roda apenas como shadow.
- O motor recebe só agregados (`PainReported`, médias), sem datas. Por isso ele
  não distingue dor de ontem de dor de 27 dias atrás, nem sabe se houve melhora
  depois do sinal.

## Cenários

Números do "Hoje" vêm do teste de caracterização (ciclo de 11 treinos, atleta
avançado de estrada).

| # | Cenário | Hoje | Alvo proposto |
|---|---------|------|---------------|
| 1 | Sem sinais | Plano normal (2 treinos de qualidade). | Sem mudança. |
| 2 | Dor ontem | 11/11 protegidos, RPE ≤ 3,5, ≤ 36 min, por 28 dias. | Nível **forte** por 3 dias; depois **moderado** até o 7º dia; depois normal. Um check-in posterior sem dor e com fadiga ≤ 2 (a partir do 3º dia) baixa um nível. |
| 3 | Dor há 20 dias, sem recorrência | Idêntico ao cenário 2: 11/11 protegidos. | **Nenhuma** proteção (passou de 7 dias). |
| 4 | Dor recorrente: 2 ou mais registros em 14 dias | Igual ao cenário 2. | Nível **forte** por 7 dias, com aviso para procurar avaliação profissional se a dor persistir. |
| 5 | Dor há 5 dias, já com 2 check-ins bons depois | 11/11 protegidos. | Nível **leve**: mantém os tipos de treino, duração −10% e RPE −1 nos de qualidade. |
| 6 | Um dia de fadiga 5 entre 5 sessões boas (média 2,6) | Sem proteção (média abaixo de 4). | Sinal **leve** de 1 a 2 sessões, sem trocar o tipo de treino. |
| 7 | Fadiga média 4,0 com poucos registros (2 sessões: 5 e 3) | 2/11 treinos de qualidade trocados, por 28 dias. | Nível **leve**; só sobe para **moderado** se houver um segundo sinal nos últimos 7 dias. |
| 8 | Fadiga de check-in 4,0 e melhora nos 2 check-ins mais recentes | 2/11 trocados. | **Nenhuma** proteção; o registro mais recente pesa mais que a média. |
| 9 | Fadiga alta persistente: 3 sessões seguidas com fadiga ≥ 4 | 2/11 trocados. | Nível **moderado**: troca os treinos de qualidade por aeróbico leve até a fadiga cair. |
| 10 | Usuário se sente recuperado ("Estou recuperado") sem dor recente | Não existe. | Reavalia na hora; se não houve dor nos últimos 7 dias, encerra proteção leve e moderada. Com dor recente, mantém um nível abaixo. |
| 11 | Treino já concluído, iniciado ou editado pelo usuário | Nunca é alterado depois de salvo. | Continua intocado pela reavaliação. |
| 12 | Limitação ativa ou restrição médica no perfil | Proteção permanente do perfil, independente do histórico. | Sem mudança; a reavaliação nunca a afrouxa. |

## Níveis propostos

| Nível | Efeito |
|-------|--------|
| Nenhuma | Plano normal. |
| Leve | Mantém os tipos de treino; duração −10%; RPE alvo −1 nos treinos de qualidade. |
| Moderada | Troca apenas os treinos de qualidade por aeróbico leve; demais mantêm duração −10%. |
| Forte | "Giro leve protegido" (RPE 3,5, até 45 min) em todos os treinos afetados. |

## Dados que faltam hoje (para a etapa 2)

O motor precisa de sinais **com data**. Hoje o contexto traz só agregados de 28 dias.
A etapa 2 adiciona ao contexto, a partir das mesmas tabelas (`feedback`,
`recovery_data`), os registros recentes com data: dor, fadiga pós-treino e fadiga de
check-in dos últimos 14 dias. Nenhuma migração é necessária.

A dor é hoje um sim/não, sem intensidade. Por isso "dor leve" não é distinguível e a
proposta acima usa o **tempo desde a dor e os registros posteriores** para graduar a
proteção. Se quiser distinguir intensidade, é preciso um campo novo no formulário
pós-treino (decisão de produto, fora desta etapa).

## Pontos para o produto decidir

1. Prazos: dor forte por 3 dias, moderada até o 7º, nenhuma depois. Aceita?
2. Recorrência: 2 ou mais dores em 14 dias mantêm proteção forte por 7 dias e
   sugerem avaliação profissional.
3. Um check-in bom depois da dor (a partir do 3º dia) rebaixa um nível.
4. A reavaliação atua na semana atual e na seguinte, só em treinos ainda planejados.
