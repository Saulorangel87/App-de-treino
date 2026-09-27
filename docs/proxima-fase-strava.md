# Próxima fase: dados reais de execução (Strava) e evolução do produto

Criado em 27 de setembro de 2026. Este documento **define o escopo** da integração com o Strava, condição que o [`planejamento.md`](../planejamento.md) já exigia ("avaliar integrações externas, como Strava, somente após definir consentimento, custos e segurança"). Ele não é um segundo roadmap: a visão de produto continua no `planejamento.md` e o estado atual em [`STATUS.md`](STATUS.md). Não inclua segredos, tokens ou conteúdo de `.env` aqui.

## Por que esta fase vem primeiro

O motor `rules-v1` é deliberadamente conservador e `rules-v2` e os shadows só observam ([`roadmap-acceptance.md`](roadmap-acceptance.md)). O motivo registrado é a falta de dados reais de execução. Hoje o atleta digita o feedback de cada sessão; potência, frequência cardíaca, duração e elevação medidos são raros. Melhorar a "inteligência" sem esses dados seria calibrar no escuro. Importar as atividades executadas é o que mais aproxima o app de decisões melhores sem ampliar o risco de segurança.

Regra que continua valendo: **a integração não altera a prescrição.** Dados importados alimentam observação, comparação planejado × realizado e os shadows. Só ganham autoridade sobre o motor por decisão explícita, com comparação controlada (ver `roadmap-acceptance.md`, seção 2).

## Ordem proposta

1. **Importar atividades do Strava** (este documento, abaixo).
2. **Privacidade e LGPD:** exportar meus dados, excluir a conta com apagamento real (inclusive dados do Strava), política de retenção. Depende de a etapa 1 definir o que é armazenado, então o desenho é feito junto.
3. **Painel interno dos shadows:** comparar o que `rules-v1` prescreveu com o que `rules-v2` prescreveria, com os dados importados.
4. **Resumo semanal com IA, apenas explicativo:** usa a infraestrutura de IA existente e nunca altera a prescrição. Sujeito à decisão de termos de uso (ver "Restrições do Strava").
5. **Profissionalização operacional:** ambiente de staging, e2e no CI a cada PR, rastreamento de erros e métricas na API, onboarding e lembretes (PWA).

## Etapa 1: integração com o Strava

### Escopo (versão 1)

Dentro:
- Conectar a conta do Strava do atleta (OAuth 2.0) com consentimento explícito.
- Importar atividades de **ciclismo** (`Ride`, `VirtualRide`, `MountainBikeRide` e `GravelRide`; bicicleta elétrica fica de fora). Corrida e demais esportes são ignorados, pois estão fora do produto.
- Guardar por atividade: data/hora, duração em movimento, distância, elevação, FC média/máxima, potência média e normalizada (quando existir), cadência média, tipo.
- Casar a atividade importada com a sessão planejada do dia (mesma data e disciplina), sugerindo o vínculo ao atleta em vez de aplicá-lo às cegas.
- Desconectar a qualquer momento, com a opção de apagar os dados importados.

Fora da v1: enviar treinos ao Strava, ler segmentos e dados de outros atletas, guardar *streams* segundo a segundo (avaliar depois, pois custam armazenamento e têm mais sensibilidade), Garmin/Wahoo e importação de arquivos `.fit`/`.gpx` (fase seguinte).

### Fluxo

1. O atleta clica em "Conectar ao Strava" em `/configuracoes`. A API gera `state` aleatório (guardado na sessão, uso único) e redireciona ao Strava com os escopos mínimos: `read` e `activity:read` (atividades públicas e "só seguidores"; atividades privadas nunca são pedidas, decisão 2 abaixo).
2. O Strava redireciona ao callback da API. A API valida o `state`, troca o `code` por tokens e grava a conexão.
3. **Importação inicial** limitada aos últimos 90 dias, em segundo plano e com pausas para respeitar o limite de requisições.
4. **Atualização contínua** por webhook do Strava (assinatura de eventos do app, endpoint público validado por `hub.verify_token`), com verificação periódica como reserva. Os eventos de `deauthorize` do atleta apagam os tokens na hora.

### Modelo de dados (migração `000031`, aditiva)

- `external_connections`: `user_id`, `provider` (`strava`), `external_athlete_id`, `access_token` e `refresh_token` **cifrados**, `expires_at`, `scopes`, `connected_at`, `revoked_at`.
- `external_activities`: `user_id`, `provider`, `external_id` (único por provedor), `sport_type`, `started_at`, `moving_seconds`, `distance_m`, `elevation_gain_m`, `avg_hr`, `max_hr`, `avg_power`, `normalized_power`, `avg_cadence`, `payload_version`, `imported_at`, `workout_session_id` (nulo até o vínculo).
- Atualizar `database.RequiredMigrations`, adicionar fixture SQL e `down`, como nas migrações anteriores.

### Segurança e privacidade

- **Tokens cifrados em repouso** com AES-256-GCM. A chave fica só em variável de ambiente da VPS (`STRAVA_TOKEN_KEY`), fora do repositório e do banco, e entra no procedimento de backup de segredos. O `client_secret` do Strava também só existe no ambiente.
- Renovação do `access_token` sob demanda, com controle de concorrência (uma renovação por usuário por vez).
- Tokens e payloads brutos **nunca** aparecem em logs (o middleware de acesso já não registra corpos).
- O endpoint do webhook aceita apenas eventos da assinatura do próprio app e é idempotente (`external_id` único).
- Consentimento: tela explicando quais dados são lidos, para que servem, como revogar; registro de data do consentimento.
- Dados de atividades são **dados de saúde** para efeito da LGPD: entram na exportação e na exclusão de conta (etapa 2). Desconectar revoga o token no Strava (`/oauth/deauthorize`) e, se o atleta pedir, apaga `external_activities`.
- O backup diário já é criptografado antes de sair da VPS; os tokens cifrados vão junto, e a chave não.

### Restrições do Strava a confirmar antes de codificar

Estas são condições de terceiros que podem mudar o desenho. **Confirme no contrato e na documentação atuais do Strava antes de começar**, porque não foram verificadas nesta sessão:

- **Limite de atletas:** aplicativos novos costumam começar limitados a **um atleta** (o dono) até passarem por uma análise do Strava. Sem a aprovação, a integração serve para desenvolvimento e para você, mas não para os demais usuários.
- **Limites de requisição:** o Strava aplica limites por 15 minutos e por dia (na faixa de 100 leituras a cada 15 min e mil por dia por aplicativo, valores a confirmar). A importação inicial e o reprocesso precisam de fila com espera.
- **Uso dos dados em IA:** o Strava restringiu no contrato de API o uso de dados obtidos pela API para treinar ou alimentar modelos de IA/ML, e limitou a exibição a outras pessoas além do próprio atleta. Isto afeta diretamente as etapas 3 e 4: **antes** de usar dados importados em calibração de motor ou em resumos gerados por IA, é preciso confirmar o que o contrato permite. Se não permitir, os dados do Strava ficam restritos a mostrar ao próprio atleta e à comparação planejado × realizado por regras, e a importação por arquivo `.fit`/`.gpx` (dado do próprio atleta, sem esse contrato) passa a ser o caminho para alimentar a evolução do motor.
- **Marca:** o uso do nome, logotipo e do botão "Conectar com o Strava" segue as diretrizes de marca deles (o selo "Powered by Strava" costuma ser obrigatório).

### Testes

- Go: troca de código por tokens e renovação com servidor HTTP falso; casamento atividade × sessão (data, disciplina, tolerância de duração); idempotência do webhook; revogação; cifragem e decifragem dos tokens; recusa de `state` inválido ou reaproveitado.
- SQL: fixture da `000031` (unicidade de `external_id`, vínculo, exclusão em cascata ao apagar o usuário).
- e2e (Playwright) com um Strava simulado: conectar, importar, vincular e desconectar.
- Nenhum teste chama o Strava real. O teste manual em produção usa a conta do dono.

### Critérios de aceitação

- Atleta conecta, vê as atividades de ciclismo importadas e o vínculo sugerido com a sessão planejada.
- Desconectar remove o acesso; apagar dados remove as atividades importadas.
- Nenhum token ou dado bruto em logs; tokens ilegíveis no banco sem a chave.
- Prescrição de `rules-v1` **idêntica** com e sem dados importados (invariante coberta por teste).
- CI verde, migração `000031` aplicada pelo `deploy.sh` com backup e smoke test.

### Decisões (registradas em 27/09/2026, com o dono do produto)

1. **Só o dono no início.** Não solicitar a análise do Strava agora; a v1 vale só para a conta do dono (`activity:read`, sem liberação para outros atletas). Se funcionar bem, avaliar a análise depois para liberar aos demais usuários.
2. **Sem atividades privadas.** Nunca pedir `activity:read_all`; a integração lê apenas atividades públicas e "só seguidores".
3. **Sem *streams*.** Confirmado que a v1 guarda só os totais por atividade (seção "Modelo de dados"), não a série segundo a segundo. Ver explicação abaixo.
4. **Sim, `.fit`/`.gpx` como alternativa.** Se o contrato do Strava vetar o uso desses dados em calibração de motor ou resumo por IA, a importação de arquivo próprio do atleta (`.fit`/`.gpx`) passa a ser o caminho para alimentar essas etapas, sem depender desse contrato.

#### O que são streams (para referência)

Cada atividade do Strava tem dois níveis: o **resumo** (um valor por atividade — duração, distância, FC média, potência média) e os **streams**, a série de valores a **cada segundo** da atividade inteira (milhares de pontos por hora pedalada, incluindo o trajeto GPS exato). A v1 usa só o resumo. Streams custariam muito mais armazenamento e são mais sensíveis, porque revelam a rota exata (e, por tabela, onde o atleta mora ou treina). Ficam fora até haver um motivo concreto, como gráficos de sessão por segundo.

## Versão

A entrega visível ao usuário incrementa `APP_VERSION` em `frontend/lib/release.ts` (próxima versão menor, `0.34.0`) e descreve a novidade em `UPDATE_NOTES`, conforme o [`README.md`](README.md) das docs.
