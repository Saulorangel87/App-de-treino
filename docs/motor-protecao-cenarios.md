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

## Etapa 2: implementada, ainda desligada

`backend/internal/planning/protection.go` (`assessProtection`) é uma função pura de
sinais datados e da data de hoje. Ela **não é chamada** pelo gerador de planos nem
pelo repositório; o `rules-v1` continua como está. Cada cenário da tabela acima tem
um caso em `protection_test.go`, e um teste compara com a regra atual.

Regras aplicadas, em dias inteiros (UTC):

- **Dor isolada:** forte nos dias 0 a 2, moderada do 3º ao 7º dia, nenhuma depois.
- **Dor recorrente** (2 ou mais registros em 14 dias, o mais recente com menos de 7
  dias): forte, com recomendação de avaliação profissional; não é rebaixada por
  check-in posterior.
- **Check-in bom** (sem dor, fadiga 1 a 2) feito pelo menos 3 dias depois da dor
  rebaixa um nível, com piso "leve" enquanto a dor tem até 7 dias.
- **Fadiga alta** (≥ 4) nos últimos 7 dias: leve com um registro; moderada com dois
  ou mais, ou com 3 sessões seguidas de fadiga alta.
- **Melhora posterior:** um registro mais novo que a fadiga alta e com fadiga ≤ 2
  anula o sinal de fadiga.
- O nível final é o maior entre dor e fadiga, e `expires_on` é a data em que todos os
  sinais deixam de valer, se não houver novos registros.

Ajustes em relação à tabela de cenários, por decisão de implementação:

- **Cenário 6:** um dia de fadiga 5 já seguido de sessões boas não protege nada.
  Protege levemente apenas se for o registro mais recente.
- **Cenário 7:** fadiga 5 e depois 3 fica "leve", porque 3 ainda não conta como
  melhora.

Ainda não feito (etapa 3): ler esses sinais do banco (`feedback` e `recovery_data`,
últimos 14 dias) e usá-los para reavaliar os treinos futuros.

## Etapa 3a: o gerador entende os níveis (ainda sem efeito no app)

- `Context` ganhou `RecentSignals` e `Protection`. Com `Protection` nulo, o gerador
  segue exatamente a regra antiga (testes existentes inalterados).
- Com `Protection` preenchido: **forte** protege todos os treinos; **moderada** troca
  só os de qualidade por giro leve protegido e encurta os demais em 10%; **leve**
  encurta em 10% e reduz em 1 ponto o RPE dos de qualidade; **nenhuma** não altera
  nada. Limitações e restrição médica do perfil continuam mandando sempre.
- Cada treino novo grava em `explanation.prescription_inputs` os parâmetros que o
  geraram (tipo, semana, multiplicador, dia da semana). `ReprescribeWorkout` usa isso
  para reconstruir um treino futuro sem regenerar o plano, e voltar ao treino original
  quando a proteção some (coberto por teste). Planos já gerados não têm o campo e não
  são reavaliados; para eles vale gerar um plano novo.
- O nível aplicado fica em `explanation.protection` (nível, motivos, data de fim).

Falta (etapa 3b): ler os sinais do banco, reavaliar os treinos planejados da semana
atual e da seguinte ao concluir treino ou registrar check-in, e a chave de
configuração que liga tudo (desligada por padrão).

## Etapa 3b: leitura dos sinais, reavaliação e chave de configuração

Tudo atrás de `PROTECTION_LEVELS_ENABLED` (padrão `false`). Desligada, o app segue a
regra antiga de 28 dias e nenhum treino é reavaliado; as duas consultas novas de
leitura rodam, mas o resultado não é usado.

- **Sinais datados:** `recentSignalsByProfileID` lê os últimos 14 dias de sessões
  concluídas (dor e fadiga pós-treino) e de check-ins (fadiga).
- **Dor sempre conta.** A regra antiga de 28 dias descarta sessões sem dados mínimos
  (por exemplo, sem duração). Para a dor isso seria remover uma proteção, então os
  sinais datados contam a dor mesmo nessas sessões; só a **fadiga** segue o filtro de
  integridade.
- **Quando reavalia:** ao concluir ou corrigir um treino e ao salvar o check-in diário,
  de forma "melhor esforço": se a reavaliação falhar, a operação principal não falha e
  o próximo gatilho reavalia a partir dos registros gravados.
- **O que reavalia:** treinos `planned` do plano **ativo**, de hoje até o domingo da
  semana seguinte. Reconstrói cada um com os parâmetros gravados e só grava quando algo
  mudou.
- **O que nunca é tocado:** treinos concluídos, iniciados, cancelados ou ajustados
  pelo check-in (`adapted`); plano em rascunho; treino de outro atleta; treino gerado
  antes de existirem os parâmetros; e os 1 ou 2 próximos treinos que o **gatilho do
  banco** (`feedback_adapts_future_workouts`) já reduziu após o feedback
  (`explanation.adaptation`).

### Correção sobre a adaptação que existe hoje

Antes eu disse que a adaptação pós-treino só rodava como shadow. Isso vale para a
função Go `DecideAdaptation`. A adaptação de curto prazo **ativa** vive no banco: o
gatilho `feedback_adapts_future_workouts` reduz os próximos 1 a 2 treinos planejados
(dor: −20% e RPE ≤ 3; esforço 9 ou fadiga 5: −20%; acima do esperado: −10%). Esse
gatilho só age quando o treino concluído tem dados mínimos. A reavaliação convive com
ele: não sobrescreve esses treinos, porque a redução dele pode ser mais forte que o
nível leve.

### Validação

Fluxo real (API compilada, `PROTECTION_LEVELS_ENABLED=true`, PostgreSQL local): concluir
treino com dor protege os treinos da janela; o gatilho e a reavaliação não se
sobrepõem; ao sumir o sinal, um check-in restaura os treinos; treino fora da janela,
concluído e adaptado permanecem intactos. Fixture
`database/tests/protection_reevaluation.sql` cobre as regras das duas consultas.

### Como ligar e desligar

Defina `PROTECTION_LEVELS_ENABLED=true` no ambiente da API e recrie o contêiner da API
para ligar; volte a `false` e recrie para restaurar a regra antiga. Treinos já
reescritos ficam como estão até a próxima reavaliação ou até gerar um plano novo.
