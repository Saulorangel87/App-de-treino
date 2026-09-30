# Próxima fase: dados reais de execução e evolução do motor

Criado em 27 de setembro de 2026, revisado no mesmo dia após descobrir que a API do Strava passou a exigir assinatura paga do desenvolvedor (`"A Strava subscription is a prerequisite for creating an app"`, developers.strava.com/docs/getting-started). Este documento **define o escopo** da fase de dados reais de execução, condição que o [`planejamento.md`](../planejamento.md) já exigia para qualquer integração externa ("avaliar somente após definir consentimento, custos e segurança"). Não é um segundo roadmap: a visão de produto continua no `planejamento.md` e o estado atual em [`STATUS.md`](STATUS.md). Não inclua segredos, tokens ou conteúdo de `.env` aqui.

## Por que esta fase vem primeiro

O motor `rules-v1` é deliberadamente conservador e `rules-v2` e os shadows só observam ([`roadmap-acceptance.md`](roadmap-acceptance.md)). O motivo registrado é a falta de dados reais de execução. Hoje o atleta digita o feedback de cada sessão; potência, frequência cardíaca, duração e elevação medidos são raros. Melhorar a "inteligência" sem esses dados seria calibrar no escuro. Importar as atividades executadas é o que mais aproxima o app de decisões melhores sem ampliar o risco de segurança.

Regra que continua valendo: **a importação não altera a prescrição.** Dados importados alimentam observação, comparação planejado × realizado e os shadows. Só ganham autoridade sobre o motor por decisão explícita, com comparação controlada (ver `roadmap-acceptance.md`, seção 2).

## Decisão sobre o Strava (27/09/2026)

A integração direta por OAuth com o Strava (desenhada na primeira versão deste documento) foi **adiada**, não descartada. O motivo: a Strava API agora exige assinatura paga do desenvolvedor para gerar credenciais (Client ID/Secret), custo recorrente que não se justifica enquanto o app está validando o produto com poucos usuários. O dono do produto decidiu revisitar essa integração **quando o app estiver mais maduro**. O desenho técnico completo (fluxo OAuth, modelo de dados, segurança, restrições de terceiros) fica preservado no anexo "Integração com o Strava (desenho preservado)", no fim deste arquivo, para não perder o trabalho já feito.

Em vez disso, a etapa 1 é a **importação de arquivo `.fit`/`.gpx`**, abaixo. Ela cobre qualquer relógio ou ciclocomputador (Garmin, Wahoo, Polar, e o XOSS que o dono e os amigos dele usam), sem custo de assinatura e sem depender de aprovação de terceiro.

## Ordem proposta

1. **Importar atividades por arquivo `.fit`/`.gpx`**, com atalho de compartilhamento no Android (este documento, abaixo).
2. **Melhorias no motor de treino**, usando os dados importados na etapa 1: ver "Etapa 2" abaixo.
3. **Privacidade e LGPD:** exportar meus dados, excluir a conta com apagamento real (inclusive atividades importadas), política de retenção.
4. **Painel interno dos shadows:** comparar o que `rules-v1` prescreveu com o que `rules-v2` prescreveria, com os dados importados.
5. **Resumo semanal com IA, apenas explicativo:** usa a infraestrutura de IA existente e nunca altera a prescrição.
6. **Profissionalização operacional:** ambiente de staging, e2e no CI a cada PR, rastreamento de erros e métricas na API, onboarding e lembretes (PWA).
7. **Reavaliar o Strava**, quando fizer sentido: reabrir o anexo preservado, confirmar preço e termos atuais, e decidir.

## Etapa 1: importação de atividades por arquivo (`.fit`/`.gpx`)

**Status (27/09/2026):** mesclada no master (PRs #31, #32, #33), com o CI verde (migração/fixture rodaram num Postgres real). Falta testar o atalho de compartilhar num Android real.

- **PR #31:** implementação inicial (parser, migração `000031`, endpoints, `/atividades/importar`, atalho de compartilhar no Android).
- **PR #32 (correção):** testado com um arquivo `.fit` real de um XOSS (o aparelho que motivou esta etapa) e a biblioteca `tormoder/fit` rejeitava o arquivo (ela só aceita a mensagem `file_id` sozinha, sem outra definição logo depois — o XOSS grava várias em sequência, válido no protocolo). Trocado para `github.com/muktihari/fit`, que decodifica o mesmo arquivo sem erro e é mais ativamente mantida. Também corrigido: sem sensor de FC/cadência pareado, esse aparelho grava `0` em vez do valor "inválido" do FIT; uma média de sessão inteira igual a 0 passou a ser tratada como ausente.
- **Fuso horário (mesmo dia do PR #32):** o FIT/GPX grava o horário em UTC; sem correção, uma pedalada perto da meia-noite local pode "virar o dia" em UTC e deixar de bater com a data do treino planejado. Corrigido usando o campo `LocalTimestamp` da mensagem `activity` do FIT quando presente (a maioria tem); como rede de segurança para quando falta (sempre o caso do GPX, que não tem esse campo), a sugestão de vínculo passou a buscar também o dia anterior e o seguinte.
- **PR #33:** o link do treino sugerido em `/atividades/importar` agora abre `/plano` com o formulário de conclusão (ou correção) já preenchido com os dados extraídos; o atleta ainda confirma manualmente. Sem isso, a atividade importada não entrava no histórico (`/atividades`), que só é alimentado por `workout_sessions`.

### Escopo (versão 1)

Dentro:
- Upload manual de um arquivo `.fit` ou `.gpx` por vez, em `/atividades` (nome de rota a confirmar).
- **Atalho Android:** o Cadência se registra como alvo do menu "Compartilhar" do sistema (Web Share Target API do PWA). No app do relógio/ciclocomputador (XOSS, Garmin Connect, etc.), o atleta toca em Compartilhar → Cadência, sem abrir o navegador manualmente nem procurar o arquivo. iOS fica de fora nesta etapa (suporte de Share Target mais limitado e inconsistente entre versões); nesse caso o fluxo é upload manual mesmo.
- Extrair da atividade: data/hora, duração em movimento, distância, elevação acumulada, FC média/máxima, potência média e normalizada (quando existir), cadência média, tipo de esporte (aceitar só ciclismo; arquivo de outro esporte é rejeitado com mensagem clara).
- Casar a atividade importada com a sessão planejada do dia (mesma data e disciplina), sugerindo o vínculo ao atleta em vez de aplicá-lo às cegas. **Desde a `0.36.0`** o atleta grava o vínculo com o botão "Vincular" (`PUT /v1/activities/imported/{id}/workout`); até a `0.35.0` só a sugestão existia e nada persistia o vínculo.
- Permitir apagar uma atividade importada.

Fora da v1: streams segundo a segundo (ver explicação abaixo — mesma decisão tomada para o Strava, vale aqui também), importação em lote de muitos arquivos de uma vez (avaliar na v2 se o upload único gerar atrito para quem quer importar o histórico), OAuth com qualquer provedor externo.

#### O que são streams (para referência)

Cada atividade tem dois níveis: o **resumo** (um valor por atividade — duração, distância, FC média, potência média), que é o que a v1 guarda, e os **streams**, a série de valores a **cada segundo** da atividade inteira (milhares de pontos por hora pedalada, incluindo o trajeto GPS exato). Streams custam muito mais armazenamento e são mais sensíveis, porque revelam a rota exata. Ficam fora até haver um motivo concreto, como gráficos de sessão por segundo.

### Formato dos arquivos

- **`.fit`**: formato binário da Garmin, hoje um padrão de fato (Garmin, Wahoo, Polar, Zwift, exportação do próprio Strava). Mais rico em dados (inclui potência e cadência quando o equipamento mede).
- **`.gpx`**: formato em texto (XML), mais antigo, focado em rota e tempo; nem sempre traz potência ou cadência.
- Parser em Go: usar uma biblioteca existente para `.fit` (ex.: `tormoder/fit`) em vez de escrever o parser binário do zero; `.gpx` é XML simples e dá para decodificar com `encoding/xml` da biblioteca padrão.
- Tamanho máximo de upload a definir (arquivos `.fit` de pedaladas longas costumam ficar na casa de poucos MB; propor limite de 20 MB por arquivo, generoso o suficiente).

### Web Share Target (atalho Android)

- Declarar `share_target` no `manifest.json` do PWA, apontando para um endpoint que recebe o arquivo compartilhado (`POST` com `multipart/form-data`).
- O navegador precisa reconhecer o Cadência como PWA instalado (o atleta precisa ter adicionado à tela inicial) para o atalho aparecer no menu Compartilhar do Android; documentar esse passo na tela de importação.
- Testar no Chrome Android real (o Share Target não é simulável de forma confiável em CI); registrar o resultado manual no `STATUS.md` quando testado.
- iOS: sem Share Target neste momento. A tela de importação mostra só o upload manual para quem acessa de iPhone.

### Modelo de dados (migração `000031`, implementada)

- `imported_activities`: `user_id`, `workout_id` (nulo; aponta para `workouts`, não para `workout_sessions` — a execução ainda pode não existir quando o arquivo é importado), `source` (`fit`/`gpx`), `file_hash` (SHA-256, único por `user_id`), `started_at`, `moving_seconds`, `distance_km`, `elevation_gain_m`, `average_heart_rate`, `max_heart_rate`, `average_power_watts`, `normalized_power_watts`, `average_cadence_rpm`, `imported_at`. `ON DELETE SET NULL` no treino, `ON DELETE CASCADE` no usuário.
- **Decisão tomada na implementação:** a importação não duplica o armazenamento de execução. `workout_sessions` já guarda os mesmos campos (duração, distância, potência, FC, cadência) quando o atleta conclui ou corrige um treino pelos endpoints que já existiam (`POST /v1/workouts/{id}/complete` e `/correct`). A tela de importação mostra o resumo extraído e sugere o treino do mesmo dia; o atleta decide se usa esses números ao concluir/corrigir esse treino em `/plano`. `imported_activities` não é escrita nem lida pelo pacote `planning`.
- `database.RequiredMigrations` atualizado; fixture SQL cobre unicidade por usuário, `SET NULL` ao apagar o treino e `CASCADE` ao apagar o usuário.

### Segurança e privacidade

- Validar o arquivo antes de processar: tamanho máximo, assinatura binária correta (`.fit` tem um cabeçalho identificável), tempo limite de parsing (arquivo corrompido não pode travar a requisição).
- O arquivo bruto **não é guardado em disco** depois de extrair os campos acima; só os campos extraídos vão para o banco. Reduz a superfície de dados sensíveis (o `.fit`/`.gpx` cru inclui o trajeto GPS completo).
- Atividades importadas são **dados de saúde** para efeito da LGPD: entram na exportação e na exclusão de conta (etapa 3). Apagar a conta apaga `imported_activities` em cascata.
- Sem tokens nem segredos de terceiros nesta etapa (diferença direta em relação ao desenho do Strava): não há chave de API para proteger, porque não há OAuth.

### Testes

- Go: parser de `.fit` e `.gpx` com arquivos de exemplo (válido, corrompido, de outro esporte, maior que o limite); casamento atividade × sessão (data, disciplina, tolerância de duração); rejeição de arquivo duplicado (`file_hash`); exclusão de atividade importada.
- SQL: fixture da `000031` (unicidade de `file_hash` por usuário, vínculo, exclusão em cascata ao apagar o usuário).
- e2e (Playwright): upload manual de um `.fit` de exemplo, vínculo sugerido, exclusão. O Share Target não entra no e2e automatizado (ver acima); teste manual documentado à parte.
- Nenhum teste depende de rede externa.

### Critérios de aceitação

- Atleta sobe um `.fit` ou `.gpx`, vê a atividade extraída e o vínculo sugerido com a sessão planejada.
- No Android, com o Cadência instalado como PWA, "Compartilhar" a partir do app do relógio abre o Cadência já com o arquivo pronto para confirmar.
- Apagar uma atividade importada remove o registro.
- Arquivo de outro esporte, corrompido ou maior que o limite é rejeitado com mensagem clara, sem erro 500.
- Prescrição de `rules-v1` **idêntica** com e sem dados importados (invariante coberta por teste).
- CI verde, migração `000031` aplicada pelo `deploy.sh` com backup e smoke test.

## Etapa 2: melhorias no motor de treino (depois da etapa 1)

**Primeira parte feita (30/09/2026, na `master`, não publicada; deploy combinado para junto da integração com o Strava):** quando uma atividade importada está vinculada a um treino, `GET /v1/plans/current` devolve em cada treino o bloco `imported_execution` (`imported-execution-v1`, modo observação): duração planejada × minutos em movimento medidos pelo aparelho, a duração registrada no app × a medida, e FC, potência, cadência, distância e elevação do arquivo, com `missing_data` explícito. É calculado na leitura (sem migração, nada gravado), só com atividades do próprio atleta e a importada por último quando há mais de uma. **Não altera a prescrição:** `buildPlan`, a proteção e as adaptações não leem esse bloco (`used_for_prescription: false`). Isso muda uma regra da etapa 1 só na leitura: o repositório passa a ler `imported_activities` para montar o plano, mas o pacote `planning` só recebe os fatos já extraídos e não os usa para decidir. Validação: testes unitários, fixture `imported_execution.sql` e fluxo real (importar, vincular, ler o plano, desvincular).

**Por que o restante ainda não:** em 30/09/2026 a produção tinha 0 atividades importadas e 3 treinos concluídos nos últimos 42 dias. Calibrar shadows com isso seria decidir sem dados. O restante continua dependendo de algumas semanas de uso real.

Restante ainda não desenhado em detalhe; entra em um próximo ciclo, com dados reais de pelo menos algumas semanas de uso da etapa 1 disponíveis. Direção pretendida, a refinar quando chegar a vez:
- Comparar sistematicamente planejado × realizado (duração, RPE percebido vs. esforço medido) usando `imported_activities`, sem alterar prescrição.
- Usar essa comparação para calibrar os shadows (`rules-v2`, tolerância de carga, etc.) com dados de execução reais, não só o feedback subjetivo já coletado hoje.
- Qualquer ganho de autoridade de um shadow sobre a prescrição exige a comparação controlada e a revisão explícita já exigidas em `roadmap-acceptance.md` (seção 2); esta etapa não muda esse critério, só melhora a base de dados disponível para a decisão.

## Anexo: integração com o Strava (desenho preservado, adiado)

Mantido como referência para quando a integração for revisitada. Nada aqui está em desenvolvimento agora.

### Por que foi adiada

A criação de um app na Strava API exige assinatura paga do desenvolvedor (confirmado em developers.strava.com/docs/getting-started, 27/09/2026: *"A Strava subscription is a prerequisite for creating an app"*). Antes de retomar, confirmar o preço atual e se ele mudou.

### Limites confirmados na documentação oficial (27/09/2026)

- **Single-player mode:** todo app novo nasce liberado só para a conta do próprio desenvolvedor. É possível expandir para **até 10 atletas conectados sem aprovação formal**. Além de 10, é preciso enviar o app para revisão do Strava (formulário próprio, critérios não detalhados publicamente).
- **Limites de requisição, padrão (antes de revisão):** 100 requisições/15 min e 1.000/dia. Após revisão aprovada: até 400/15 min e 4.000/dia.
- **Uso de dados em IA/ML:** não encontrada uma cláusula específica na documentação técnica pública consultada; **confirmar nos termos de serviço/contrato de API atuais antes de usar dados do Strava em calibração de motor ou em texto gerado por IA**, e não assumir que é permitido.
- **Marca:** uso do nome, logotipo e do botão "Conectar com o Strava" segue as diretrizes de marca deles (o selo "Powered by Strava" costuma ser obrigatório).

### Escopo que estava desenhado para a v1

- Conectar a conta do Strava do atleta (OAuth 2.0) com consentimento explícito; escopos mínimos `read` e `activity:read` (nunca `activity:read_all` — decisão de não incluir atividades privadas).
- Importar atividades de ciclismo (`Ride`, `VirtualRide`, `MountainBikeRide`, `GravelRide`), sem streams (mesma decisão da etapa 1 de agora).
- Importação inicial dos últimos 90 dias; atualização contínua por webhook.
- Modelo de dados equivalente ao de `imported_activities`, com tabela adicional `external_connections` para os tokens.
- Tokens de acesso e renovação cifrados em repouso (AES-256-GCM), chave só em variável de ambiente da VPS, nunca em log; revogação e exclusão de dados ao desconectar.
- Testes: troca de código por tokens, renovação, idempotência de webhook, cifragem/decifragem, casamento atividade × sessão — sem chamar o Strava real.

### Antes de retomar

1. Confirmar preço e termos atuais da assinatura de desenvolvedor.
2. Confirmar a cláusula de uso de dados em IA/ML nos termos vigentes na época.
3. Decidir se a v1 continua limitada a "só o dono" (dentro do single-player mode) ou já vale abrir para até 10 atletas amigos, já que esse limite existe sem aprovação.
4. Se muitos atletas já tiverem `imported_activities` da etapa 1 funcionando bem, avaliar se o Strava ainda agrega o suficiente para justificar o custo.

## Versão

A entrega visível ao usuário incrementa `APP_VERSION` em `frontend/lib/release.ts` (próxima versão menor, `0.34.0`) e descreve a novidade em `UPDATE_NOTES`, conforme o [`README.md`](README.md) das docs.
