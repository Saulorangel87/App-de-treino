# Estado atual do Cadência

Última atualização: 30 de setembro de 2026.

Este é o documento de continuidade: curto e sempre atual. O diário cronológico das fatias de trabalho (com datas, validações e decisões) está em [`changelog/project-status-history.md`](changelog/project-status-history.md). Não inclua senhas, tokens, chaves de API nem conteúdo de `.env`.

## Produção (VPS Oracle)

| Item | Valor |
| --- | --- |
| Frontend | <https://cadencia.devsaulo.com.br> |
| API | <https://cadencia-api.devsaulo.com.br> |
| Versão publicada | `0.36.0`, commit `d9e3f4f` (deploy de 30/09/2026 UTC: vínculo de atividade importada; antes, em 29/09/2026, a identidade visual "carta topográfica" e a correção da rolagem do aviso de novidades) |
| Migrações aplicadas | `000001` a `000031` (a `000032` está no código, ainda não aplicada em produção) |
| Motor prescritivo | `rules-v1` (único); `rules-v2` e demais shadows são somente observacionais |
| Último backup preventivo | `cadencia-20260930T002403Z.dump` |
| Validação pós-deploy | login, `GET /v1/plans/current` e `logout-others` = 200 com a conta de smoke test; `/ready` verifica o schema |

Escopo: somente ciclismo (estrada, MTB XCO/XCM, gravel e indoor). Corrida e musculação são produtos separados.

## Dados reais de execução — publicado na `0.34.0`

Primeira fatia da fase descrita em [`proxima-fase-dados-reais.md`](proxima-fase-dados-reais.md): importar atividades por arquivo `.fit`/`.gpx`, sem depender do Strava (adiado por exigir assinatura paga da API; desenho preservado no mesmo documento). Publicada em 27/09/2026 (PRs #31–#34).

- **Migração `000031_imported_activities`:** registro bruto por atividade (duração, distância, elevação, FC, potência, cadência), deduplicado por hash do arquivo, com vínculo opcional a um treino planejado (`SET NULL` ao apagar o treino, `CASCADE` ao apagar o usuário).
- **`POST /v1/activities/import`** (multipart, 20 MB), **`GET /v1/activities/imported`**, **`DELETE .../{id}`**. A importação nunca escreve em `workout_sessions` nem chama o pacote `planning`; só quando o atleta confirma pelos endpoints de conclusão/correção que já existiam é que os dados entram no histórico.
- **`/atividades/importar`:** upload, resumo extraído, sugestão de treino do mesmo dia. O link do treino sugerido abre `/plano` com o formulário de conclusão (ou correção) já preenchido, para o atleta revisar e confirmar.
- **Atalho "Compartilhar" no Android:** `share_target` no `app.webmanifest` + service worker; o app do relógio/ciclocomputador pode compartilhar o arquivo direto para o Cadência (exige o PWA instalado na tela inicial). Ainda não testado num Android real.
- **Correção de biblioteca (PR #32):** um arquivo `.fit` real de um ciclocomputador XOSS expôs um bug no decodificador inicialmente usado (`tormoder/fit`, que rejeita arquivos com várias mensagens de definição em sequência antes dos dados — válido no protocolo, mas fora do que a biblioteca aceitava). Trocado por `github.com/muktihari/fit`, mais ativa. Também corrigido: sem sensor pareado, esse aparelho grava `0` em vez do valor "inválido" do FIT nos campos de FC/potência/cadência da sessão; tratado como ausente.
- **Correção de fuso horário (PR #34):** o arquivo grava o horário em UTC; sem correção, uma pedalada perto da meia-noite local podia cair no dia UTC errado e não sugerir o treino do dia certo. Usa `LocalTimestamp` do `.fit` quando presente; sem essa informação (sempre o caso do `.gpx`), a busca por treino candidato olha também o dia anterior e o seguinte.

## Identidade visual "carta topográfica" — publicada na `0.35.0`

Publicada em 29/09/2026 (PRs #39 e #40; deploy sem migração, `0 migração(ões) pendente(s)`). Troca completa do visual do frontend, sem mudança de API, banco ou regras do motor. Detalhes em [`changelog/project-status-history.md`](changelog/project-status-history.md).

- Tokens em `frontend/app/globals.css`; estilos por área em `frontend/app/styles/`. `contrast-fixes.css` removido.
- Navegação única (`components/app-header.tsx`): cabeçalho no desktop, barra inferior no celular.
- Novos ícones do PWA, `og.png`, tela offline e cache do service worker `v4`.
- Validação: `npm run e2e` (2 testes) e uma verificação da navegação nos dois tamanhos passaram contra a API local antes do deploy; o smoke test público e autenticado passou em produção. Ainda não testado num aparelho real: instalação do PWA com os ícones novos (fechar e abrir o app, ou reinstalar, para trocar o ícone) e o atalho de compartilhar do Android.

## Vínculo de atividade importada a um treino — publicado na `0.36.0`

Corrige uma falha da `0.34.0`: a importação só *sugeria* o treino do mesmo dia e nenhum código gravava o vínculo, então a lista sempre mostrava "Sem treino vinculado". Agora o atleta vincula pelo botão, na sugestão logo após importar ou depois, na lista.

- **`PUT /v1/activities/imported/{id}/workout`** (`workout_id` ou `null` para desvincular) e **`GET /v1/activities/imported/{id}/candidates`** (dia da atividade e vizinhos). O treino precisa ser de um plano do próprio atleta (`active`, `draft` ou `completed`).
- Só grava `imported_activities.workout_id`. Nada é copiado para `workout_sessions`, o plano não muda e o pacote `planning` não é chamado; a execução registrada continua vindo de `/complete` e `/correct`.
- A listagem devolve `workout_name` e `workout_scheduled_on` do treino vinculado. Sem migração.
- Testes: serviço e handlers, mais a fixture `database/tests/imported_activity_link.sql` (dono da atividade e do treino, plano cancelado, desvínculo). Validado também com a API e o banco locais (18 verificações).

## Proteção graduada do motor — código na `master`, desligada por padrão

Substitui a trava de 28 dias (qualquer dor ou fadiga média alta protegia o ciclo inteiro) por sinais com data, níveis e reavaliação. Plano, critérios e decisões em [`motor-protecao-cenarios.md`](motor-protecao-cenarios.md). Etapas 1 a 4 de 5 (PRs #46 a #49 e o da etapa 4); a etapa 5 é esta documentação e a revisão das referências científicas.

- **Chave `PROTECTION_LEVELS_ENABLED` (padrão `false`).** Desligada, o app segue a regra antiga e nenhum treino é reavaliado. Para ligar ou desligar, altere a variável no ambiente da API e recrie o contêiner da API.
- **Níveis:** nenhuma, leve (−10% e RPE −1 nos de qualidade), moderada (qualidade vira giro protegido, demais −10%) e forte (todos viram giro protegido). Dor isolada: forte por 3 dias, moderada até o 7º; dor recorrente (2+ em 14 dias): forte por 7 dias com sugestão de avaliação profissional; check-in bom rebaixa um nível; dor sempre conta, mesmo em sessão sem dados mínimos.
- **Reavaliação:** ao concluir ou corrigir um treino, salvar o check-in ou tocar em "Estou recuperado", os treinos `planned` do plano ativo (hoje até o domingo da semana seguinte) são refeitos. Nunca toca concluídos, iniciados, cancelados, `adapted`, rascunho, os 1 a 2 treinos que o gatilho do banco já adaptou, nem treinos gerados antes de `explanation.prescription_inputs` existir (o plano atual do atleta só muda ao gerar um plano novo).
- **Migração `000032_recovery_self_reports`** (declaração "Estou recuperado", única por dia, em cascata com o perfil) e **`POST /v1/protection/recovered`** (404 com a chave desligada).
- **Tela do plano:** aviso com o motivo, a data em que a proteção termina e o botão "Estou recuperado"; marca "Proteção leve/moderada/forte" nos treinos.
- **Validação:** testes unitários, fixtures `protection_reevaluation.sql` e `account_deletion.sql`, fluxo real com API e PostgreSQL (21 verificações) e verificação visual no celular e no desktop.
- **Antes de ligar em produção:** aplicar a `000032` (o `deploy.sh` faz, com backup), ligar a chave e gerar um plano novo para o atleta.

## Segurança do repositório (29/09/2026)

- Dependabot: 0 alertas. Corrigidos `fast-uri` 3.1.8 (alto, PR #37) e `undici` 7.29.1 (alto, PR #38), este via `wrangler` 4.143.1 e `@cloudflare/vite-plugin` 1.62.1. O CI falhava em todo PR do frontend pelo `npm audit` do `fast-uri`.
- Code scanning (CodeQL) ativado em Go, JavaScript/TypeScript e Actions. O primeiro alerta (`go/incorrect-integer-conversion`, sem risco real) foi corrigido no PR #40. Secret scanning e proteção de push já estavam ativos, com 0 alertas.
- `master` protegida: 5 jobs obrigatórios do CI (`backend`, `database`, `frontend`, `openapi`, `docker`).

## Pendências operacionais (fora do código)

- Testar o atalho de compartilhar no Android num aparelho real.
- Etapa 2 da fase de dados reais: melhorias no motor de treino usando os dados importados (ainda não desenhada em detalhe; ver [`proxima-fase-dados-reais.md`](proxima-fase-dados-reais.md)).
- Depois: LGPD (exportar/apagar dados), painel interno dos shadows, resumo semanal com IA.
- Hardening da VPS e limpeza gradual do que restar de dívida técnica.
- Coleta longitudinal de dados reais antes de dar autoridade adicional aos shadows.
- Decisão de produto: a hospedagem do frontend usa `vinext` (beta) com dependências herdadas do ambiente de criação (`wrangler`, `@openai/sites-vite-plugin`); avaliar migração para uma base mais estável.
