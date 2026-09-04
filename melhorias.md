Quero evoluir o Cadência para que ele se torne um treinador digital de ciclismo mais inteligente, seguro, explicável, adaptativo e baseado em evidências científicas.

Corrida e musculação ainda não fazem parte do produto. Não implemente essas modalidades agora.

O objetivo não é apenas adicionar mais tipos de treino. O sistema deve saber quando prescrever, manter, substituir, reduzir ou interromper cada estímulo com base no contexto atual do atleta e na resposta real aos treinos.

Antes de alterar qualquer arquivo:

1. Leia `planejamento.md`, `README.md` e toda a documentação existente.
2. Verifique a arquitetura atual do frontend, backend, banco de dados e motor de regras.
3. Localize e entenda o funcionamento atual do `rules-v1`.
4. Analise os fluxos de onboarding, perfil, plano, avaliação, recuperação, atividades, evolução e feedback.
5. Confira `git status` e os commits recentes.
6. Identifique o que já está implementado, o que está incompleto e o que é apenas planejamento.
7. Apresente primeiro um diagnóstico técnico e um plano de implementação priorizado.
8. Não faça commit, deploy, publicação ou alteração de infraestrutura sem minha autorização explícita.

## 1. Separar experiência de prontidão atual

O perfil não deve depender apenas dos níveis iniciante, intermediário e avançado.

Mantenha o nível de experiência, mas crie também uma classificação de prontidão atual, considerando:

- tempo parado ou destreinamento;
- frequência recente de treino;
- volume dos últimos 7, 28 e 42 dias;
- sessões concluídas;
- sessões canceladas;
- RPE médio;
- fadiga;
- sono;
- estresse;
- dor;
- recuperação;
- consistência;
- histórico recente de desempenho;
- distância, duração e elevação;
- frequência cardíaca;
- potência, quando disponível;
- aderência ao plano.

Um atleta avançado que ficou meses sem pedalar deve ser tratado como experiente, mas com prontidão reduzida. A experiência pode definir a complexidade máxima do treinamento, mas não deve justificar sozinha uma carga elevada.

Crie estados como:

- iniciante;
- intermediário;
- avançado;
- retornando após pausa;
- baixa consistência;
- baixa prontidão;
- pronto para progressão;
- preparação específica para evento.

## 2. Evoluir o motor de regras

Analise o `rules-v1` e crie uma evolução versionada, preferencialmente sem apagar o comportamento anterior até que a nova versão seja testada.

O novo motor deve considerar:

- objetivo principal;
- objetivo secundário;
- experiência;
- prontidão atual;
- período de destreinamento;
- volume recente;
- frequência semanal;
- disponibilidade;
- recuperação;
- dor e limitações;
- equipamento;
- terreno;
- tipo de bicicleta;
- métricas disponíveis;
- evento-alvo;
- tempo até o evento;
- histórico de aderência;
- resposta aos treinos anteriores;
- sessões difíceis recentes;
- necessidade de recuperação.

As regras devem ser:

- determinísticas;
- testáveis;
- versionadas;
- auditáveis;
- explicáveis;
- conservadoras quando houver incerteza.

Não use IA generativa para inventar treinos, cargas, volumes, intensidades ou progressões. A IA pode explicar a decisão em linguagem natural, mas a prescrição deve ser produzida pelo motor de regras.

## 3. Criar uma base de evidências científicas

Não espalhe referências científicas diretamente pelo código.

Crie uma estrutura versionada de evidências com:

- fonte;
- título;
- autores;
- ano;
- URL ou DOI;
- tipo de evidência;
- população estudada;
- objetivo;
- estímulo analisado;
- benefícios esperados;
- limitações;
- riscos;
- contraindicações;
- nível de confiança;
- data da última revisão;
- regras relacionadas.

Priorize:

1. diretrizes oficiais;
2. consensos de organizações esportivas;
3. revisões sistemáticas;
4. meta-análises;
5. ensaios clínicos;
6. estudos observacionais;
7. opinião de especialistas apenas quando necessário.

Não invente referências. Não transforme uma conclusão específica de um estudo em regra universal. Cada regra deve indicar em quais situações a evidência é aplicável e quais são suas limitações.

A explicação científica deve deixar claro que evidência orienta a decisão, mas não substitui avaliação individual ou orientação profissional.

## 4. Ampliar a biblioteca de treinos de ciclismo

Além dos treinos atuais, crie templates estruturados para:

- recuperação ativa;
- descanso completo;
- endurance contínuo;
- pedal longo;
- giro de base;
- trabalho de cadência;
- tempo;
- sweet spot;
- limiar;
- VO2max;
- intervalos curtos;
- sprints;
- estímulo neuromuscular;
- subidas;
- resistência muscular;
- técnica;
- teste submáximo;
- retorno após pausa;
- redução de carga;
- taper pré-prova;
- recuperação pós-prova.

Cada tipo de treino deve possuir:

- objetivo fisiológico;
- objetivo prático;
- indicação;
- contraindicação;
- nível recomendado;
- pré-requisitos;
- aquecimento;
- parte principal;
- intervalos de recuperação;
- desaquecimento;
- duração;
- intensidade por RPE;
- intensidade por frequência cardíaca, quando possível;
- intensidade por potência, quando possível;
- cadência recomendada, quando aplicável;
- critério de interrupção;
- critérios de progressão;
- critérios de regressão;
- evidências relacionadas.

O motor não deve selecionar um treino apenas pelo nome do objetivo. Deve verificar se o atleta está preparado para aquele estímulo.

## 5. Diferenciar iniciantes, intermediários e avançados

Crie comportamentos específicos para cada nível.

### Iniciantes

Priorizar:

- adaptação gradual;
- consistência;
- sessões simples;
- baixa complexidade;
- recuperação suficiente;
- progressão conservadora;
- uso de RPE;
- educação sobre esforço;
- limites rígidos para intensidade.

### Intermediários

Priorizar:

- progressão de volume e intensidade;
- alternância entre estímulos;
- maior variedade de sessões;
- controle da distribuição semanal;
- evolução da resistência;
- introdução gradual de treinos de qualidade;
- análise da aderência.

### Avançados

Priorizar:

- especificidade do objetivo;
- potência;
- frequência cardíaca;
- cadência;
- zonas de treinamento;
- distribuição de intensidade;
- sessões de limiar e VO2max;
- subidas;
- periodização;
- taper;
- preparação para evento;
- comparação entre carga interna e externa.

O nível avançado nunca deve substituir a análise da prontidão atual.

## 6. Criar uma adaptação realmente fechada

Após cada sessão concluída, compare o planejado com o realizado:

- duração;
- RPE planejado e realizado;
- distância;
- elevação;
- frequência cardíaca;
- potência;
- cadência;
- fadiga;
- dor;
- sono;
- estresse;
- recuperação;
- conclusão total ou parcial;
- motivo de não conclusão.

Use esses dados para decidir automaticamente se a próxima sessão deve:

- progredir;
- permanecer igual;
- reduzir volume;
- reduzir intensidade;
- trocar o tipo de estímulo;
- incluir recuperação;
- ser remanejada;
- ser adiada;
- ser cancelada por segurança.

Registre o motivo de cada alteração e exiba-o ao usuário.

O sistema deve evitar manter um plano fixo quando os dados recentes mostrarem que o atleta está respondendo de forma diferente do esperado.

## 7. Controlar carga e progressão

Use múltiplos sinais para controlar carga:

- duração;
- RPE;
- carga da sessão;
- frequência;
- distância;
- elevação;
- potência;
- frequência cardíaca;
- fadiga;
- sono;
- estresse;
- dor;
- aderência;
- número de sessões intensas;
- densidade dos treinos.

Não utilize uma fórmula isolada como verdade absoluta. Em especial, não aplique ACWR de forma automática sem considerar suas limitações científicas.

A progressão deve exigir evidências de tolerância. O sistema deve reduzir ou manter a carga quando houver:

- recuperação ruim;
- fadiga persistente;
- piora de desempenho;
- dor;
- baixa aderência;
- repetição de RPE acima do planejado;
- sono ruim combinado com estresse ou fadiga;
- queda importante de volume realizado.

Evite aumentar volume e intensidade simultaneamente sem uma justificativa clara.

## 8. Melhorar a periodização

Implemente fases coerentes com o objetivo:

- retorno ou readaptação;
- reconstrução da base;
- desenvolvimento geral;
- desenvolvimento específico;
- construção;
- pico;
- taper;
- recuperação;
- transição.

A periodização deve considerar:

- nível atual;
- experiência;
- tempo até o evento;
- disponibilidade;
- histórico recente;
- tolerância aos treinos;
- objetivo da prova;
- necessidade de recuperação.

Uma semana de recuperação não deve apenas reduzir minutos. O motor também deve avaliar:

- intensidade;
- quantidade de estímulos difíceis;
- densidade das sessões;
- carga interna;
- necessidade de dias de descanso;
- dor e fadiga acumuladas.

## 9. Criar seleção inteligente de estímulos

O motor deve escolher o estímulo com base na necessidade do atleta, e não apenas alternar templates.

Exemplos:

- pouca base aeróbica: priorizar endurance;
- baixa consistência: reduzir complexidade e aumentar aderência;
- bom volume, mas baixa tolerância à intensidade: manter base e introduzir qualidade gradualmente;
- prova longa: priorizar duração, resistência e especificidade;
- prova curta e intensa: inserir intensidade somente após base adequada;
- retorno após pausa: usar bloco de readaptação;
- fadiga elevada: trocar intensidade por recuperação;
- sessões repetidamente difíceis: reduzir carga ou alterar estímulo;
- boa tolerância com objetivo específico: aumentar progressivamente a especificidade.

O sistema deve evitar:

- duas sessões intensas próximas;
- excesso de sessões de qualidade;
- aumento de carga após baixa aderência;
- treino avançado com dados insuficientes;
- prescrição de intensidade elevada para atleta destreinado;
- usar apenas o objetivo declarado para determinar carga;
- ignorar dor, fadiga ou recuperação.

## 10. Melhorar segurança

Amplie o questionário para registrar:

- tipo de dor;
- localização;
- intensidade;
- movimento que agrava;
- lesão atual ou anterior;
- restrição médica;
- data de início;
- orientação profissional;
- sintomas durante ou após o treino.

Crie regras claras para reduzir ou impedir treinos intensos quando houver:

- dor;
- tontura;
- falta de ar incomum;
- mal-estar;
- fadiga extrema;
- piora importante da recuperação;
- combinação de sono ruim, estresse alto e fadiga alta.

O Cadência não deve diagnosticar doenças. Porém, diante de sinais de alerta, deve:

- impedir ou desaconselhar claramente o início de sessões intensas;
- recomendar avaliação profissional quando apropriado;
- registrar que o treino foi bloqueado por segurança;
- evitar que o usuário contorne o bloqueio sem confirmação explícita.

## 11. Corrigir inconsistências de dados

Implemente validações para impedir registros impossíveis ou incoerentes, como:

- duração igual a zero com distância registrada;
- distância incompatível com duração;
- RPE ausente;
- sessão concluída sem dados mínimos;
- evolução mostrando volume negativo;
- métricas incompatíveis com o estado da sessão;
- frequência cardíaca impossível;
- valores duplicados ou parcialmente salvos.

Quando houver inconsistência:

- não use automaticamente o registro para progressão;
- marque-o como incompleto ou inconsistente;
- explique o problema ao usuário;
- permita correção quando apropriado;
- registre o impacto no motor;
- preserve o histórico original para auditoria.

## 12. Melhorar o feedback pós-treino

Além de RPE e fadiga, capture:

- treino fácil, adequado ou difícil demais;
- motivo de não conclusão;
- dor ou desconforto;
- percepção de recuperação;
- confiança para repetir o treino;
- satisfação;
- observações sobre terreno;
- equipamento utilizado;
- condições externas;
- diferença entre o esforço planejado e o percebido.

Use esse feedback diretamente na decisão da próxima sessão.

O feedback geral do produto deve continuar separado do feedback fisiológico do treino, mas ambos podem ser analisados para melhorar o produto e o motor.

## 13. Tornar todas as decisões auditáveis

Para cada sessão gerada, registrar:

- regras avaliadas;
- regras aplicadas;
- regras rejeitadas;
- dados utilizados;
- dados ausentes;
- limitações encontradas;
- evidências relacionadas;
- justificativa final;
- nível de confiança;
- condições que fariam o treino mudar.

Na interface, mostrar de forma clara:

- por que aquele treino foi escolhido;
- quais dados tiveram maior peso;
- quais restrições foram aplicadas;
- quais alternativas foram descartadas;
- quais informações estavam ausentes;
- o que faria o plano mudar.

A explicação não deve ser genérica nem esconder decisões importantes atrás de IA.

## 14. Manter o escopo exclusivamente no ciclismo

Corrida e musculação ainda não fazem parte do produto e não devem ser implementadas nesta etapa.

Não criar agora:

- telas de corrida;
- telas de musculação;
- exercícios de musculação;
- regras de impacto;
- séries, repetições ou cargas de musculação;
- protocolos específicos de corrida;
- opções de modalidade não funcionais;
- banco de dados incompleto para outras modalidades;
- fluxos de onboarding para corrida ou musculação.

Toda a implementação deve se concentrar em melhorar profundamente:

- onboarding do ciclista;
- classificação de prontidão;
- motor de regras do ciclismo;
- biblioteca de treinos de ciclismo;
- periodização;
- controle de carga;
- recuperação;
- adaptação ao desempenho;
- segurança;
- feedback;
- evolução;
- explicação das decisões;
- base científica.

A arquitetura pode manter pontos simples de extensão futura, mas não crie abstrações genéricas ou complexidade adicional apenas por antecipação.

Corrida e musculação serão planejadas em uma etapa futura, depois que o motor de ciclismo estiver validado com usuários e dados reais.

## 15. Testes obrigatórios

Crie testes automatizados para:

- iniciante;
- intermediário;
- avançado;
- retorno após pausa;
- baixa recuperação;
- dor;
- baixa aderência;
- evento próximo;
- ausência de potência;
- ausência de frequência cardíaca;
- dados inconsistentes;
- sessão parcialmente concluída;
- aumento de carga;
- manutenção de carga;
- redução de carga;
- semana de recuperação;
- duas sessões intensas próximas;
- excesso de fadiga;
- progressão após boa resposta;
- regressão após resposta ruim.

Crie também testes de integração para verificar:

- geração do plano;
- registro de sessão;
- feedback;
- atualização da evolução;
- adaptação da próxima sessão;
- explicação das regras;
- preservação dos dados históricos.

## 16. Critérios de aceitação

A implementação deve:

- preservar o funcionamento atual do app;
- manter o motor explicável;
- evitar prescrições baseadas apenas no nível de experiência;
- considerar prontidão atual e destreinamento;
- adaptar o plano com base no realizado;
- bloquear progressões quando houver sinais de risco;
- impedir dados inconsistentes de aumentarem carga;
- possuir evidências relacionadas às regras;
- não inventar citações científicas;
- possuir testes automatizados;
- registrar todas as decisões importantes;
- manter o PostgreSQL privado;
- não implementar corrida ou musculação nesta etapa;
- não alterar produção sem autorização;
- não fazer commit sem autorização.

Ao final, entregue:

1. diagnóstico do estado atual;
2. plano de implementação;
3. arquitetura proposta;
4. alterações realizadas;
5. arquivos modificados;
6. regras novas, alteradas e removidas;
7. evidências científicas utilizadas;
8. testes executados;
9. cenários ainda não cobertos;
10. limitações científicas e técnicas;
11. próximos passos recomendados.

Não faça commit, deploy ou publicação. Aguarde minha autorização explícita.