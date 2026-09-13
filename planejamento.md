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

- observar os primeiros relatos reais pela aba `/feedback`, incluindo a entrega e utilidade do resumo semanal do Resend;
- definir cópia externa dos backups e monitoramento de falhas;
- concluir hardening das portas dos demais aplicativos hospedados na VPS, após mapear domínios, túneis, proxies e regras da Oracle Cloud;
- validar visualmente e publicar o ajuste visual de privacidade e altura da tela inicial desktop;
- monitorar a IA explicativa com o Worker remoto, incluindo latência, limites e fallback; o Ollama foi instalado e testado, mas permanece parado porque a inferência local consumiu capacidade excessiva da VPS;
- ampliar a coleta progressiva de dados e a variedade de sessões específicas de ciclismo;
- avaliar integrações externas, como Strava, somente após definir consentimento, custos e segurança.

## Atualização operacional — 3 de setembro de 2026

O commit `57c241a` foi implantado na VPS. O serviço `cadencia-ollama-1` está isolado na rede Docker interna, limitado a 4 GiB de memória, 1 CPU e uma chamada simultânea, sem porta publicada. O modelo `qwen3:4b-instruct` foi baixado e uma inferência simples foi validada, mas uma chamada completa levou cerca de 72 segundos, usou praticamente 100% do limite de CPU e deixou a VPS com pouca memória livre. O serviço foi parado e a API passou a usar temporariamente o Worker remoto com `AI_ENABLED=true` e `AI_PROVIDER=worker`; o padrão seguro continua sendo `false`. A próxima etapa é validar a explicação autenticada pelo Worker e monitorar seus limites.

O Worker foi atualizado para a versão `35f24685` com `max_completion_tokens: 512` e `reasoning_effort: 'low'`. A rota rejeita respostas cujo `finish_reason` não seja `stop`, encaminhando a API ao fallback determinístico. Os testes autenticados de “Giro de base” e “Subidas controladas” retornaram texto completo; a próxima etapa é monitorar latência e limites em uso normal.

## Atualização de produto — feedback dos primeiros ciclistas

Foi incluída a rota `/feedback` para que atletas autenticados registrem a experiência com o app, um problema ou uma sugestão. O relato contém uma nota de 1 a 5 e uma mensagem validada, fica vinculado à conta na tabela `user_feedback` e não interfere no motor de treinos. As migrações `000013_user_feedback` e `000014_feedback_digest` foram aplicadas na produção.

O plano de recebimento desta primeira coleta é centralizar os relatos no PostgreSQL, sem e-mail a cada envio. O comando separado `cadencia-feedback-digest` consolida até 50 relatos pendentes em um resumo semanal enviado pelo Resend ao endereço administrativo configurado em `FEEDBACK_DIGEST_TO`. O timer systemd está ativo e executará o job às segundas-feiras, às 08:00 no horário de São Paulo, sem manter processo adicional residente. Depois dos primeiros testes divulgados nos grupos de ciclismo, será avaliada a necessidade de uma tela administrativa protegida.

## Atualização de identidade visual — favicon

O favicon e os ícones instaláveis do PWA foram redesenhados com um símbolo próprio da Cadência: um arco de cadência em verde-lima, combinado a um núcleo de movimento em verde-claro sobre o fundo verde da marca. O mesmo desenho foi exportado para SVG e PNG (192, 512, maskable e Apple Touch Icon), evitando a aparência genérica do ícone anterior. A troca foi publicada no deploy oficial do frontend (`33de28a`).

## Atualização operacional — deploy oficial do ajuste mobile — 3 de setembro de 2026

O commit `33de28a` foi publicado na VPS Oracle por fast-forward após o commit já estar sincronizado no GitHub. Foi criado e verificado o backup preventivo `cadencia-20260904T023630Z.dump` (UTC). Somente a imagem e o container `cadencia-frontend-1` foram reconstruídos/recriados; a API, o PostgreSQL, o túnel e os demais aplicativos da VPS não foram alterados. As rotas públicas `cadencia.devsaulo.com.br` e `/evolucao` retornaram HTTP 200 e o frontend permaneceu saudável.

O ajuste resolve a sobreposição dos intervalos semanais e da barra de rolagem nos gráficos da Evolução em telas pequenas. Cada período mantém largura mínima legível e o gráfico usa rolagem interna horizontal quando necessário, sem criar rolagem horizontal indevida na página.

Durante o processo, uma cópia privada foi publicada por engano no ambiente Sites. Ela foi excluída manualmente pelo proprietário. A produção oficial continua sendo exclusivamente a VPS Oracle com o Cloudflare Tunnel dedicado; o Sites não deve ser usado como destino de deploy deste projeto.

## Atualização operacional — publicação do catálogo e correção do check-in — 4 de setembro de 2026

O commit `5fbc668` foi publicado na VPS Oracle por fast-forward. Foi criado e verificado o backup preventivo `cadencia-20260905T003553Z.dump` (UTC), a migração `000015` foi aplicada pelo perfil `maintenance` e as imagens da API e do frontend foram reconstruídas. Os containers de API e frontend foram recriados; PostgreSQL e túnel permaneceram ativos e saudáveis. A API interna respondeu `{"status":"ready"}` e os domínios públicos retornaram HTTP 200.

Na sequência, o commit `c768ef7` atualizou somente o frontend para publicar a versão `0.7.0` e a nota sobre o catálogo de ciclismo baseado em evidências. A tela de novidades foi confirmada em uma conta autenticada. O check-in de recuperação foi validado em produção após a correção, com redução da próxima sessão para 20 minutos e RPE 4 quando os sinais combinados indicaram necessidade de recuperação.

## Próxima etapa registrada

A próxima etapa é a observação acompanhada: aguardar relatos reais pela aba `/feedback` e observar o primeiro resumo semanal enviado pelo Resend. O fluxo de feedback já foi testado, e a latência, os limites e o fallback do Worker de IA já foram medidos; não há necessidade de repetir esses testes agora. A ampliação futura do catálogo de protocolos específicos de ciclismo deverá continuar com evidências e instruções operacionais mais amigáveis para cada sessão. Nenhuma etapa de IA generativa ou integração externa deve substituir o motor determinístico `rules-v1` antes dessa revisão.

## Direção atual da próxima fase — 4 de setembro de 2026

Após a publicação do catálogo inicial e do piloto de intervalos moderados de estrada, o roadmap vigente da próxima fase está em [`melhorias.md`](melhorias.md). A primeira fatia deve tratar prontidão e qualidade dos dados; depois vêm a evolução versionada das regras, adaptação em ciclo fechado, progressão/carga, segurança, feedback e auditabilidade. O `rules-v1` deve ser preservado durante a validação, e o escopo permanece exclusivo de ciclismo.

Toda atualização com funcionalidade visível deve atualizar `frontend/lib/release.ts`, incrementando `APP_VERSION` e registrando a novidade em `UPDATE_NOTES`, para que ela apareça na tela de novidades após a atualização. A produção oficial está no commit `53cbadc`, na versão `0.12.0`, com a release [v0.12.0](https://github.com/Saulorangel87/App-de-treino/releases/tag/v0.12.0) publicada no GitHub. O deploy de 11 de setembro foi validado após backup preventivo, build, manutenção e checagem dos domínios oficiais; a correção do acesso ao perfil no mobile foi somente frontend e não exigiu nova nota de versão. Não fazer commit, deploy ou mudança de infraestrutura sem autorização explícita.

### Continuidade — pesquisa do próximo protocolo — 11 de setembro de 2026

A revisão de estudos recentes não encontrou base suficiente para criar agora um protocolo próprio de gravel/XCM: os trabalhos localizados descrevem demandas de prova ou hidratação, mas não validam uma prescrição de treino transferível para o público geral. O próximo candidato documentado é o taper pré-prova, orientado por uma data de evento válida e tratado como alteração de volume do plano, não como simples nome de sessão. A literatura de endurance apoia redução de volume em janela curta com manutenção cuidadosa da intensidade e frequência, mas não autoriza uma dose universal nem sobrecarga automática antes da prova.

Antes de editar o motor, a próxima tarefa é especificar elegibilidade, janela temporal, distribuição semanal, limites de volume, proteção por dor/recuperação e testes de evento distante ou próximo. `rules-v1` continua sendo o único motor prescritivo; a futura funcionalidade visível deverá atualizar `APP_VERSION` e `UPDATE_NOTES`. Nenhum código do taper, commit, deploy, migração ou mudança de infraestrutura foi realizado nesta etapa.

### Continuidade — 5 de setembro de 2026

A partir do checkout limpo `70103b0`, foi implementada localmente a primeira leitura de prontidão (`readiness-v1`), independente de experiência, em modo observacional no snapshot de novos planos. Ela registra motivos, lacunas e cobertura dos dados de 28 dias; não altera a prescrição `rules-v1`, a avaliação ou a adaptação do check-in. Não autoriza progressão nem infere baixa consistência a partir de poucos registros. Os testes Go sem cache e `go vet` passaram; depois que o proprietário ligou o Docker, as consultas PostgreSQL passaram em quatro cenários fictícios, sem ler ou alterar dados reais. Falta a conferência ponta a ponta do novo campo via API/navegador. Detalhes, limitações e próximo passo estão em `docs/project-status.md` e `docs/training-adaptation-rules.md`.

Próxima fatia: consolidar janelas 7/28/42 dias, qualidade temporal dos dados e aderência/tolerância antes de usar prontidão em decisões de carga. Feedback real e confirmação do Resend seguem em paralelo, sem bloquear melhorias. Sem mudança visual nesta entrega, a versão do produto permanece `0.7.0`; atualizar a tela de novidades na próxima funcionalidade visível. Nenhum commit, deploy, migração ou mudança de infraestrutura foi realizado nesta etapa. Este planejamento raiz e `melhorias.md` foram preservados.

### Continuidade — histórico 7/28/42 dias e aderência

Sobre a base local `caea641`, a segunda fatia foi implementada sem commit ou deploy. `training-history-v1` registra no snapshot de novos rascunhos as sessões previstas já fechadas, concluídas, canceladas, pendentes vencidas e em andamento vencidas, além de sessões realizadas, minutos e carga por session-RPE em janelas cumulativas de 7, 28 e 42 dias. Sessões do dia ainda abertas não reduzem a aderência; campos nulos ou inválidos são registrados como falta de cobertura. Rascunhos e planos cancelados ficam fora do denominador de aderência, enquanto a carga inclui sessões concluídas do atleta pelo momento real de conclusão.

Essa leitura permanece observacional (`used_for_prescription: false`) e não alterou `rules-v1`, treinos, migrações, interface ou infraestrutura. Não foram criados limiares de aderência, ACWR, destreinamento, tolerância ou progressão. `go test -count=1 ./...`, `go vet ./...`, build do frontend, OpenAPI e a consulta PostgreSQL com fixtures sintéticas em transação somente leitura passaram. O `GET /v1/plans/current` confirmou a persistência de `training-history-v1`, das janelas 7/28/42 e das lacunas esperadas da conta de teste. Próxima etapa: definir a qualidade temporal e os critérios conservadores necessários antes de conectar essas medições à classificação de prontidão. A versão visível segue `0.7.0`.

### Continuidade — qualidade temporal e sinais protetivos

Após o commit local `a4e5f9f`, a terceira fatia evolui novos rascunhos para `training-history-v2`. O snapshot passa a registrar a recência da última sessão concluída, da última carga session-RPE válida e do último check-in; exclui e sinaliza registros futuros; diferencia ausência de feedback de campos incompletos; e conta dor, fadiga alta, RPE acima do planejado e sinais protetivos de recuperação nas janelas 7/28/42. Os cortes usados são apenas os já aplicados pelo `rules-v1` e não criam uma nova prescrição.

Tolerância à carga, destreinamento, mudança de condicionamento, atividades fora do Cadência, fuso do atleta e progressão continuam explicitamente não avaliados. Um intervalo sem atividade no aplicativo representa somente uma lacuna de registros: estudos de cessação ou redução controlada não autorizam concluir que o atleta parou de treinar. `used_for_prescription` permanece `false`; não houve mudança visual, migração ou infraestrutura, e a versão comunicada continua `0.7.0`. A suíte Go completa, `go vet`, build do frontend, OpenAPI e a consulta PostgreSQL sintética passaram; o lint geral preserva pendências anteriores fora desta fatia. A validação manual confirmou `training-history-v2`, as três janelas, o check-in protetivo, cinco feedbacks completos, `data_issues: []` e ausência correta de carga para sessões com duração zero. Próxima etapa: comparar períodos não sobrepostos de forma observacional antes de propor qualquer regra versionada de carga.

### Continuidade — comparação observacional entre períodos

Sobre o commit local `b0f32b7`, a quarta fatia foi implementada e validada no commit `810183c`. Novos rascunhos usam `training-history-v3` e registram `period_comparison` com seis períodos semanais não sobrepostos nos últimos 42 dias: `last_7d`, `days_8_14`, `days_15_21`, `days_22_28`, `days_29_35` e `days_36_42`. Cada período preserva as medições brutas de aderência, sessões realizadas, carga session-RPE, feedback, sinais protetivos e recuperação.

A comparação é somente observacional, com `period-comparison-v1`, `mode: observation` e `used_for_prescription: false`. Não foram adicionados tendência, ACWR, limiar de carga, tolerância, progressão, regressão ou inferência de destreinamento; `period_trend_for_prescription` permanece em `not_evaluated`. Sessões/carga usam intervalos de `completed_at` pelo relógio do banco, e aderência/recuperação usam datas relativas a `CURRENT_DATE`, sem fingir um fuso individual.

Os testes Go, `go vet`, build do frontend, validação estrutural do OpenAPI e a consulta PostgreSQL com fixtures sintéticas em transação somente leitura passaram. A conferência manual confirmou no `GET /v1/plans/current` a versão v3, os seis períodos e a invariância dos treinos. Não houve mudança visual, migração, infraestrutura ou prescrição; `APP_VERSION` e `UPDATE_NOTES` continuam em `0.7.0`.

### Continuidade — avaliação shadow do `rules-v2`

Após o commit `810183c`, a quinta fatia foi versionada no commit `64e554d`, ainda sem deploy. O snapshot de novos rascunhos passa a registrar `rules_v2_shadow`, uma avaliação paralela com versão `rules-v2`, modo `shadow` e escopo `plan_generation_only`; o `engine_version` permanece `rules-v1`.

A avaliação verifica de forma determinística a integridade dos períodos, sinais protetivos e evidência mínima de dois períodos recentes. Pode registrar `protective_signal`/`prefer_recovery`, `observation_only`/`maintain_observed` ou `not_evaluated`, sempre com regras adiadas, motivos, lacunas e inconsistências. Nenhum resultado é aplicado: `progression_eligible`, `applied` e `used_for_prescription` permanecem `false`, sem alterar duração, RPE, estímulo ou adaptação.

Essa primeira versão não calcula tolerância, destreinamento, mudança de condicionamento, ACWR ou resposta de atividades fora do Cadência. Os testes específicos e a suíte Go passaram nos pacotes executados; a execução agregada foi bloqueada apenas pelo Controle de Aplicativos do Windows ao abrir o executável temporário de `internal/repository`, mas o mesmo pacote passou compilado e executado dentro do workspace. `go vet`, build do frontend, OpenAPI e a consulta PostgreSQL sintética também passaram. A conferência manual confirmou no `GET /v1/plans/current` o sinal protetivo e todas as barreiras de não aplicação; a matriz controlada também confirmou dados completos, dados insuficientes, período inconsistente e invariância dos treinos do `rules-v1`. A alteração não é visual, então a versão comunicada continua `0.7.0`; a próxima funcionalidade visível deverá atualizar `APP_VERSION` e `UPDATE_NOTES`.

### Continuidade — primeiro desenho de adaptação pós-treino em shadow

Após a validação dos cenários do `rules-v2`, a sexta fatia foi versionada no commit `de23add`, ainda sem deploy. O `rules-v1` já possui um trigger SQL de adaptação pós-feedback; por isso, esta etapa não o substitui. O novo módulo `rules-v2-adaptation-v1` apenas avalia uma resposta candidata em Go e mantém `mode: shadow`, `progression_eligible: false`, `applied: false` e `used_for_prescription: false`.

A regra protege dor, esforço muito alto, fadiga máxima e sinais protetivos recentes. Uma resposta fácil isolada não autoriza progressão: ela fica em `defer_progression` quando faltam seis períodos íntegros, dois períodos recentes com carga e feedback completos e recuperação registrada. Com essa evidência, o shadow pode registrar `progress_duration_5pct`, sempre sem alterar a sessão. Respostas dentro do esperado ficam em `maintain_observed`; dados inválidos ou inconsistentes não produzem candidato.

Os cinco testes do novo avaliador passaram junto dos testes anteriores do shadow. Na sétima fatia, a avaliação é executada na mesma transação do feedback e gravada em `workouts.explanation.adaptation_shadow`; uma falha histórica é isolada por savepoint e registrada como `not_evaluated`. A resposta protetiva foi confirmada manualmente no `GET /v1/plans/current`, preservando o trigger existente e as barreiras de não aplicação.

### Continuidade — matriz controlada entre `rules-v1` e shadow

A oitava fatia cria uma matriz regressiva em `rules_v2_adaptation_comparison_test.go`. Ela compara dor, esforço alto, resposta neutra, recuperação recente, resposta fácil sem evidência, resposta fácil com evidência completa e histórico inconsistente. A proteção do `rules-v1` deve coincidir com a candidata protetiva do shadow; a progressão deve ser adiada sem evidência e continuar não aplicada mesmo quando a evidência é suficiente. O teste não altera prescrição, trigger, migração ou interface. A validação automatizada é o próximo passo desta fatia.

### Continuidade — integridade observacional da sessão

A nona fatia cria `data-integrity-v1` para classificar cada sessão concluída como `valid`, `incomplete` ou `inconsistent`. O gate separa dados ausentes de valores incompatíveis, verifica duração positiva, RPE, feedback, fadiga e métricas opcionais, e grava a leitura em `workouts.explanation.data_integrity` sem substituir o registro original. O shadow não produz candidato quando a sessão atual não é elegível para histórico; o `rules-v1` continua inalterado enquanto essa barreira é revisada. A validação manual confirmou uma sessão curta como `incomplete` e outra com duração positiva como `valid`, mantendo `used_for_prescription: false`.

### Continuidade — piloto publicado de intervalos aeróbicos XCO

A décima fatia amplia o catálogo de ciclismo com o protocolo `xco_aerobic_intervals`, selecionado somente para contexto `mtb_xco` explícito, atleta avançado, objetivo de performance ou prova, avaliação submáxima apta, disponibilidade de pelo menos 75 minutos, semana de construção e ausência de sinais protetivos. A sessão usa cinco blocos de 4 minutos com 4 minutos leves, RPE 7, sem sprint máximo, técnica de trilha, descida, salto ou meta rígida de potência. É uma adaptação conservadora do HIT estudado em mountain bikers treinados e da síntese de XCO contemporânea; não reproduz a carga original nem deve ser interpretada como receita universal.

A fonte `xco-hit-2016` é registrada na migração `000016`. Gravel permanece contextual, sem protocolo próprio baseado apenas em estudo de campo; pista sprint/BMX e downhill/enduro permanecem fora do motor. Como a escolha do protocolo é visível, `frontend/lib/release.ts` foi atualizado para `0.8.0` com nota na tela de novidades. A fatia foi validada localmente, publicada no commit `9d8c624` e registrada na release `v0.8.0`; a observação seguinte deve acompanhar o uso do piloto para XCO elegível e sua ausência para estrada, limitações, semana de recuperação e perfil sem avaliação apta.

### Continuidade — rotação segura da recuperação ativa — 10 de setembro de 2026

A décima primeira fatia, versionada no commit local `ac8ed3b`, reduz a repetição nominal do catálogo sem criar uma nova modalidade. Uma sessão de base da quarta semana passa a poder ser apresentada como `Recuperação ativa`, com alvo RPE 3,5, volume reduzido pelo multiplicador de recuperação e instrução de pedal leve contínuo. O protocolo usa a evidência geral `acsm-1998` apenas para progressão gradual e controle de carga; os minutos são uma escolha conservadora do produto, não uma dose universal de estudo.

### Continuidade — registro explícito de treino não realizado — 11 de setembro de 2026

A décima segunda fatia adiciona uma ação factual para treinos `planned` ou `adapted` cuja data já passou. Após confirmação do atleta, a rota `POST /v1/workouts/{workoutID}/missed` marca o treino como `skipped`, fecha a pendência de aderência e preserva a data original. Ela não cria uma sessão fictícia, não reagenda o treino, não substitui a sessão nem altera a carga seguinte; também não transforma ausência de registro em inferência de destreinamento.

Treinos futuros e estados já iniciados ou encerrados são protegidos por transições inválidas. O cancelamento de uma sessão em andamento continua separado. A funcionalidade é visível, então `frontend/lib/release.ts` foi atualizado para `0.10.0` e a tela de novidades explica o comportamento. Os testes automatizados do backend, build, componente, OpenAPI, PostgreSQL transacional, histórico e a confirmação visual local passaram. O commit `051d285` foi feito; não houve deploy, migração ou mudança de infraestrutura.

### Continuidade — piloto de intervalos intensos para estrada — 11 de setembro de 2026

A décima terceira fatia amplia o catálogo com `road_high_intensity_intervals`, apresentado como **Intervalos intensos de estrada**. O protocolo é liberado somente para atleta avançado com disciplina `road`, pelo menos oito semanas e três pedais semanais recentes, objetivo de performance/evento, avaliação submáxima apta, pelo menos 75 minutos disponíveis, semana de construção e ciclo alternado. A sessão usa até cinco blocos de 8 minutos com 4 minutos leves, RPE 8, sem sprint máximo ou meta rígida de potência; a estrutura pode reduzir blocos quando a duração adaptada exigir.

A escolha usa `road-block-comparison-2025` e `rosenblat-2020` como referências, mas não copia o bloco concentrado de ciclistas bem treinados nem o transforma em dose universal. Dor, limitação, recuperação insuficiente, baixa consistência, dados inelegíveis e níveis abaixo do avançado continuam bloqueando o piloto. Como a funcionalidade é visível, `frontend/lib/release.ts` passou para `0.11.0` e a tela de novidades foi atualizada. A implementação foi registrada no commit local `4312fa9`; não houve deploy, migração ou mudança de infraestrutura.

A escolha é determinística e não altera os protocolos específicos de estrada ou XCO. Limitação ativa, dor ou necessidade recente de recuperação continuam substituindo qualquer variação por `Giro leve protegido`. A validação local passou em `go test -count=1 ./...`, `go vet ./...` e `npm run build`. Antes do deploy posterior da versão `0.12.0`, a produção estava no commit `9d8c624` e na versão `0.8.0`; o planejamento raiz foi preservado.

### Continuidade — proteção explícita da semana de recuperação — 11 de setembro de 2026

A correção seguinte impede que o slot de qualidade seja criado na quarta semana do ciclo. O pedal mais longo continua como `Endurance contínuo`, e os demais slots usam `Recuperação ativa`; assim, meta de prova, avaliação submáxima ou preferência por intervalos não introduzem `Ritmo de prova controlado`, tempo ou intervalos na semana marcada como recuperação.

A mudança é visível, então `frontend/lib/release.ts` foi atualizado para `0.11.1` com uma nota de novidades. O teste de regressão para contexto de prova, a suíte Go completa e `go vet` passaram. Esta correção está pendente de commit e não foi publicada na produção; não houve migração nem mudança de infraestrutura.

### Continuidade — auditoria de segurança e correções locais — 11 de setembro de 2026

A auditoria do código identificou e corrigiu o bloqueio de início de sessões adaptadas, limites e validação dos corpos JSON, rate limiting das rotas de autenticação, proteção de origem para mutações autenticadas, cabeçalhos HTTP de segurança, enumeração de contas no cadastro/recuperação, risco de hidratação no horário de início e concorrência entre execuções do resumo semanal. A trava advisory transacional impede timer e execução manual simultâneos; uma queda exatamente após o envio e antes do commit ainda exige conferência manual, pois não há transação distribuída entre PostgreSQL e Resend.

A data da prova agora é validada no fuso local, não aceita datas passadas e controla a janela da fase específica: eventos próximos podem orientar o estímulo de prova, enquanto eventos distantes permanecem na progressão regular. A data e a distância são preservadas no snapshot. A tela de novidades foi ajustada para marcar a leitura somente após a dispensa e a versão local passou para `0.12.0`.

Os testes locais `go test -count=1 ./...`, `go vet ./...`, `npm run build`, lint direcionado dos componentes alterados e `npm audit --omit=dev --audit-level=high` passaram. O lint geral ainda conserva pendências anteriores; o grafo de desenvolvimento mantém avisos sem correção automática disponível. O compose fixa o digest do `cloudflared 2026.7.3` ARM64 validado na VPS. As alterações foram commitadas em `e803d46` e `61d7939`, publicadas na VPS e registradas na release `v0.12.0`. O backup preventivo foi `cadencia-20260911T234851Z.dump`; não houve migração nova.

### Continuidade — implementação local do taper pré-prova — 11 de setembro de 2026

A próxima fatia do catálogo foi implementada localmente como `taper-v1`, uma redução de volume orientada por evento, sem criar uma nova modalidade. O taper só pode ser prescritivo quando há evento futuro entre 7 e 21 dias, nível avançado, avaliação submáxima apta, pelo menos oito semanas de treino e três pedais semanais informados, sem limitação ativa, dor ou necessidade recente de recuperação.

Quando elegível, o plano mantém sua frequência e seu RPE-alvo, mas reduz em 50% a duração das sessões anteriores ao evento e dentro dos 14 dias que o antecedem, respeitando o mínimo de 20 minutos. O dia do evento fica fora da redução, a quarta semana continua sendo de recuperação e não há sobrecarga automática antes do taper. A decisão e os motivos ficam no snapshot; as sessões afetadas registram `event_taper_applied` e as evidências correspondentes.

A migração `000017` registra as três fontes do taper e foi aplicada somente no PostgreSQL de desenvolvimento local existente. `go test -count=1 ./...`, `go vet ./...`, `npm run build` e `git diff --check` passaram. A versão local foi atualizada para `0.13.0`, a tela de novidades foi preparada e o conjunto foi registrado no commit local `0eb34d6`; produção permanece em `0.12.0` e `000016`. A validação no navegador confirmou o evento elegível em `20/09/2026`, oito dias de distância, cinco sessões pré-prova reduzidas pela metade e preservação de frequência/RPE; a suíte automatizada confirmou a não aplicação para evento distante e cenário protegido. O rascunho não foi aceito. Próxima etapa: ampliar o catálogo de protocolos de ciclismo com novas evidências, mantendo o taper local fora da produção até autorização explícita.

### Continuidade — piloto local de intervalos VO₂max de estrada — 12 de setembro de 2026

Após a validação do taper, a próxima expansão do catálogo foi implementada localmente como `road_vo2_intervals`. A opção aparece no perfil como preferência explícita `VO₂max`, mas o motor só a libera para disciplina `road`, atleta avançado, objetivo `performance` ou `event`, avaliação submáxima apta, pelo menos oito semanas de treino recente, três pedais semanais, 60 minutos disponíveis, semana de construção e ausência de sinais de proteção.

A sessão usa quatro blocos de 4 minutos em RPE 8, com quatro minutos leves entre os blocos, aquecimento e desaquecimento. Não há sprint máximo, cadência baixa obrigatória, potência fixa ou conversão de frequência cardíaca em VO₂max. Dor, limitação, recuperação insuficiente, evento fora da fase específica, perfil intermediário, histórico insuficiente e outras modalidades não recebem o piloto; `rules-v1` continua como única autoridade prescritiva.

A migração `000018_road_vo2_catalog_evidence` foi criada e aplicada somente no PostgreSQL local, registrando `road-vo2-intervention-2024` e `road-vo2-response-2024`. A versão visível local passou para `0.14.0` e a tela de novidades foi atualizada. Os testes direcionados de planejamento e onboarding passaram; ainda faltam a suíte completa, o build final, a conferência ponta a ponta da preferência e das fontes via API/navegador e o registro do resultado. Produção continua em `0.12.0`/`000016`, sem deploy ou mudança de infraestrutura. Após essa validação, oferecer o commit e só então avaliar o próximo protocolo do catálogo.

### Continuidade — validação do piloto VO₂max de estrada — 12 de setembro de 2026

A validação local do piloto `road_vo2_intervals` foi concluída. A conta de teste salvou a preferência explícita `VO₂max` no perfil, a API local aceitou o contexto depois da reinicialização do binário atual e a geração do plano apresentou **Intervalos VO₂max de estrada** para o cenário elegível. O teste também confirmou que o `400` observado anteriormente vinha de um processo antigo da API, não de dados inválidos do formulário.

`go test -count=1 ./...`, `go vet ./...`, `npm run build` e `git diff --check` passaram. A implementação e a migração `000018` estão no commit `01875c9`; a versão local permanece `0.14.0`, com a nota de novidades atualizada. Produção continua em `0.12.0`/`000016`, sem deploy, alteração de infraestrutura ou publicação da release.

Próxima etapa: avaliar o próximo candidato do catálogo com pesquisa específica, elegibilidade, limites de segurança e teste de não seleção antes de implementar outro protocolo. O taper e o piloto VO₂max devem permanecer locais até revisão do catálogo, backup e autorização explícita de publicação.

### Continuidade — pesquisa do candidato de intervalos curtos — 12 de setembro de 2026

A pesquisa do próximo protocolo recomenda avaliar intervalos curtos autorregulados para estrada ou indoor, sem sprint máximo. Um ensaio de 2025 comparou, em adultos anteriormente inativos, formatos de 30 segundos com 120 segundos de recuperação e de 1 minuto com 1 minuto de recuperação; ambos melhoraram o VO₂peak. Estudos em ciclistas treinados e uma meta-análise recente ajudam a contextualizar o formato, mas as populações, doses e níveis de controle são diferentes do público geral do Cadência.

Por isso, o candidato permanece somente documentado. Antes de qualquer código, será necessário definir a população elegível, avaliação apta, histórico mínimo, número máximo de blocos, recuperação, RPE, interrupção, uma única sessão de qualidade e travas por dor, fadiga e recuperação. Não foram criados protocolo, migração, nota de versão ou alteração de produção nesta etapa.

### Continuidade — especificação preliminar do candidato de intervalos curtos — 12 de setembro de 2026

A especificação preliminar foi registrada em `docs/training-adaptation-rules.md`. Ela propõe, para eventual primeiro piloto, estrada ou indoor, preferência explícita, nível avançado, objetivo de performance/evento, avaliação submáxima apta, oito semanas e três pedais semanais, 50 minutos disponíveis, semana de construção e ausência de proteções. A estrutura candidata é aquecimento de 10 minutos, seis repetições de 1 minuto em RPE 7–8 com recuperação leve autorregulada e desaquecimento de 10 minutos; não há sprint máximo, potência fixa, cadência obrigatória ou progressão automática.

A mesma especificação define interrupção por sinais preocupantes, prioridade das travas de recuperação, uma única sessão de qualidade, registro auditável e a matriz mínima de testes positivos e de não seleção. O candidato continua fora do código, do `rules-v1`, da migração, da versão visível e da produção. Próxima etapa: revisar a especificação e, se aprovada, implementar em uma fatia isolada com testes antes de qualquer publicação.

### Continuidade — implementação local do piloto de intervalos curtos — 12 de setembro de 2026

A especificação foi implementada localmente como `short_self_regulated_intervals`. A preferência explícita `short_intervals` aparece no perfil; o motor libera o piloto somente para estrada ou indoor, atleta avançado, objetivo de performance/evento, avaliação submáxima apta, oito semanas e três pedais semanais, disponibilidade mínima de 50 minutos, semana de construção, fase compatível com evento e ausência de proteções. A sessão usa seis repetições de 1 minuto em RPE 7,5 com 1 minuto leve entre elas, aquecimento e desaquecimento, sem sprint máximo, potência fixa ou cadência obrigatória.

A migração `000019_short_intervals_evidence` registra as fontes do ensaio de 2025 e do estudo com ciclistas de 2020. A versão local foi atualizada para `0.15.0` e a tela de novidades foi preparada. Foram adicionados testes positivos, de não seleção, de proteção por dor e de validação da preferência; a suíte Go, o `go vet`, o build e a verificação da migração local passaram. A opção e a nota aparecem no navegador. O lint geral continua com pendências antigas fora desta fatia. Produção continua em `0.12.0`/`000016`, sem deploy ou alteração de infraestrutura.

### Continuidade — validação manual do piloto de intervalos curtos — 12 de setembro de 2026

A validação manual local foi concluída pelo proprietário. Após marcar `Intervalos curtos` no perfil, salvar as alterações e atualizar o plano, a preferência foi aceita e o plano apresentou **Intervalos curtos autorregulados**. A estrutura exibida corresponde ao piloto de seis blocos de 1 minuto em RPE 7,5, com recuperações leves e proteção do `rules-v1`.

O piloto está validado no checkout local, mas continua fora da produção junto com a migração `000019` e a versão local `0.15.0`. A próxima etapa é revisar o catálogo ampliado e decidir, com backup e autorização explícita, se os pilotos devem ser publicados.

### Continuidade — avaliação de resistência específica na bicicleta — 12 de setembro de 2026

A pesquisa seguinte encontrou um ensaio randomizado de 2025 com 37 ciclistas bem treinados, comparando resistência fora da bicicleta com esforços na própria bicicleta durante dez semanas. O protocolo na bicicleta usou resistência muito alta, cadência muito baixa e carga calibrada por força dinâmica máxima; houve melhora de força e potência, mas não de VO₂max.

O estímulo é relevante para o catálogo, porém não pode ser transformado em prescrição do Cadência nesta etapa: o produto não mede força dinâmica máxima nem calibra a resistência estudada, e RPE isolado não representa essa carga. O candidato fica documentado, sem código, preferência, migração ou nota de versão. Próxima etapa: avaliar se vale criar primeiro os pré-requisitos de medição e segurança; gravel/XCM continuam contextos de endurance sem protocolo próprio, enquanto sprint de pista/BMX e downhill/enduro estão fora do produto.

### Continuidade — exclusão permanente de sprint/pista/BMX e downhill/enduro — 12 de setembro de 2026

Foi definida a exclusão permanente de sprint/pista/BMX e downhill/enduro do Cadência. As opções foram removidas do perfil, os valores `track_sprint` e `dh_enduro` deixaram de ser aceitos pela API e registros legados são tratados como disciplina não informada no carregamento do perfil, sem liberar protocolos.

A mudança visível atualizou a versão local para `0.16.0` e a tela de novidades. Não houve migração, deploy ou alteração de infraestrutura; a produção continua em `0.12.0`/`000016`.

### Continuidade — publicação do catálogo ampliado e exclusão de modalidades — 12 de setembro de 2026

Após a revisão do catálogo local, foram publicados na VPS os pilotos `taper-v1`, `road_vo2_intervals` e `short_self_regulated_intervals`, junto da decisão de excluir sprint/pista/BMX e downhill/enduro. O commit `6fdbe45` foi buscado por fast-forward, o backup `cadencia-20260912T155746Z.dump` foi criado e verificado, e as migrações `000017`, `000018` e `000019` foram aplicadas pelo perfil `maintenance` antes da recriação somente de API e frontend. PostgreSQL e Tunnel permaneceram ativos.

A versão comunicada passou para `0.16.0`. A API respondeu `/health` e `/ready`, os dois domínios oficiais retornaram HTTP 200 a partir da VPS e os quatro serviços ficaram saudáveis. Não houve alteração de infraestrutura além da atualização controlada das imagens da API e do frontend. A release [v0.16.0](https://github.com/Saulorangel87/App-de-treino/releases/tag/v0.16.0) foi criada no GitHub.

### Próxima etapa — pós-publicação do catálogo `0.16.0` — 12 de setembro de 2026

O smoke test visual da nota de novidades e do perfil em produção foi validado pelo proprietário, confirmando a presença dos pilotos permitidos e a ausência de sprint/pista/BMX e downhill/enduro. A release [v0.16.0](https://github.com/Saulorangel87/App-de-treino/releases/tag/v0.16.0) foi criada no GitHub apontando para o commit publicado `6fdbe45`.

Depois da conferência inicial, a operação deve observar os pilotos de taper, VO₂max de estrada e intervalos curtos dentro dos gates documentados, enquanto o primeiro resumo semanal do Resend e os relatos reais seguem em paralelo. A próxima fatia de código deve retomar a evolução em shadow de adaptação, carga/progressão e integridade dos dados, sem substituir o `rules-v1` antes de haver comparação, testes e auditabilidade suficientes.

### Continuidade — observação de tolerância à carga — 12 de setembro de 2026

Como o primeiro acompanhamento de produção foi validado e o resumo do Resend depende do uso real, a próxima fatia técnica foi iniciada localmente com `load-tolerance-v1`. O avaliador observa os dois períodos semanais mais recentes, não sobrepostos, exigindo em cada um carga por session-RPE, feedback completo e ao menos um check-in de recuperação completo. Ele também considera a sessão recém-concluída: dor, fadiga alta, recuperação necessária ou esforço pelo menos dois pontos acima do RPE-alvo mantêm a resposta protetiva.

Com os dados completos e sem sinal protetivo, o resultado é somente `observation_only`/`maintain_observed`; não representa tolerância fisiológica comprovada nem autoriza progressão. A leitura foi anexada a `workouts.explanation.adaptation_shadow`, mantendo `rules-v1`, o trigger pós-feedback e as sessões inalterados. Não houve migração, mudança visual, atualização de versão, alteração de infraestrutura, commit ou deploy nesta fatia.

Foram adicionados testes para evidência completa, lacuna de carga, esforço alto, períodos inconsistentes e isolamento prescritivo. `go test -count=1 ./...`, `go vet ./...`, `npm run build` e `git diff --check` passaram; o lint geral continua com pendências antigas fora desta mudança. A API local foi reiniciada com a implementação nova e, após uma sessão concluída, o `GET /v1/plans/current` confirmou `explanation.adaptation_shadow.load_tolerance` com `version: "load-tolerance-v1"`, `progression_eligible: false` e `used_for_prescription: false`; `status: "not_evaluated"` foi aceito como esperado para a conta de teste. A implementação foi registrada no commit `36cfabb` e ainda não foi publicada em produção.

A etapa seguinte foi documentada e especificada como uma fatia observacional de adaptação em ciclo fechado, comparando o treino planejado com o realizado, sem substituir o motor prescritivo `rules-v1`.

### Continuidade — comparação planejado versus realizado — 12 de setembro de 2026

A fatia seguinte foi implementada localmente como `planned-vs-actual-v1`. Ao concluir uma sessão, o backend registra no `workouts.explanation.adaptation_shadow` a duração planejada e realizada, a diferença percentual de duração, o RPE-alvo e realizado, a diferença de RPE, a cobertura das métricas disponíveis e as lacunas ainda não coletadas. O bloco é descritivo, mantém `progression_eligible: false` e `used_for_prescription: false`, e não altera o `rules-v1`, o trigger, as sessões ou a interface.

Foram adicionados testes para comparação válida, dados essenciais incompletos, métricas opcionais ausentes e isolamento prescritivo. `go test -count=1 ./...`, `go vet ./...`, `npm run build` e `git diff --check` passaram. A validação manual via API local confirmou `explanation.adaptation_shadow.planned_vs_actual` com 3 minutos realizados de 35 planejados, `status: "observed"`, diferença de duração e RPE registrada, além de `progression_eligible: false` e `used_for_prescription: false`. Não houve migração, atualização de versão ou deploy desta fatia; o código foi registrado no commit `6619565`.

### Continuidade — contexto de conclusão parcial — 12 de setembro de 2026

A próxima fatia técnica foi implementada localmente para registrar se o ciclista realizou o treino completo ou apenas parte dele. Quando a conclusão é parcial, o formulário exige um motivo controlado: falta de tempo, fadiga/recuperação, dor/desconforto, equipamento/clima/terreno ou outro motivo. O resumo da sessão e o histórico também exibem esse contexto.

O campo foi incorporado ao feedback, ao bloco `planned-vs-actual-v1` e à integridade observacional. Uma conclusão parcial não é evidência de tolerância ao treino completo e não libera progressão; as proteções já existentes para dor, fadiga e esforço alto continuam válidas. O trigger atualizado na migração `000020` mantém o `rules-v1` como motor prescritivo e bloqueia somente a ramificação de progressão para sessões parciais.

A funcionalidade visível atualizou a versão local para `0.17.0` e a tela de novidades. A migração `000020_completion_context` foi aplicada no PostgreSQL local, o teste SQL transacional passou e a validação manual confirmou uma sessão parcial com o motivo `Fiquei sem tempo`, registrada no histórico como `Parcial · sem tempo`; a progressão não foi liberada. O contraste dos menus e do placeholder do motivo também foi corrigido. A suíte Go, o `go vet`, o build do frontend e a verificação do diff passaram; o lint geral mantém pendências preexistentes fora desta fatia. A implementação foi registrada em `58e3887`, com os ajustes visuais em `b5ef413` e `d7ce5bb`. A versão `0.17.0` e a migração `000020` continuam somente locais; não houve deploy ou alteração de infraestrutura.

### Continuidade — feedback pós-treino com recuperação e confiança — 12 de setembro de 2026

A próxima fatia amplia o feedback fisiológico observado com `recovery_after` e `repeat_confidence`, ambos em escala de 1 a 5. O primeiro registra a recuperação percebida após a sessão e o segundo a confiança para repetir o treino. O formulário atual envia uma resposta neutra inicial, enquanto a migração `000021` mantém os campos opcionais para preservar históricos antigos e impõe as restrições de faixa.

Os valores aparecem no resumo, no histórico e em `planned-vs-actual-v1` quando presentes, mas não alteram `rules-v1`, o trigger, a duração, o RPE, o estímulo ou o status das próximas sessões. A interpretação continua observacional e não substitui o check-in diário de sono, estresse e fadiga. `go test -count=1 ./...`, `go vet ./...`, `npm run build`, o teste SQL transacional e `git diff --check` passaram. A validação manual confirmou o registro e a exibição dos dois campos; a implementação foi registrada no commit `34f17b3`. A versão local permanece em `0.18.0`, sem deploy ou alteração de infraestrutura.

### Continuidade — histórico permanente de novidades — 12 de setembro de 2026

A próxima melhoria de interface reduz o tamanho do aviso exibido após uma atualização e cria a rota autenticada `/novidades`. O aviso mostra somente as notas da versão atual e oferece acesso ao histórico completo; a nova página agrupa as notas por versão e mantém as versões anteriores recolhidas. O menu lateral, o menu móvel e os cabeçalhos das telas internas oferecem acesso permanente ao histórico.

As notas continuam centralizadas em `frontend/lib/release.ts`, agora com a versão de cada item. A versão local passou para `0.19.0` para comunicar a nova funcionalidade na tela de primeiro acesso. O build, o lint específico e a formatação dos arquivos da interface passaram; a validação visual foi concluída pelo proprietário. Produção permanece em `0.16.0`/`000019`, sem deploy ou alteração de infraestrutura. A implementação foi registrada no commit `8798ea3`.

### Continuidade — contexto pós-treino no `adaptation_shadow` — 12 de setembro de 2026

A próxima fatia técnica integra `recovery_after` e `repeat_confidence` à avaliação paralela `rules-v2-adaptation-v1` por meio do bloco `post-workout-context-v1`. A avaliação registra se os sinais estão completos, parciais, ausentes ou inválidos e explicita os campos observados, as lacunas, os problemas e o motivo da classificação. Ela não interpreta fisiologia, não define limiares de tolerância ou progressão e não substitui o check-in diário.

Mesmo com os dois sinais válidos, o resultado permanece `candidate_response: maintain_observed`, `progression_eligible: false` e `used_for_prescription: false`. O `rules-v1`, o trigger, o banco e o comportamento das próximas sessões permanecem inalterados; não foi criada migração nem houve mudança visual ou de versão. A tipagem TypeScript e o contrato OpenAPI foram atualizados para leitura do novo bloco.

Os testes cobrem contexto completo, parcial e inválido, além da anexação ao shadow sem torná-lo autoritativo. `go test -count=1 ./...`, `go vet ./...`, `npm run build` e `git diff --check` passaram. A validação manual via API local confirmou os dois sinais com `status: "observed"`, sem progressão ou aplicação prescritiva. A implementação foi registrada no commit `46a900f`; não houve deploy ou alteração de infraestrutura. Próxima etapa: registrar a proveniência da decisão shadow antes de qualquer uso prescritivo.

### Continuidade — auditoria da decisão em shadow — 12 de setembro de 2026

A etapa seguinte tornou a avaliação `rules-v2-adaptation-v1` mais auditável com o bloco `adaptation-audit-v1`. Ele registra os dados usados, as lacunas, as restrições aplicadas, os caminhos de adaptação não selecionados e as condições informativas para uma futura revisão. O campo `confidence` fica em `not_calibrated`, pois ainda não há calibração estatística nem evidência de efeito que permita apresentar um nível de confiança.

O bloco não cria gatilhos, não muda a duração, o RPE, o estímulo ou o status das próximas sessões e mantém `used_for_prescription: false`. O `rules-v1`, o trigger e a infraestrutura permanecem inalterados; não houve migração, mudança visual ou atualização da versão visível. A tipagem TypeScript e o contrato OpenAPI foram atualizados.

Foram adicionados testes de proveniência, lacunas e serialização JSON. `go test -count=1 ./...`, `go vet ./...`, `npm run build`, `oxlint` da tipagem alterada e `git diff --check` passaram. Esta fatia permanece local, sem deploy. Próxima etapa: validar manualmente o novo bloco no `GET /v1/plans/current` após reiniciar a API atual e, depois, oferecer o commit.

### Continuidade — correção do acesso ao perfil e da versão das novidades — 13 de setembro de 2026

A validação manual da auditoria shadow foi concluída. O bloco `planned_vs_actual-v1` registrou 3 minutos realizados de 31 planejados, sem `data_issues`, com estado `observed`, `progression_eligible: false` e `used_for_prescription: false`. O RPE realizado acima do alvo acionou a proteção observacional de esforço alto, sem substituir o `rules-v1` ou alterar a prescrição fora das proteções já existentes.

Também foi corrigida a interface: a barra lateral preserva o tamanho dos menus, pode rolar quando a altura da janela é insuficiente e mantém o acesso ao perfil fora da área coberta pelo rodapé fixo. A mensagem `Plano explicável` foi preservada. A versão principal foi corrigida para `0.20.0`, e a tela de novidades passou a exibir a nota dessa atualização. `npm run build` e `git diff --check` passaram; a confirmação visual local validou o aviso `NOVIDADES · V0.20.0`, o menu lateral e o rodapé. A correção foi registrada no commit `2828049`; não houve deploy, migração ou alteração de infraestrutura.

Próxima etapa: iniciar a próxima melhoria técnica do shadow somente após este registro documental, mantendo `rules-v1` como único motor prescritivo e a produção em `0.16.0` até revisão e autorização explícita.

### Continuidade — filtro de integridade no histórico observado — 13 de setembro de 2026

A auditoria da fatia de integridade encontrou uma lacuna de integração: `data-integrity-v1` gravava `eligible_for_history: false` para sessões incompletas ou inconsistentes, e o shadow da sessão atual respeitava esse resultado, mas as consultas históricas ainda podiam incluir esses registros em agregados de carga, dor, fadiga e recência. Isso permitia que uma observação já classificada como inelegível influenciasse a leitura de períodos posteriores.

As consultas de resumo observado, janelas cumulativas de 7/28/42 dias, seis períodos semanais e agregados da tela de Evolução agora usam somente sessões sem marca explícita de inelegibilidade. A aderência planejada continua separada, e os treinos legados sem `data_integrity` permanecem legíveis. Nenhum registro original é apagado ou corrigido automaticamente; o filtro apenas impede que a sessão inelegível alimente métricas observacionais usadas na análise do plano, do shadow e da Evolução.

A fixture `scripts/test-training-history-query.ps1` passou a conter uma sessão com `eligible_for_history: false` e confirmou a exclusão dos seus minutos, carga session-RPE, dor e fadiga nas janelas e períodos. `go test -count=1 ./...`, `go vet ./...`, `npm run build`, a consulta PostgreSQL em transação somente leitura e `git diff --check` passaram. Não houve migração, alteração visual, atualização de versão, deploy ou mudança de infraestrutura.

Próxima etapa: revisar o diff desta fatia e, após o commit, avaliar a próxima evolução observacional de adaptação/carga sem transferir autoridade ao `rules-v2`.

### Continuidade — consistência temporal do resumo observado — 13 de setembro de 2026

A revisão após o filtro de integridade encontrou uma segunda inconsistência: as janelas cumulativas, os períodos semanais e a qualidade temporal já excluíam sessões com `completed_at` no futuro, mas o resumo observado de 28 dias carregado em `PlanningContextByUserID` não tinha o limite superior `completed_at <= now()`.

O resumo agora usa a mesma referência temporal das demais consultas. Uma sessão futura não influencia sessões concluídas, minutos, RPE médio, fadiga, dor ou cobertura de dados; sessões inelegíveis por `data-integrity-v1` continuam excluídas. A aderência planejada permanece separada e nenhum registro original é removido.

O teste `scripts/test-readiness-queries.ps1` foi alinhado à consulta atual e passou em transação PostgreSQL somente leitura com cenários de sessão inelegível, futura, fora da janela, cancelada e conta sem registros. `go test -count=1 ./...`, `go vet ./...` e `git diff --check` também passaram. Não houve migração, alteração visual, atualização de release, deploy ou mudança de infraestrutura.

Esta correção é somente local e aguarda o commit do proprietário. Próxima etapa: revisar os agregados observacionais restantes com a mesma consistência temporal antes de propor qualquer ampliação prescritiva do `rules-v2`.

### Continuidade — consistência temporal da tela de Evolução — 13 de setembro de 2026

A revisão dos agregados restantes encontrou a mesma classe de risco na tela de Evolução: o resumo total, a série semanal, as sessões recentes e os check-ins não tinham uma barreira uniforme contra eventos futuros. Uma correção de relógio ou dado inconsistente poderia aparecer como atividade já realizada.

As quatro consultas agora aceitam somente sessões concluídas com `completed_at <= now()`, sessões canceladas com `cancelled_at <= now()` e check-ins com `recorded_on <= CURRENT_DATE`. A exclusão de sessões explicitamente inelegíveis por `data-integrity-v1` e o isolamento por atleta permanecem preservados. Nenhum registro é apagado e a Evolução continua somente observacional.

O novo teste `scripts/test-evolution-queries.ps1` executa as consultas reais com CTEs sintéticas em transação somente leitura. Ele confirmou o resumo total, a semana atual, as sessões recentes e os pontos de recuperação sem sessões/check-ins futuros ou dados de outro atleta. Não houve migração, alteração visual, atualização de release, deploy ou mudança de infraestrutura.

Esta fatia está validada localmente e aguarda o commit do proprietário. Próxima etapa: revisar os demais agregados observacionais e, só depois, continuar a avaliação shadow de adaptação/carga sem transferir autoridade ao `rules-v2`.

### Continuidade — gate de tolerância integrado ao shadow de adaptação — 13 de setembro de 2026

A revisão da integração encontrou uma lacuna: `load-tolerance-v1` já classificava esforço acima do alvo no período mais recente como sinal protetivo, mas `rules-v2-adaptation-v1` não usava esse resultado para bloquear sua candidata de progressão. Uma sessão atual fácil poderia, portanto, coexistir com uma resposta recente incompatível com aumento de carga.

O shadow agora incorpora esse gate. Quando a tolerância observada está protetiva, ou quando o período recente contém `AboveTargetRPESessions`, a resposta candidata fica em `protective_signal`/`prefer_recovery` e registra `recent_above_target_rpe`; `load_tolerance_gate` também aparece em `rules_evaluated`. A alteração continua observacional e não modifica o `rules-v1`, o trigger, a sessão seguinte ou qualquer prescrição.

A matriz comparativa foi ampliada com o cenário de resposta fácil e esforço acima do alvo recente. Os testes direcionados de planejamento, repositório e HTTP, `go vet` e `git diff --check` passaram. Não houve migração, mudança visual, atualização de release, deploy ou alteração de infraestrutura.

Esta fatia está validada localmente e aguarda o commit do proprietário. Próxima etapa: continuar a revisão dos gates de adaptação em shadow, especialmente a coerência entre contexto de conclusão, integridade, carga e evidência, antes de qualquer integração prescritiva.

### Continuidade — conclusão parcial explícita no shadow — 13 de setembro de 2026

A revisão da coerência entre contexto de conclusão, integridade, carga e evidência encontrou uma lacuna de auditabilidade: uma sessão `partial` já não liberava a adaptação ativa do `rules-v1`, mas a avaliação `rules-v2-adaptation-v1` podia terminar como `observation_only`/`maintain_observed`, sem deixar explícito no resultado principal que a progressão deveria ser adiada.

O shadow agora prioriza as proteções de dor, fadiga e esforço alto e, na ausência delas, aplica o gate de conclusão parcial: `status: not_evaluated`, `candidate_response: defer_progression` e motivo `partial_completion`. A auditoria também registra `load_tolerance_gate` quando o avaliador de tolerância está protetivo. Isso não interpreta o motivo da interrupção como diagnóstico nem transforma a sessão parcial em evidência de tolerância ao treino completo.

Foram adicionados testes para a conclusão parcial e para a proveniência do gate de tolerância. `go test -count=1 ./...`, `go vet` e `git diff --check` passaram. `rules-v1`, trigger, migrações, banco, interface, release, infraestrutura e produção permanecem inalterados; `progression_eligible`, `applied` e `used_for_prescription` continuam falsos. Esta fatia está validada localmente e aguarda o commit do proprietário. Próxima etapa: continuar a revisão dos gates de adaptação em shadow, mantendo a separação entre observação e prescrição.

### Continuidade — fidelidade da proveniência na auditoria shadow — 13 de setembro de 2026

A revisão seguinte encontrou uma limitação de auditabilidade: `adaptation-audit-v1` copiava apenas as lacunas do resultado principal e declarava campos básicos como `data_used` mesmo quando o feedback era inválido. As lacunas específicas de `post_workout_context`, `load_tolerance` e `planned_vs_actual` podiam ficar restritas aos blocos aninhados, dificultando uma leitura única da decisão.

A auditoria agora consolida essas lacunas quando os blocos estão disponíveis, identifica `training_history_periods` quando o histórico foi avaliado e só lista os campos básicos como usados quando o feedback e o RPE passam pela validação mínima. Dados inválidos não são apresentados como fundamento da decisão; a avaliação continua observacional e o `rules-v1` permanece prescritivo.

Foram adicionados testes para lacunas aninhadas, feedback inválido e proveniência do gate de tolerância. `go test -count=1 ./...`, `go vet` e `git diff --check` passaram. Não houve mudança visual, release, migração, infraestrutura ou deploy. Esta fatia está validada localmente e será registrada no commit autorizado. Próxima etapa: continuar a revisão de coerência entre integridade, carga, contexto pós-treino e evidência.
