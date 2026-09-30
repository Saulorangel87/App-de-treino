# Estado atual do Cadência

Última atualização: 30 de setembro de 2026.

Este é o documento de continuidade: curto e sempre atual. O diário cronológico das fatias de trabalho (com datas, validações e decisões) está em [`changelog/project-status-history.md`](changelog/project-status-history.md). Não inclua senhas, tokens, chaves de API nem conteúdo de `.env`.

## Produção (VPS Oracle)

| Item | Valor |
| --- | --- |
| Frontend | <https://cadencia.devsaulo.com.br> |
| API | <https://cadencia-api.devsaulo.com.br> |
| Versão publicada | `0.37.0`, commit `0ba6efa` (deploys de 30/09/2026: `8786881`, proteção graduada do motor e "Estou recuperado"; `0ba6efa`, correção das referências científicas sem mudar a versão; o vínculo de atividade importada, `0.36.0`, foi publicado antes, no mesmo dia em UTC) |
| Migrações aplicadas | `000001` a `000033` |
| Motor prescritivo | `rules-v1` (único); `rules-v2` e demais shadows são somente observacionais. Proteção graduada ligada (`PROTECTION_LEVELS_ENABLED=true` no `.env.production` da VPS) |
| Último backup preventivo | `cadencia-20260930T222018Z.dump` |
| Validação pós-deploy | `/health`, `/ready`, frontend e `GET /v1/plans/current` autenticado = 200 com a conta de smoke test; `/ready` verifica o schema; `recovery_self_reports` criada; chave `true` confirmada dentro do contêiner da API; `POST /v1/protection/recovered` sem sessão = 401 |

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

## Proteção graduada do motor — publicada na `0.37.0`, ligada em produção

Substitui a trava de 28 dias (qualquer dor ou fadiga média alta protegia o ciclo inteiro) por sinais com data, níveis e reavaliação. Plano, critérios e decisões em [`motor-protecao-cenarios.md`](motor-protecao-cenarios.md). Etapas 1 a 4 de 5 (PRs #46 a #50), publicadas em 30/09/2026 pelo `deploy.sh` (fast-forward `d9e3f4f` → `8786881`, backup `cadencia-20260930T110311Z.dump`, 1 migração aplicada). Release [v0.37.0](https://github.com/Saulorangel87/App-de-treino/releases/tag/v0.37.0). A etapa 5 (auditoria das referências científicas) foi publicada no mesmo dia pelo `deploy.sh` (fast-forward `8786881` → `0ba6efa`, backup `cadencia-20260930T222018Z.dump`, migração `000033`), sem release novo.

- **Chave `PROTECTION_LEVELS_ENABLED` (padrão `false`, `true` em produção desde o deploy).** Desligada, o app segue a regra antiga e nenhum treino é reavaliado. Para desligar, mude a linha no `.env.production` da VPS para `false` e recrie o contêiner da API (`docker compose ... up -d api`); não precisa de novo deploy de código.
- **Níveis:** nenhuma, leve (−10% e RPE −1 nos de qualidade), moderada (qualidade vira giro protegido, demais −10%) e forte (todos viram giro protegido). Dor isolada: forte por 3 dias, moderada até o 7º; dor recorrente (2+ em 14 dias): forte por 7 dias com sugestão de avaliação profissional; check-in bom rebaixa um nível; dor sempre conta, mesmo em sessão sem dados mínimos.
- **Reavaliação:** ao concluir ou corrigir um treino, salvar o check-in ou tocar em "Estou recuperado", os treinos `planned` do plano ativo (hoje até o domingo da semana seguinte) são refeitos. Nunca toca concluídos, iniciados, cancelados, `adapted`, rascunho, os 1 a 2 treinos que o gatilho do banco já adaptou, nem treinos gerados antes de `explanation.prescription_inputs` existir (o plano atual do atleta só muda ao gerar um plano novo).
- **Migração `000032_recovery_self_reports`** (declaração "Estou recuperado", única por dia, em cascata com o perfil) e **`POST /v1/protection/recovered`** (404 com a chave desligada).
- **Tela do plano:** aviso com o motivo, a data em que a proteção termina e o botão "Estou recuperado"; marca "Proteção leve/moderada/forte" nos treinos.
- **Validação:** testes unitários, fixtures `protection_reevaluation.sql` e `account_deletion.sql`, fluxo real com API e PostgreSQL (21 verificações) e verificação visual no celular e no desktop.
- **Plano existente:** os planos gerados antes da `0.37.0` não têm `prescription_inputs` e não são reavaliados. A proteção só passa a se adaptar depois que o atleta gera um plano novo em Plano > Atualizar plano.
- **Validado com o dono do produto** num ambiente local com seis contas de demonstração (nenhuma, leve, moderada, forte, dor recorrente, dor antiga), já apagadas.

## Segurança do repositório (29/09/2026)

- Dependabot: 0 alertas. Corrigidos `fast-uri` 3.1.8 (alto, PR #37) e `undici` 7.29.1 (alto, PR #38), este via `wrangler` 4.143.1 e `@cloudflare/vite-plugin` 1.62.1. O CI falhava em todo PR do frontend pelo `npm audit` do `fast-uri`.
- Code scanning (CodeQL) ativado em Go, JavaScript/TypeScript e Actions. O primeiro alerta (`go/incorrect-integer-conversion`, sem risco real) foi corrigido no PR #40. Secret scanning e proteção de push já estavam ativos, com 0 alertas.
- `master` protegida: 5 jobs obrigatórios do CI (`backend`, `database`, `frontend`, `openapi`, `docker`).

## Pendências operacionais (fora do código)

- **Na `master`, não publicado (deploy combinado para junto da integração com o Strava):** migração `000034` (auditoria das 27 referências contra o PubMed; troca da fonte do "Limiar controlado") a primeira parte da etapa 2 da fase de dados reais (bloco observacional `imported_execution` no plano; ver [`proxima-fase-dados-reais.md`](proxima-fase-dados-reais.md)) e a LGPD: exportação dos dados (`GET /v1/auth/account/export`, botão em Configurações), Política de Privacidade (`/privacidade`) e Termos de Uso (`/termos`). Os textos estão em `frontend/lib/legal.ts`; o dono do produto precisa revisá-los antes do deploy (prazos de resposta e de aviso são compromissos dele).

- Plano novo gerado pelo dono do produto em produção: a proteção graduada passou a valer nele (30/09/2026). Ainda não foi exercitado em produção o botão "Estou recuperado".
- Atalho de compartilhar no Android: o app XOSS não mostra o Cadência na lista; manifesto e Chrome auditados sem erro em 30/09/2026. Decisão do dono do produto: deixar como está (Atividades → Importar funciona).
- Etapa 2 da fase de dados reais: melhorias no motor de treino usando os dados importados (ainda não desenhada em detalhe; ver [`proxima-fase-dados-reais.md`](proxima-fase-dados-reais.md)).
- LGPD, o que falta: registrar o aceite dos termos no cadastro (hoje só há links e aviso; o aceite não é gravado) e confirmar o prazo da cópia de segurança externa, citado de forma genérica na política.
- Depois: painel interno dos shadows, resumo semanal com IA.
- Hardening da VPS e limpeza gradual do que restar de dívida técnica.
- Coleta longitudinal de dados reais antes de dar autoridade adicional aos shadows.
- Decisão de produto: a hospedagem do frontend usa `vinext` (beta) com dependências herdadas do ambiente de criação (`wrangler`, `@openai/sites-vite-plugin`); avaliar migração para uma base mais estável.
