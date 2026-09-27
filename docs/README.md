# Documentação do Cadência

Este índice organiza as fontes de referência do projeto e separa visão de produto, estado operacional, decisões técnicas e procedimentos de produção.

## Fontes canônicas

- [`planejamento.md`](../planejamento.md): documento único de visão, escopo do MVP, estado atual e próximos passos do produto.
- [`STATUS.md`](STATUS.md): estado atual, curto: produção, checkout local, pendências e checklist pré-deploy.
- [`changelog/project-status-history.md`](changelog/project-status-history.md): diário cronológico das fatias de trabalho, com validações e decisões (histórico; não é o estado atual).
- [`architecture-decisions.md`](architecture-decisions.md): decisões de arquitetura aceitas e seus limites.
- [`training-adaptation-rules.md`](training-adaptation-rules.md): comportamento atual do motor `rules-v1`, adaptação e evidências gerais.
- [`training-cycle-lifecycle.md`](training-cycle-lifecycle.md): estados, transições e regras do ciclo de treino.
- [`cycling-evidence-catalog.md`](cycling-evidence-catalog.md): catálogo de evidências e critérios de elegibilidade dos protocolos específicos de ciclismo.
- [`proxima-fase-dados-reais.md`](proxima-fase-dados-reais.md): desenho da próxima fase (dados reais de execução): importação de atividades por arquivo `.fit`/`.gpx` com atalho no Android, melhorias no motor a partir desses dados, e a ordem das etapas seguintes; a integração com o Strava fica preservada em anexo, adiada por custo.
- [`roadmap-acceptance.md`](roadmap-acceptance.md): matriz histórica de aceitação técnica; não é um segundo roadmap e não substitui o `planejamento.md`.
- [`README.md`](../README.md): visão geral, instalação local, rotas e fluxo funcional.
- [`infrastructure/README.md`](../infrastructure/README.md) e [`infrastructure/cadencia/README.md`](../infrastructure/cadencia/README.md): somente operação da produção na VPS Oracle.

## Como atualizar a documentação

1. Registre a mudança no documento canônico mais específico e atualize `STATUS.md` quando o estado do projeto mudar (mantenha-o curto; o detalhe cronológico vai para `changelog/project-status-history.md`).
2. Se a mudança for visível para o usuário, atualize na mesma entrega `frontend/lib/release.ts`, incrementando `APP_VERSION` e descrevendo a novidade em `UPDATE_NOTES`. A tela de novidades aparece uma vez por conta, versão e navegador.
3. Atualize o `README.md` apenas quando instalação, API ou comportamento público forem afetados.
4. Atualize `architecture-decisions.md` somente quando uma decisão, um limite ou o estado de uma decisão mudar.
5. Registre separadamente o que está no checkout local e o que foi publicado. Não trate um commit local como produção.
6. Nunca inclua segredos, tokens, senhas ou conteúdo de `.env` na documentação.

## Estado documentado neste ciclo

A produção oficial está na VPS Oracle, exposta pelos domínios `cadencia.devsaulo.com.br` e `cadencia-api.devsaulo.com.br`; a versão visível é `0.34.0` e as migrações estão aplicadas até `000031`. A entrega publicada permite importar atividades reais de execução por arquivo `.fit`/`.gpx`. O estado atual e as pendências operacionais estão em [`STATUS.md`](STATUS.md).

## Organização avaliada

Não foi identificada uma pasta de documentação redundante ou sem função. A pasta `docs/` reúne documentos canônicos referenciados pelo projeto, enquanto os READMEs de `infrastructure/` ficam junto dos arquivos de operação. Portanto, nenhuma pasta ou arquivo documental foi excluído nesta organização.
