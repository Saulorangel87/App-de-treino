# PROMPT — APP INTELIGENTE DE PLANEJAMENTO DE TREINOS

## 1. VISÃO GERAL DO PROJETO

Quero desenvolver uma aplicação web de planejamento e acompanhamento de treinos personalizados, inicialmente voltada para três modalidades:

* 🚴 Ciclismo
* 🏃 Corrida
* 🏋️ Musculação

O objetivo principal do aplicativo é criar planos de treinamento individualizados com base em:

* perfil do usuário;
* objetivo;
* experiência;
* disponibilidade;
* histórico de treinamento;
* capacidade física atual;
* recuperação;
* equipamentos disponíveis;
* limitações informadas;
* evolução observada ao longo do tempo;
* evidências científicas relacionadas ao treinamento.

O sistema deverá utilizar Inteligência Artificial, porém a IA **não deverá simplesmente inventar treinos**.

A aplicação deverá possuir um **motor de prescrição baseado em regras, evidências científicas e dados do usuário**, enquanto a IA atuará principalmente na interpretação, explicação, personalização e adaptação do plano dentro dos limites definidos pelo sistema.

---

# 2. PRINCÍPIO FUNDAMENTAL

A aplicação deve seguir esta lógica:

```text
USUÁRIO
   ↓
QUESTIONÁRIO ADAPTATIVO
   ↓
PERFIL DO ATLETA
   ↓
OBJETIVO
   ↓
NÍVEL ATUAL
   ↓
DISPONIBILIDADE
   ↓
SEGURANÇA / LIMITAÇÕES
   ↓
AVALIAÇÃO INICIAL
   ↓
MOTOR DE PRESCRIÇÃO
   ↓
BASE DE EVIDÊNCIAS
   ↓
PLANO DE TREINO
   ↓
TREINO REALIZADO
   ↓
FEEDBACK
   ↓
ANÁLISE DE EVOLUÇÃO
   ↓
ADAPTAÇÃO DO PRÓXIMO TREINO
```

O sistema deve ser progressivo e adaptativo.

Não quero um aplicativo que faça apenas:

```text
questionário → treino
```

Quero:

```text
questionário → avaliação → treino → feedback → adaptação → evolução
```

---

# 3. QUESTIONÁRIO ADAPTATIVO

Não criar um formulário inicial com 50–80 perguntas obrigatórias.

Isso prejudica a experiência do usuário.

O sistema deverá utilizar **perguntas condicionais**.

O usuário responde inicialmente apenas às perguntas essenciais.

Dependendo das respostas, novas perguntas são apresentadas.

Exemplo:

```text
Você utiliza medidor de potência?

( ) Sim
( ) Não
```

Se responder:

```text
Não
```

não apresentar perguntas relacionadas a FTP ou potência.

Se responder:

```text
Sim
```

apresentar:

* FTP;
* data do último teste;
* protocolo utilizado;
* potência média;
* outros dados relevantes.

O questionário deve ser inteligente e dinâmico.

---

# 4. ETAPA 1 — PERFIL BÁSICO

Inicialmente coletar aproximadamente 8 informações:

* idade;
* sexo;
* altura;
* peso;
* modalidade;
* experiência;
* objetivo principal;
* nível atual de atividade.

Essas perguntas devem ser simples e rápidas.

Não solicitar informações que não tenham utilidade para a tomada de decisão.

---

# 5. ETAPA 2 — OBJETIVOS

O usuário deverá informar:

### Objetivo principal

Possibilidades:

* emagrecimento;
* ganho de massa muscular;
* ganho de força;
* aumento de resistência;
* melhora do condicionamento;
* melhora de performance;
* preparação para competição;
* manutenção da condição física;
* outro.

### Objetivo secundário

Permitir selecionar outro objetivo.

Exemplo:

```text
Objetivo principal:
Emagrecimento

Objetivo secundário:
Melhorar resistência
```

O sistema deverá considerar a prioridade entre os objetivos.

---

# 6. ETAPA 3 — DISPONIBILIDADE

Perguntar:

* quantos dias por semana pode treinar;
* quanto tempo possui por sessão;
* quais dias estão disponíveis;
* horário preferido;
* local de treinamento;
* equipamentos disponíveis.

Permitir disponibilidade individual por dia.

Exemplo:

```text
Segunda: 60 min
Terça: 0
Quarta: 60 min
Quinta: 45 min
Sexta: 0
Sábado: 120 min
Domingo: 90 min
```

O plano deverá respeitar obrigatoriamente a disponibilidade real do usuário.

Não criar planos teoricamente excelentes, mas impossíveis de executar.

---

# 7. ETAPA 4 — HISTÓRICO ESPORTIVO

Essa etapa deverá ser diferente para cada modalidade.

## CICLISMO

Perguntar somente informações relevantes ao perfil do ciclista:

* tempo praticando ciclismo;
* frequência semanal;
* distância semanal aproximada;
* duração média dos treinos;
* maior distância realizada;
* tipo de ciclismo;
* tipo de bicicleta;
* terreno predominante;
* participação em provas;
* objetivo de prova, caso exista.

### Dados opcionais

Perguntar se possui:

* GPS;
* relógio esportivo;
* monitor cardíaco;
* medidor de potência;
* smart trainer.

Somente quando possuir determinado equipamento, apresentar as perguntas correspondentes.

Exemplo:

```text
Você utiliza medidor de potência?
```

Sim:

```text
Qual seu FTP?
Quando foi realizado o teste?
Qual protocolo foi utilizado?
```

---

# 8. CORRIDA

Para corredores, adaptar as perguntas.

Coletar, quando aplicável:

* tempo praticando corrida;
* frequência semanal;
* distância semanal;
* distância média por treino;
* maior distância;
* pace médio;
* melhores marcas;
* 5 km;
* 10 km;
* meia maratona;
* maratona;
* objetivo de prova;
* terreno;
* esteira ou rua;
* frequência cardíaca;
* relógio esportivo.

Se o usuário estiver treinando para uma prova, perguntar:

* distância;
* data da prova;
* objetivo de tempo;
* experiência anterior na distância.

---

# 9. MUSCULAÇÃO

Para musculação, adaptar completamente o questionário.

Coletar:

* experiência;
* frequência semanal;
* objetivo;
* local de treinamento;
* equipamentos disponíveis;
* exercícios realizados;
* experiência com treinamento de força;
* duração das sessões.

Perguntar sobre cargas ou desempenho somente quando fizer sentido.

Não exigir 1RM de usuários iniciantes.

O sistema deverá reconhecer diferentes níveis de experiência.

---

# 10. SAÚDE E SEGURANÇA

Essa etapa é obrigatória.

Perguntar de forma clara e simples:

* possui alguma lesão atual?
* sente dor durante exercícios?
* possui alguma limitação de movimento?
* realizou cirurgia recentemente?
* existe algum exercício que não pode realizar?
* possui alguma condição que possa interferir na prática de exercícios?

Se houver respostas que indiquem possível risco, o sistema deverá:

1. sinalizar a situação;
2. evitar prescrição potencialmente inadequada;
3. recomendar avaliação de profissional habilitado quando necessário.

A IA não deve diagnosticar doenças ou lesões.

O sistema deve diferenciar:

```text
PERSONALIZAÇÃO DE TREINO
```

de

```text
DIAGNÓSTICO / PRESCRIÇÃO CLÍNICA
```

---

# 11. RECUPERAÇÃO

Coletar inicialmente:

* horas de sono;
* qualidade do sono;
* nível de estresse;
* percepção de fadiga.

Essas informações deverão influenciar a interpretação da carga de treinamento.

Exemplo:

Se o usuário normalmente treina 5 vezes por semana, mas informa:

```text
Sono: ruim
Estresse: muito alto
Fadiga: alta
```

o sistema deverá considerar redução ou adaptação da carga, quando apropriado.

---

# 12. BIOTIPO E COMPOSIÇÃO CORPORAL

Não utilizar classificações simplistas como:

* ectomorfo;
* mesomorfo;
* endomorfo;

como base principal para prescrição.

Priorizar dados mensuráveis e relevantes:

* idade;
* sexo;
* altura;
* peso;
* composição corporal, quando disponível;
* circunferência da cintura, quando relevante;
* histórico de peso;
* nível de treinamento;
* desempenho;
* recuperação;
* objetivo.

O sistema deve priorizar dados objetivos em vez de classificações corporais genéricas.

---

# 13. AVALIAÇÃO INICIAL

O questionário não deve tentar descobrir tudo.

Após o cadastro inicial, o aplicativo poderá propor uma **avaliação inicial**.

A avaliação deverá ser específica para a modalidade e para o nível do usuário.

O objetivo é obter dados reais sobre a capacidade atual.

Exemplo:

```text
PERFIL INICIAL
      ↓
AVALIAÇÃO
      ↓
DADOS REAIS
      ↓
AJUSTE DO PLANO
```

Para usuários sem histórico suficiente, o sistema poderá utilizar treinos iniciais de avaliação.

A avaliação nunca deverá ser excessivamente complexa para iniciantes.

---

# 14. MOTOR DE PRESCRIÇÃO

O aplicativo deverá possuir um motor responsável por transformar os dados do usuário em parâmetros de treinamento.

Exemplo conceitual:

```text
Perfil
+
Objetivo
+
Experiência
+
Disponibilidade
+
Capacidade atual
+
Recuperação
+
Histórico
+
Equipamentos
+
Restrições
+
Evidências científicas
=
Plano personalizado
```

O motor deverá trabalhar com regras explícitas.

Não depender exclusivamente de uma chamada para um LLM.

---

# 15. PAPEL DA INTELIGÊNCIA ARTIFICIAL

A IA deverá ser utilizada para:

* interpretar informações;
* explicar o plano;
* explicar por que determinado treino foi escolhido;
* responder dúvidas;
* adaptar a comunicação ao usuário;
* interpretar feedback;
* auxiliar na adaptação do planejamento;
* identificar padrões nos dados;
* transformar informações complexas em linguagem simples.

A IA NÃO deverá:

* inventar evidências científicas;
* inventar estudos;
* diagnosticar doenças;
* ignorar regras de segurança;
* prescrever algo contrário às restrições do motor;
* substituir avaliação profissional quando necessária.

---

# 16. BASE CIENTÍFICA

O aplicativo deverá possuir uma estrutura para armazenar referências científicas.

Exemplo conceitual:

```text
ScientificSource

id
title
authors
year
journal
doi
url
sport
goal
population
evidence_level
summary
```

As fontes deverão ser classificadas por:

* modalidade;
* objetivo;
* população;
* tipo de treinamento;
* nível de evidência.

O sistema deverá priorizar:

* revisões sistemáticas;
* meta-análises;
* consensos;
* posicionamentos de organizações reconhecidas;
* estudos relevantes e de boa qualidade.

Evitar utilizar conteúdo de blogs ou redes sociais como fundamento científico principal.

---

# 17. EXPLICABILIDADE

O usuário deverá conseguir entender por que recebeu determinado treino.

Exemplo:

```text
Por que este treino?

Você recebeu este treino porque:

• seu objetivo principal é resistência;
• seu nível atual permite este volume;
• você possui 60 minutos disponíveis;
• sua recuperação está adequada;
• o treino complementa o estímulo realizado anteriormente.

Base científica:
[Referências]
```

A intenção é criar confiança no sistema.

---

# 18. GERAÇÃO DO PLANO

O plano deverá considerar:

* frequência;
* volume;
* intensidade;
* duração;
* recuperação;
* progressão;
* distribuição dos estímulos;
* objetivo;
* nível;
* disponibilidade.

O sistema não deverá aumentar carga indiscriminadamente.

Deverá existir lógica de progressão e recuperação.

---

# 19. TREINO DIÁRIO

Cada treino deverá apresentar:

### Nome

Exemplo:

```text
Treino de resistência aeróbica
```

### Objetivo

```text
Desenvolvimento da capacidade aeróbica.
```

### Duração

```text
60 minutos
```

### Estrutura

```text
Aquecimento
↓
Parte principal
↓
Recuperação
↓
Desaquecimento
```

### Intensidade

Utilizar métricas adequadas à modalidade e aos dados disponíveis:

* percepção subjetiva de esforço;
* frequência cardíaca;
* pace;
* potência;
* carga;
* repetições;
* outras métricas relevantes.

---

# 20. FEEDBACK PÓS-TREINO

Após cada treino, o usuário deverá informar como foi.

Exemplo:

```text
Como foi o treino?

😄 Muito fácil
🙂 Fácil
😐 Moderado
😓 Difícil
🥵 Muito difícil
```

Também registrar, quando disponível:

* duração;
* distância;
* carga;
* potência;
* frequência cardíaca;
* pace;
* repetições;
* RPE;
* dor;
* fadiga.

---

# 21. SISTEMA ADAPTATIVO

O aplicativo deverá comparar:

```text
TREINO PLANEJADO
        ↓
TREINO REALIZADO
        ↓
FEEDBACK
        ↓
RECUPERAÇÃO
        ↓
HISTÓRICO
        ↓
PRÓXIMO TREINO
```

Exemplo:

O sistema planejou:

```text
60 minutos
RPE 6
```

O usuário informou:

```text
60 minutos
RPE 9
Muito cansativo
```

O sistema deverá considerar esse dado no planejamento seguinte.

Da mesma maneira:

```text
Treino planejado: RPE 6
Treino realizado: RPE 3
```

poderá indicar que o estímulo foi abaixo do esperado.

A adaptação deverá considerar contexto e histórico, não apenas uma sessão isolada.

---

# 22. PERFIL DO ATLETA

O aplicativo deverá manter um perfil dinâmico.

Exemplo:

```text
Perfil inicial
      ↓
Semana 1
      ↓
Novos dados
      ↓
Semana 2
      ↓
Novos dados
      ↓
Semana 3
      ↓
Atualização do perfil
```

O sistema deve aprender com os dados registrados.

---

# 23. DASHBOARD

Criar um dashboard simples e moderno mostrando:

* treino do dia;
* próximo treino;
* progresso;
* sessões realizadas;
* volume;
* evolução;
* consistência;
* metas;
* histórico;
* indicadores de recuperação;
* evolução de performance.

Evitar excesso de informações.

O usuário deve conseguir entender seu estado atual rapidamente.

---

# 24. EVOLUÇÃO

O aplicativo deverá mostrar evolução ao longo do tempo.

Dependendo da modalidade:

### Ciclismo

* distância;
* duração;
* potência;
* velocidade;
* frequência cardíaca;
* carga;
* volume semanal.

### Corrida

* distância;
* pace;
* tempo;
* frequência cardíaca;
* volume semanal;
* melhores marcas.

### Musculação

* carga;
* repetições;
* volume;
* exercícios;
* evolução de desempenho;
* frequência.

---

# 25. ARQUITETURA TÉCNICA

A aplicação deverá ser construída pensando em escalabilidade.

### Frontend

Preferência:

```text
React
JavaScript
Vite
```

### Backend

Preferência:

```text
Go
REST API
```

### Banco de dados

```text
PostgreSQL
```

### Autenticação

Implementar autenticação segura.

### IA

Utilizar API de modelo de linguagem.

A integração deverá ficar no backend.

Não expor chaves da API no frontend.

---

# 26. MODELO DE DADOS INICIAL

Estruturar entidades semelhantes a:

```text
users

athlete_profiles

sports

goals

availability

training_history

training_plans

workouts

workout_sessions

exercises

exercise_categories

measurements

recovery_data

injuries_or_limitations

scientific_sources

training_rules

feedback
```

As relações devem ser projetadas de forma que o sistema possa futuramente suportar outras modalidades.

---

# 27. QUESTIONÁRIO COMO SISTEMA DINÂMICO

O questionário não deve ser codificado como uma sequência fixa de dezenas de perguntas.

Criar uma estrutura capaz de definir:

```text
Pergunta
    ↓
Resposta
    ↓
Regra
    ↓
Próxima pergunta
```

Exemplo:

```text
Possui medidor de potência?

SIM
 ↓
Perguntar FTP

NÃO
 ↓
Pular FTP
```

Outro exemplo:

```text
Está treinando para uma prova?

SIM
 ↓
Distância?
Data?
Objetivo de tempo?
Experiência anterior?

NÃO
 ↓
Continuar questionário normal
```

Isso deverá permitir que o questionário cresça sem obrigar todos os usuários a responder tudo.

---

# 28. EXPERIÊNCIA DO USUÁRIO

O processo inicial deverá parecer uma entrevista.

Não apresentar:

```text
Formulário 1 de 80
```

Preferir:

```text
Vamos conhecer você e entender seu objetivo.
```

Mostrar progresso:

```text
● ● ● ○ ○
```

Apresentar poucas perguntas por tela.

Usar linguagem simples.

Quando uma pergunta técnica for necessária, explicar brevemente o significado.

---

# 29. PRIMEIRA VERSÃO DO PRODUTO — MVP

Não desenvolver todas as funcionalidades imediatamente.

O MVP deverá começar com:

### Modalidade

Ciclismo.

### Funcionalidades

1. Cadastro;
2. perfil;
3. questionário adaptativo;
4. objetivo;
5. disponibilidade;
6. histórico;
7. avaliação inicial;
8. geração do plano;
9. calendário de treinos;
10. treino diário;
11. registro do treino;
12. feedback;
13. adaptação básica;
14. histórico de evolução;
15. explicação do motivo do treino.

Somente depois expandir para:

```text
Ciclismo
      ↓
Corrida
      ↓
Musculação
```

---

# 30. PRINCÍPIO DE DESENVOLVIMENTO

Antes de escrever código, definir:

1. requisitos;
2. regras de negócio;
3. questionário;
4. modelo de dados;
5. motor de prescrição;
6. estrutura da base científica;
7. fluxo da IA;
8. arquitetura;
9. API;
10. frontend.

Não começar criando telas aleatoriamente.

Primeiro criar a especificação do sistema.

---

# 31. PRINCIPAL DIFERENCIAL DO PRODUTO

O diferencial não deverá ser simplesmente:

> "Treinos feitos por IA."

O posicionamento deverá ser mais próximo de:

> **Um sistema inteligente de treinamento personalizado que combina dados individuais, evidências científicas, avaliação de desempenho e adaptação contínua.**

A IA é uma parte do sistema.

O verdadeiro produto é o **motor inteligente de treinamento**.

---

# 32. OBJETIVO FINAL

Criar uma plataforma capaz de responder:

> "Dado quem é essa pessoa, qual é seu objetivo, qual sua capacidade atual, quanto tempo ela possui, como está sua recuperação, o que ela realizou anteriormente e quais evidências científicas se aplicam ao caso, qual é o estímulo de treinamento mais adequado neste momento?"

E não simplesmente:

> "Qual treino a IA consegue inventar para essa pessoa?"

Essa diferença deve orientar toda a arquitetura do projeto.


1. 🧠 Regras do sistema
Definir como o app toma decisões.

2. 📝 Questionário adaptativo
Começar pelo ciclismo e definir perguntas, respostas e ramificações.

3. 📊 Perfil do atleta
Definir quais dados realmente precisamos guardar.

4. 📚 Base científica
Mapear quais evidências sustentam cada tipo de treinamento.

5. ⚙️ Motor de prescrição
Transformar perfil + objetivo + evidências em parâmetros de treino.

6. 🤖 IA
Definir exatamente onde ela entra e quais limites terá.

7. 🗄️ Banco + API
Modelar PostgreSQL e backend em Go.

8. 💻 Frontend
Construir o React/TypeScript em cima de tudo que já foi definido.

9. 🔄 Sistema adaptativo
Fazer o aplicativo aprender com os treinos realizados e feedback do usuário.

---

# Registro de execução do MVP — 2 de setembro de 2026

O MVP de ciclismo previsto neste planejamento foi implementado e está em produção real. O fluxo validado inclui cadastro, perfil, objetivos, disponibilidade, limitações, avaliação submáxima, geração e ativação do plano, execução da sessão, feedback, adaptação, histórico de atividades, evolução e logout.

O motor atual é o `rules-v1`: determinístico, explicável e baseado em regras e referências científicas. Ele não depende de um modelo de linguagem para inventar treinos. A integração de IA no backend ainda é uma próxima fase e deverá respeitar as regras de segurança, as evidências e os limites do motor.

A infraestrutura de produção usa PostgreSQL próprio na VPS Oracle, frontend e API expostos por Cloudflare Tunnel dedicado e PostgreSQL em rede Docker interna. Backups diários estão ativos. A documentação operacional detalhada está em `docs/project-status.md` e `infrastructure/cadencia/README.md`.

Pendências principais:

- definir cópia externa dos backups e monitoramento de falhas;
- concluir hardening das portas dos demais aplicativos hospedados na VPS, após mapear domínios, túneis, proxies e regras da Oracle Cloud;
- validar visualmente e publicar o ajuste visual de privacidade e altura da tela inicial desktop;
- validar a IA explicativa com o Worker remoto; o Ollama foi instalado e testado, mas permanece parado porque a inferência local consumiu capacidade excessiva da VPS;
- ampliar a coleta progressiva de dados e a variedade de sessões específicas de ciclismo;
- avaliar integrações externas, como Strava, somente após definir consentimento, custos e segurança.

## Atualização operacional — 3 de setembro de 2026

O commit `bef4e60` foi implantado na VPS. O serviço `cadencia-ollama-1` está isolado na rede Docker interna, limitado a 4 GiB de memória, 1 CPU e uma chamada simultânea, sem porta publicada. O modelo `qwen3:4b-instruct` foi baixado e uma inferência simples foi validada, mas uma chamada completa levou cerca de 72 segundos, usou praticamente 100% do limite de CPU e deixou a VPS com pouca memória livre. O serviço foi parado e a API passou a usar temporariamente o Worker remoto com `AI_ENABLED=true` e `AI_PROVIDER=worker`; o padrão seguro continua sendo `false`. A próxima etapa é validar a explicação autenticada pelo Worker e monitorar seus limites.
