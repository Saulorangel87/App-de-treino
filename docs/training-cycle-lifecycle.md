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

Um plano concluído continua disponível na consulta do plano atual até que um novo rascunho seja gerado. A interface apresenta a ação **Gerar próximo ciclo**.

O início do novo ciclo é calculado assim:

1. próxima segunda-feira em relação à data atual;
2. primeira segunda-feira após o término do último ciclo concluído;
3. utiliza-se a data que ocorrer por último.

Isso evita sobreposição entre ciclos e não apaga o histórico anterior. Depois de gerado, o novo plano permanece como rascunho até o usuário revisá-lo e aceitá-lo.
