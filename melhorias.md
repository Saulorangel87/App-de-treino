Quero evoluir o motor de treinos do Cadência para torná-lo mais inteligente, seguro e adaptativo, especialmente para ciclismo nos níveis iniciante, intermediário e avançado.

Antes de alterar qualquer arquivo:

1. Leia `planejamento.md`, `README.md` e a documentação existente do projeto.
2. Verifique o estado atual do código, banco, regras de prescrição e adaptação.
3. Confira `git status` e os commits recentes.
4. Identifique quais funcionalidades já existem e quais são apenas planejadas.
5. Apresente um diagnóstico e um plano técnico priorizado antes de implementar.
6. Não faça commit, deploy ou publicação sem minha autorização explícita.

Objetivos principais:

### 1. Separar experiência de prontidão atual

O perfil não deve usar apenas “iniciante”, “intermediário” ou “avançado”.

Crie também uma classificação de prontidão atual, considerando:

- tempo parado ou destreinamento;
- frequência recente de treino;
- sessões concluídas;
- volume recente;
- RPE médio;
- fadiga;
- dor;
- recuperação;
- consistência;
- histórico recente de desempenho.

Um atleta avançado que ficou meses parado deve começar com um bloco de retorno progressivo, mesmo mantendo experiência avançada.

### 2. Criar uma adaptação realmente fechada

Após cada sessão concluída, compare o planejado com o realizado:

- duração;
- RPE planejado e realizado;
- distância;
- elevação;
- frequência cardíaca;
- potência, quando disponível;
- cadência;
- fadiga;
- dor;
- recuperação;
- conclusão parcial ou total.

Use esses dados para decidir automaticamente se a próxima sessão deve:

- progredir;
- permanecer igual;
- reduzir volume;
- reduzir intensidade;
- trocar o tipo de estímulo;
- incluir recuperação;
- ser remanejada;
- ser adiada.

Registre também o motivo de cada alteração.

### 3. Corrigir inconsistências de dados

Implemente validações para evitar registros impossíveis ou incoerentes, como:

- duração igual a zero com distância registrada;
- distância incompatível com duração;
- RPE ausente;
- sessão concluída sem dados mínimos;
- evolução mostrando volume negativo;
- métricas que não correspondem ao estado da sessão.

Quando houver inconsistência:

- não use o registro automaticamente para progressão;
- marque-o como incompleto ou inconsistente;
- explique o problema ao usuário;
- permita correção quando apropriado;
- registre o impacto no motor de adaptação.

### 4. Melhorar segurança

Amplie o questionário de segurança para permitir registrar:

- tipo de dor;
- localização;
- intensidade;
- movimento que agrava;
- lesão atual ou anterior;
- restrição médica;
- data de início;
- orientação profissional;
- sintomas durante ou após o treino.

Crie regras claras para impedir ou reduzir sessões intensas quando houver:

- dor;
- tontura;
- falta de ar incomum;
- mal-estar;
- fadiga extrema;
- combinação de sono ruim, estresse alto e fadiga alta.

O Cadência não deve diagnosticar doenças, mas deve aplicar bloqueios e recomendações conservadoras quando houver sinais de risco.

### 5. Diferenciar melhor os níveis

Crie comportamentos distintos para:

- iniciante;
- intermediário;
- avançado;
- retorno após pausa;
- baixa consistência;
- preparação para evento.

O nível deve alterar:

- volume inicial;
- frequência semanal;
- complexidade dos treinos;
- proporção entre intensidade e recuperação;
- velocidade de progressão;
- tolerância a sessões de qualidade;
- necessidade de avaliação inicial.

Não permita que o nível “avançado” sozinho justifique carga elevada quando os dados recentes indicarem baixa prontidão.

### 6. Melhorar o ciclismo avançado

Quando o usuário possuir dados adequados, permita prescrição baseada em:

- zonas de potência;
- zonas de frequência cardíaca;
- RPE;
- cadência;
- terreno;
- tipo de bicicleta;
- duração da prova;
- distância da prova;
- ganho de elevação;
- fase de preparação;
- taper;
- recuperação pós-evento.

Mantenha RPE como alternativa quando não houver sensores.

### 7. Tornar a explicação mais inteligente

Para cada sessão, mostre:

- quais dados foram usados;
- quais regras tiveram maior peso;
- quais restrições foram aplicadas;
- quais alternativas foram descartadas;
- por que aquele treino foi escolhido;
- qual informação está faltando;
- o que faria o plano mudar;
- nível de confiança da recomendação.

A explicação deve ser determinística, auditável e compreensível. Não esconda decisões importantes atrás de uma resposta genérica de IA.

### 8. Melhorar o feedback pós-treino

Além de RPE e fadiga, capture:

- treino fácil, adequado ou difícil demais;
- motivo de não conclusão;
- dor ou desconforto;
- percepção de recuperação;
- confiança para repetir o treino;
- satisfação;
- observações sobre terreno, equipamento ou condições externas.

Use esse feedback diretamente na próxima decisão do motor.

### 9. Preparar arquitetura para outras modalidades

Não implemente corrida e musculação simplesmente reutilizando as mesmas regras do ciclismo.

Crie uma camada comum para:

- perfil;
- objetivo;
- segurança;
- recuperação;
- aderência;
- feedback;
- progressão.

Depois mantenha motores específicos:

- ciclismo: potência, FC, cadência, terreno e duração;
- corrida: impacto, ritmo, superfície, caminhada/corrida e risco musculotendíneo;
- musculação: exercícios, séries, repetições, carga, RIR/RPE, grupos musculares, técnica e equipamento.

Nesta etapa, priorize a melhoria do ciclismo sem quebrar o suporte futuro às outras modalidades.

### Critérios de aceitação

A implementação deve:

- preservar o funcionamento atual do app;
- manter o motor explicável;
- incluir testes unitários para as novas regras;
- testar cenários de iniciante, intermediário, avançado e retorno após pausa;
- testar dados inconsistentes;
- testar dor e recuperação ruim;
- testar sessão parcialmente concluída;
- testar aumento, manutenção e redução de carga;
- mostrar claramente os motivos das adaptações;
- não usar dados inconsistentes para aumentar carga;
- manter o PostgreSQL privado;
- não alterar infraestrutura ou produção sem autorização.

Ao final, entregue:

1. diagnóstico do estado atual;
2. plano de implementação;
3. alterações realizadas;
4. arquivos modificados;
5. testes executados;
6. limitações ainda existentes;
7. próximos passos recomendados.

Não faça commit nem deploy.