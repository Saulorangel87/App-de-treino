# Ciclo de vida dos planos de treino

Última revisão: 31 de agosto de 2026.

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
