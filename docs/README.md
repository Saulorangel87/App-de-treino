# Documentação do Cadência

Este índice organiza as fontes de referência do projeto e separa visão de produto, estado operacional, decisões técnicas e procedimentos de produção.

## Fontes canônicas

- [`planejamento.md`](../planejamento.md): visão original, escopo do MVP e histórico de decisões do produto. Este arquivo fica preservado na raiz.
- [`melhorias.md`](../melhorias.md): roadmap atual da próxima fase de inteligência, mantido na raiz do projeto.
- [`project-status.md`](project-status.md): inventário do que está implementado, validado, em produção e pendente.
- [`architecture-decisions.md`](architecture-decisions.md): decisões de arquitetura aceitas e seus limites.
- [`training-adaptation-rules.md`](training-adaptation-rules.md): comportamento atual do motor `rules-v1`, adaptação e evidências gerais.
- [`training-cycle-lifecycle.md`](training-cycle-lifecycle.md): estados, transições e regras do ciclo de treino.
- [`cycling-evidence-catalog.md`](cycling-evidence-catalog.md): catálogo de evidências e critérios de elegibilidade dos protocolos específicos de ciclismo.
- [`README.md`](../README.md): visão geral, instalação local, rotas e fluxo funcional.
- [`infrastructure/README.md`](../infrastructure/README.md) e [`infrastructure/cadencia/README.md`](../infrastructure/cadencia/README.md): somente operação da produção na VPS Oracle.

## Como atualizar a documentação

1. Registre a mudança no documento canônico mais específico e atualize `project-status.md` quando o estado do projeto mudar.
2. Se a mudança for visível para o usuário, atualize na mesma entrega `frontend/lib/release.ts`, incrementando `APP_VERSION` e descrevendo a novidade em `UPDATE_NOTES`. A tela de novidades aparece uma vez por conta, versão e navegador.
3. Atualize o `README.md` apenas quando instalação, API ou comportamento público forem afetados.
4. Atualize `architecture-decisions.md` somente quando uma decisão, um limite ou o estado de uma decisão mudar.
5. Registre separadamente o que está no checkout local e o que foi publicado. Não trate um commit local como produção.
6. Nunca inclua segredos, tokens, senhas ou conteúdo de `.env` na documentação.

## Estado documentado neste ciclo

A produção oficial registrada está no commit `c768ef7`, na VPS Oracle, exposta pelos domínios `cadencia.devsaulo.com.br` e `cadencia-api.devsaulo.com.br`; a migração `000015`, o catálogo inicial e o piloto `road_moderate_intervals` foram publicados após backup, validação e autorização explícita. A versão comunicada ao usuário é `0.7.0`. O checkout local está no commit `64e554d`, que já versiona prontidão observacional, histórico 7/28/42, qualidade temporal, comparação por períodos e a avaliação shadow do `rules-v2`; a matriz controlada e a conferência manual da API foram concluídas. O estado e as lacunas de validação estão em [`project-status.md`](project-status.md); não confundir implementação local com produção.

## Organização avaliada

Não foi identificada uma pasta de documentação redundante ou sem função. A pasta `docs/` reúne documentos canônicos referenciados pelo projeto, enquanto os READMEs de `infrastructure/` ficam junto dos arquivos de operação. Portanto, nenhuma pasta ou arquivo documental foi excluído nesta organização.
