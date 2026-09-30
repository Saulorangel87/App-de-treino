# Strava: o que as regras permitem ao Cadência

Leitura feita em 30/09/2026 dos documentos oficiais em vigor: [Strava API Agreement](https://www.strava.com/legal/api) e [Strava API Policy](https://www.strava.com/legal/api_policy) (ambos com vigência a partir de 1º de junho de 2026), [Brand Guidelines](https://developers.strava.com/guidelines/) (revisão de 29/09/2025), [Getting Started](https://developers.strava.com/docs/getting-started/), [Authentication](https://developers.strava.com/docs/authentication/), [Webhooks](https://developers.strava.com/docs/webhooks/) e [Rate Limits](https://developers.strava.com/docs/rate-limits/). As citações abaixo são literais. Este documento não é parecer jurídico; onde o texto é ambíguo, está marcado como pergunta para o Strava.

A Policy faz parte do Agreement e pode mudar a qualquer momento; o uso continuado vale como aceite. Reler os dois antes de cada entrega da integração.

## Conclusão em uma frase

A integração é viável para **mostrar ao atleta as próprias atividades e ajudá-lo a registrar o treino**, mas as regras **proíbem guardar dados do Strava por mais de 7 dias, usá-los em IA e usá-los para análise ou melhoria do produto**. Isso tira do Strava o papel que eu tinha proposto (fonte de dados para calibrar o motor na etapa 2) e exige confirmação por escrito do Strava antes de gravar qualquer dado derivado deles.

## Regras que afetam o Cadência diretamente

### 1. Guardar dados: no máximo 7 dias

> "You may not retain Strava Data in your cache for longer than seven (7) days. [...] Except for such limited caching, you may not store Strava Data" (Policy 6.2)

> "You may not store Strava Data, or any data derived from Strava Data, in any Persistent Index. The foregoing prohibits indefinite storage in [...] archives, and any other storage configured to enable subsequent retrieval, query, or use." (Policy 5.5)

> "you may use and retain Data only so long as necessary for the purpose for which it was originally obtained." (Policy 6.4)

**Efeito:** atividades vindas do Strava não podem virar histórico permanente do Cadência (a tabela `imported_activities` guarda para sempre). A proteção graduada e a comparação planejado × realizado leem até 14, 28 e 42 dias de histórico; com dados do Strava, esse histórico não pode ser formado a partir deles.

**Ponto ambíguo (perguntar ao Strava):** se o atleta usa os números de uma atividade do Strava para **concluir o treino no Cadência** (o que grava `workout_sessions`), esse registro é "dado derivado" proibido de ficar guardado? A leitura literal de 5.5 sugere que sim.

### 2. IA: proibido em qualquer forma

> "You may not use the Strava API Materials or Strava Data, directly or indirectly, in connection with the development, training, evaluation, or operation of any AI Application." [...] inclusive "ingestion into a context window or working memory" e "Any data derived from, aggregated from, anonymized from, or generated using Strava Data" (Policy 5.3)

**Efeito:** nenhum dado do Strava, nem derivado, pode chegar à explicação por IA (`/v1/workouts/{id}/explanation`) nem ao resumo semanal com IA planejado. Hoje a IA recebe só a prescrição (nome, objetivo, duração, RPE-alvo, regras e escopo de evidência), mas uma das regras cita a quantidade de sessões concluídas; se essas sessões vierem do Strava, o texto passa a ser derivado deles. O texto lido não traz uma definição formal de "AI Application"; tratar de forma ampla.

### 3. Análise e melhoria do produto: proibido

> "You may not process or disclose Strava Data [...] including in an aggregated, de-identified, or anonymized manner, for the purposes of analytics, analyses, customer insight generation, or product or service improvements. You may not combine Strava Data with other customer data for these or any other purposes." (Policy 5.4)

**Efeito:** dados do Strava **não podem calibrar o motor, os shadows nem a etapa 2** da fase de dados reais, nem entrar em painéis internos. A segunda frase ("combine [...] for these or any other purposes") é ampla; perguntar ao Strava se comparar a atividade do próprio atleta com o treino dele, para mostrar só a ele, está permitido.

### 4. Exibição: só para o próprio atleta

> "Strava Data provided by a specific Strava user may be displayed or disclosed in your Developer Application only to that user." (Policy 2.3)

**Efeito:** compatível com o Cadência hoje (cada atleta só vê os próprios dados). Impede qualquer recurso futuro de treinador, grupo ou ranking com dados do Strava sem consentimento.

### 5. Consentimento, acesso e exclusão

- Antes de acessar os dados, informar: tipos de dados, como são coletados, como retirar o consentimento, como pedir exclusão e a confirmação de que a exclusão foi feita (Policy 2.1, 7.2). Mudou o tipo de dado coletado: pedir consentimento de novo.
- O atleta pode pedir acesso aos dados coletados (2.2).
- Excluir todos os dados do atleta quando ele pedir ou revogar o acesso, com **confirmação por escrito** ao atleta (2.5); prazo máximo de **30 dias** (7.4).
- Atividade apagada no Strava deve sumir do Cadência em até **48 horas** (6.3).
- Política de privacidade própria, compatível com o GDPR, com links visíveis, sem contradizer a do Strava, e com a declaração de que o Strava coleta dados de uso da API (7.3, 6.5).
- Contato de suporte fácil de achar e links para a conta do atleta no Strava (2.4).
- Termos de uso do Cadência devem isentar terceiros de garantias e responsabilidades (9.2).

### 6. Segurança e credenciais

- `client_secret` nunca compartilhado; um token de API por aplicação (Agreement 1.1, 1.2).
- Suspeita de vazamento do token: avisar `developers@strava.com` em até **24 horas** (Agreement 1.2). Incidente com dados: avisar `legal@strava.com` em até **24 horas** (Policy 8.3).
- Medidas de segurança adequadas (Policy 8.1). O Strava pode auditar a aplicação (Agreement 6.2).

### 7. Limites de acesso

- App novo nasce em modo de um único atleta; com assinatura, sobe para **10 atletas** direto no painel. Acima disso, o app vai para revisão, sem prazo garantido (Getting Started; Policy 3.3, 3.6).
- Limites padrão: leitura 200 requisições por 15 minutos e 2.000 por dia; geral 400 e 4.000. Excedeu: HTTP 429. Proibido contornar limites (Policy 3.7).
- Nível "Standard" exige que o desenvolvedor mantenha a assinatura ativa (Policy 3.3).

### 8. Outras proibições relevantes

- Não cobrar do atleta pelo que vem do Strava (5.8); cobrar por funções próprias do Cadência é permitido.
- Não usar dados do Strava em anúncios (5.9), não vender nem repassar a terceiros (5.10).
- Não criar app que imite o Strava ou concorra com ele (5.2).
- Não guardar localização geográfica (5.7); o Cadência já não guarda trajeto.
- Não expor a API a terceiros por proxy ou servidor MCP (5.16).
- Nada de press release citando o Strava sem autorização escrita (4.6).
- Se a atividade veio de um aparelho Garmin, exibir atribuição à Garmin (4.4).

## Marca

- Botão oficial **"Connect with Strava"** (laranja ou branco, 48 px de altura), linkando para `https://www.strava.com/oauth/authorize`.
- Para citar a integração, usar só **"Powered by Strava"** ou **"Compatible with Strava"**, com os logos oficiais, sem alteração, separados e menos destacados que a marca do Cadência.
- Link para a atividade original com o texto **"View on Strava"** (negrito, sublinhado ou laranja `#FC5200`).
- Nunca usar "Strava" no nome ou no ícone do app, nem sugerir que é oficial.

## Técnica

- **OAuth 2.0:** `https://www.strava.com/oauth/authorize`, parâmetro `state` para amarrar a volta à sessão do usuário, `redirect_uri` dentro do domínio cadastrado. O atleta pode desmarcar escopos; tratar o escopo concedido que volta na resposta.
- **Escopos mínimos:** `activity:read` (não `activity:read_all`, que inclui atividades "Só você" e zonas de privacidade).
- **Tokens:** o de acesso expira em 6 horas; a cada renovação vem um novo refresh token e o anterior deixa de valer imediatamente. Guardar sempre o mais recente, cifrado.
- **Revogação pelo app:** `POST https://www.strava.com/oauth/deauthorize` quando o atleta desconecta no Cadência.
- **Webhooks:** uma assinatura por aplicação; validação com `hub.challenge` e `verify_token`; responder **200 em até 2 segundos** e processar depois (o Strava tenta 3 vezes). Eventos: atividade criada, apagada ou alterada (título, tipo, privacidade) e revogação do app (`"authorized": "false"`). Com `activity:read`, a atividade que passa a "Só você" chega como exclusão.

## O que isso muda no plano

1. **O Strava deixa de ser a solução para a etapa 2.** Eu tinha recomendado o Strava como fonte de dados para calibrar o motor; as regras 5.3, 5.4 e 6.2 proíbem isso. A etapa 2 continua dependendo de dados que o próprio atleta traz ao Cadência (arquivo `.fit`/`.gpx` exportado do aparelho, que não passa pela API do Strava).
2. **O uso que parece permitido** é de conveniência: o atleta conecta a conta, vê as atividades recentes (até 7 dias) e usa uma delas para preencher o treino que ele confirma. Mesmo isso depende da resposta do Strava sobre o dado derivado (item 1).
3. **Antes de escrever código**, mandar as perguntas abaixo para `developers@strava.com`, como o próprio Agreement orienta ("If you are unsure if a certain use [...] is permitted", 2.2), e guardar a resposta.

## Perguntas para o Strava

**Status:** e-mail enviado pelo dono do produto a `developers@strava.com` em 30/09/2026, aguardando resposta (o Strava não garante prazo). Nenhum código da integração deve gravar dados do Strava antes da resposta; registrar aqui o texto recebido.

1. Quando o atleta usa uma atividade do Strava para concluir o treino planejado no nosso app, podemos guardar o registro resultante (duração, distância, FC, potência e cadência da sessão) como histórico de treino dele, visível só para ele, por mais de 7 dias? Ou esse registro é "data derived from Strava Data" sujeito à 5.5 e à 6.2?
2. Usar esse histórico, só do próprio atleta, para ajustar a prescrição dos próximos treinos dele (regras determinísticas, sem IA) é "operation of your Developer Application" ou é vedado pela 5.4?
3. Nosso app tem uma explicação de treino gerada por um modelo de linguagem a partir da prescrição (sem dados do Strava). Se uma regra da prescrição mencionar a quantidade de sessões concluídas, e algumas vierem do Strava, isso já viola a 5.3?

## Rascunho do e-mail (em inglês)

> Subject: Permitted use question before building a Strava integration (Standard tier, up to 10 athletes)
>
> Hello Strava Developers team,
>
> I am building Cadência, a cycling training planner. Each athlete sees only their own data. Before integrating the API, I want to confirm my use case complies with the API Agreement and API Policy (effective June 1, 2026):
>
> 1. An athlete connects Strava (scope `activity:read`) and picks one of their recent Strava activities to fill in the planned workout they are completing in our app. May we keep the resulting workout record (duration, distance, average heart rate, power and cadence) in that athlete's training history, visible only to them, beyond seven days? Or is it data derived from Strava Data under Sections 5.5 and 6.2?
> 2. May that athlete's own history be used by deterministic rules (no AI) to adjust their next workouts, or does Section 5.4 prohibit it?
> 3. Our app generates a plain-language workout explanation with a language model from the prescription only. If a prescription rule mentions how many sessions the athlete completed, and some came from Strava, does that fall under Section 5.3?
>
> We would delete all Strava data on request or deauthorization, reflect deletions within 48 hours, never store location data and never share data with third parties.
>
> Thank you.
