# Adaptação do plano após o treino

## Objetivo

O aplicativo usa o feedback enviado ao concluir uma sessão para fazer ajustes pequenos nas próximas sessões ainda planejadas. A adaptação acontece na mesma transação que salva o feedback no PostgreSQL: se qualquer parte falhar, nenhuma alteração parcial é mantida.

O mecanismo é uma regra conservadora de produto, não um diagnóstico nem uma prescrição médica individualizada. RPE, dificuldade e fadiga são sinais subjetivos e devem ser interpretados junto com o contexto do atleta.

## Regras da versão 1

- **Dor relatada:** reduz em 20% a duração das duas próximas sessões e limita o RPE-alvo a 3. A interface recebe também um aviso para interromper a atividade se a dor reaparecer e buscar avaliação profissional se ela for persistente ou intensa.
- **Esforço ou fadiga muito altos:** quando o RPE real é pelo menos 9, a fadiga é 5 ou a dificuldade é `very_hard`, reduz em 20% a duração das duas próximas sessões, diminui o RPE-alvo em 1 e o limita a 4.
- **Carga acima do esperado:** quando o RPE real supera o alvo em pelo menos 2, a fadiga é 4 ou a dificuldade é `hard`, reduz em 10% a duração da próxima sessão e diminui o RPE-alvo em 1.
- **Resposta claramente fácil:** quando o RPE real fica pelo menos 2 pontos abaixo do alvo, a fadiga é no máximo 2 e a dificuldade é `easy` ou `very_easy`, aumenta somente 5% da duração da próxima sessão. A intensidade não aumenta.
- **Resposta dentro do esperado:** mantém o plano inalterado.

As alterações nunca ultrapassam os minutos disponíveis cadastrados para o dia. Sessões concluídas, puladas ou já modificadas por um feedback anterior não são recalculadas.

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
