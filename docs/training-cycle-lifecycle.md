# Ciclo de vida dos planos de treino

Última revisão: 11 de setembro de 2026.

## Estados

- `draft`: plano calculado e aguardando aprovação.
- `active`: ciclo aprovado com pelo menos uma sessão pendente ou em andamento.
- `completed`: todas as sessões terminaram como concluídas ou puladas.
- `cancelled`: ciclo ativo substituído explicitamente por outro plano.

## Encerramento automático

O PostgreSQL observa mudanças no estado dos treinos. Quando o último treino pendente de um plano ativo passa para `completed` ou `skipped`, o plano é marcado como `completed` na mesma transação. O histórico de planos, treinos, sessões e feedback permanece armazenado.

A migração `000006_training_plan_completion` também identifica planos antigos que já não possuem sessões pendentes.

## Próximo ciclo

Um plano concluído continua disponível na consulta do plano atual até que um novo rascunho seja gerado. A interface apresenta a ação **Gerar próximo ciclo**. Ao atualizar a disponibilidade, o usuário também pode gerar um rascunho revisado a partir da semana corrente.

O início do novo ciclo é calculado assim:

1. segunda-feira da semana corrente, ou a próxima segunda-feira quando a geração acontece no domingo;
2. sessões da semana corrente que já ficaram no passado não são criadas.

Isso mantém o plano útil no momento da geração e não apaga o histórico anterior. Depois de gerado, o novo plano permanece como rascunho até o usuário revisá-lo e aceitá-lo.

## Treino não realizado

Depois que a data planejada passou, o atleta pode usar **Não realizei** em um treino que ainda esteja `planned` ou `adapted`. A API `POST /v1/workouts/{workoutID}/missed` registra a decisão alterando o treino para `skipped`, sem criar uma `workout_session` fictícia.

Essa ação é uma classificação explícita do registro, não uma inferência automática. Ela não desloca o treino para outro dia, não cria uma sessão substituta e não aumenta nem reduz a carga seguinte. O histórico passa a contar o treino como encerrado e não realizado; a evolução continua mostrando somente sessões que realmente tiveram início.

Treinos futuros não podem ser marcados dessa forma. Cancelar uma sessão já iniciada continua sendo uma ação diferente, pelo fluxo **Cancelar**, e também mantém o status final `skipped`.

## Início de treino futuro e desfazer registro

**Treino futuro não inicia.** `POST /v1/workouts/{workoutID}/start` só aceita treinos agendados para hoje ou para uma data passada; um treino futuro recebe `409 workout_in_future`, e a tela mostra "Disponível no dia planejado" no lugar do botão. A tela envia a data local do atleta em `?date=AAAA-MM-DD`, porque o servidor trabalha em UTC e, à noite no Brasil, a data UTC já é a de amanhã. O servidor só aceita esse valor se estiver a no máximo um dia da data UTC; qualquer outro vira a data UTC, então um valor forjado não libera um treino futuro. Treinos passados ainda planejados continuam podendo ser iniciados (quem pedalou ontem e esqueceu de registrar).

**Desfazer registro.** `POST /v1/workouts/{workoutID}/undo` apaga o registro de um treino `completed`, `skipped` ou `in_progress` e o devolve ao plano. Em uma única transação: apaga a sessão (o feedback vai junto, em cascata); restaura os treinos seguintes que o gatilho de adaptação havia alterado por causa dessa sessão, usando a duração e o esforço anteriores que o próprio gatilho guardou em `explanation.adaptation`; volta o treino para `planned` (ou `adapted`, se ele já tinha ajuste); e reabre o plano, se encerrar aquele treino o havia concluído. Depois, o serviço reavalia a proteção, porque a dor e a fadiga daquela sessão deixam de contar. A tela mostra **Desfazer registro** no treino concluído e **Reabrir treino** no não realizado (por exemplo, depois de **Cancelar**), sempre com confirmação.

Limites: vale para o plano ativo ou para o último plano, quando ele foi concluído e não há outro ativo; um treino de plano antigo não pode ser desfeito. Treinos ajustados por um check-in antes da sessão (`pre_session_recovery`) não têm o ajuste revertido, porque os valores "anteriores" gravados descartariam também aquele ajuste. A atividade importada vinculada ao treino continua vinculada. O gatilho só ajusta treinos ainda `planned` e não ajusta de novo um que já foi ajustado, então cada treino tem no máximo um dono de ajuste, e desfazer uma sessão que não causou ajuste não toca nos ajustes de outra.

## Modo tarefa: marcar como feito

Além do cronômetro (**Iniciar treino**), o treino planejado ou adaptado de hoje ou de um dia passado tem **Marcar como feito**. É para quem pedala e registra depois, ou não quer abrir o app no pedal. O formulário é o mesmo da conclusão (zona, dificuldade, fadiga, dor e métricas opcionais), com dois campos a mais: a **duração** em minutos (já preenchida com o tempo planejado) e o **dia do treino**.

`POST /v1/workouts/{workoutID}/log` guarda uma sessão concluída com `duration_source = reported` (o cronômetro grava `timer`; a coluna também prevê `imported`, para arquivos do dispositivo). `started_at` é `completed_at` menos a duração informada. Hoje termina agora; um dia anterior é colocado ao meio-dia de Brasília (15:00 UTC). Depois disso a sessão segue o mesmo caminho da conclusão pelo cronômetro (`recordCompletion`): avaliação de integridade, comparação planejado e realizado, avaliação shadow e o gatilho de adaptação.

Regras:

- duração de 1 a 720 minutos; dia entre hoje e 7 dias atrás, e nunca anterior ao dia planejado do treino (adiantar um treino continua bloqueado, assim como iniciar um treino futuro);
- o treino precisa estar `planned` ou `adapted`; um treino em andamento se conclui pelo cronômetro, e quem marcou por engano usa **Desfazer registro**;
- vale o bloqueio de segurança de uma limitação ativa, como no início do treino;
- erros: `400 invalid_workout_log` (duração ou dia fora das regras), `409 workout_in_future`.

**A validade não depende de cumprir o tempo planejado.** O que torna uma sessão elegível para o histórico é a coerência dos dados (duração maior que zero, escalas válidas), e não o percentual feito. O que o percentual muda é só a progressão: a regra do gatilho (migração `000036`) exige **pelo menos 80% do tempo planejado** para subir a carga do treino seguinte. Reduções por dor, fadiga ou esforço acima do esperado continuam valendo com qualquer duração. O percentual de 80% é escolha de produto, e o app diz isso na tela.

**Por que guardar a origem da duração.** A etapa 2 do motor (calibrar com dados reais) deve preferir duração medida pelo cronômetro ou por arquivo; a duração digitada pelo atleta é menos confiável. A coluna permite filtrar. A tela mostra "duração informada por você" no treino concluído, e a planilha de dados traz a coluna **Origem da duração**.

O ciclo de vida permanece baseado no motor determinístico `rules-v1`. A observação de feedbacks reais e do resumo semanal é a próxima etapa antes de qualquer mudança na transição de estados ou na progressão automática.
