# Estado atual do projeto Cadência

Última atualização: 31 de agosto de 2026.

Este documento é o ponto principal de retomada do projeto. Ele registra decisões, funcionalidades implementadas, validações e próximas etapas. Não devem ser incluídas aqui senhas, tokens ou outras credenciais.

## Objetivo

Construir uma aplicação inteligente de planejamento adaptativo de treinos de ciclismo com cadastro do atleta, avaliação de objetivos, disponibilidade e limitações, geração de planos, acompanhamento de sessões e ajustes baseados no feedback e na recuperação.

O mecanismo atual é uma regra conservadora e explicável de produto. Ele não realiza diagnóstico nem substitui prescrição ou acompanhamento profissional.

## Repositório e publicação

- Repositório GitHub: <https://github.com/Saulorangel87/App-de-treino>
- Branch utilizada: `master`.
- Frontend publicado para demonstração: <https://cadencia-treino-inteligente.sauloleonardo1987.chatgpt.site>
- A URL publicada é a última demonstração registrada e pode não conter as alterações locais mais recentes. O fluxo autenticado completo continua sendo validado localmente porque depende da API Go e do PostgreSQL.

## Arquitetura definida

```text
Frontend React/TypeScript
        |
        | HTTPS / API REST
        v
Backend Go
        |
        | conexão protegida
        v
PostgreSQL local (desenvolvimento)
ou PostgreSQL próprio na VPS Oracle (produção)
```

- O navegador nunca acessa o PostgreSQL diretamente.
- O backend Go é o único componente que acessa o banco.
- Desenvolvimento e testes usam PostgreSQL local no Docker.
- Produção usará PostgreSQL controlado pelo proprietário na VPS Oracle.
- Configurações sensíveis ficam em `.env`, que está ignorado pelo Git.
- Alterações no banco são versionadas em migrações SQL.

## Organização do repositório

- `frontend/`: aplicação React/TypeScript baseada em Vinext e Sites.
- `backend/`: API REST escrita em Go.
- `database/migrations/`: migrações PostgreSQL.
- `database/tests/`: verificações SQL das migrações e regras do banco.
- `api/openapi.yaml`: contrato OpenAPI da API.
- `infrastructure/`: futura configuração de produção na VPS Oracle.
- `scripts/`: scripts auxiliares para desenvolvimento local.
- `docs/`: decisões, regras e este registro de continuidade.

## Banco de dados local

- Container: `cadencia-postgres`.
- Imagem: PostgreSQL 17 Alpine.
- Banco: `cadencia_dev`.
- Usuário local: `cadencia`.
- Endereço local confirmado: `127.0.0.1:5433`.
- A porta `5433` foi escolhida porque a `5432` já estava ocupada por outro processo PostgreSQL no Windows.
- O pgAdmin foi configurado e as tabelas foram visualizadas com sucesso.
- A senha local existe somente no `.env` ignorado e não deve ser copiada para documentos ou commits.

O esquema contém usuários, perfis, objetivos, disponibilidade, recuperação, limitações, planos, treinos, sessões, feedback e sessões de autenticação. As migrações atuais vão de `000001` a `000006`.

## Backend implementado

A API Go possui:

- `GET /health`: informa se o processo está ativo.
- `GET /ready`: verifica o acesso ao PostgreSQL.
- `POST /v1/auth/register`: cria usuário e inicia uma sessão.
- `POST /v1/auth/login`: autentica o usuário.
- `POST /v1/auth/logout`: encerra a sessão atual.
- `GET /v1/me`: retorna o usuário autenticado.
- `GET /v1/profile`: consulta o perfil básico.
- `PUT /v1/profile`: cria ou atualiza o perfil básico.
- `GET /v1/onboarding`: consulta limitações, objetivos e disponibilidade.
- `PUT /v1/onboarding/limitations`: substitui as limitações ativas.
- `PUT /v1/onboarding/goals`: salva até dois objetivos priorizados.
- `PUT /v1/onboarding/availability`: salva os sete dias da semana.
- `POST /v1/plans/generate`: gera um rascunho explicável de quatro semanas.
- `GET /v1/plans/current`: consulta plano ativo, rascunho ou ciclo concluído mais recente.
- `POST /v1/plans/{planID}/activate`: ativa um rascunho de forma transacional.
- `POST /v1/workouts/{workoutID}/start`: inicia uma sessão planejada.
- `POST /v1/workouts/{workoutID}/complete`: conclui a sessão e salva o feedback.
- `POST /v1/workouts/{workoutID}/cancel`: cancela uma sessão em andamento e preserva o histórico.

Comportamentos de domínio implementados:

- O motor `rules-v1` gera quatro semanas respeitando experiência, objetivos, limitações e disponibilidade.
- O feedback pós-treino adapta conservadoramente as próximas sessões e registra a justificativa no treino.
- Dor, fadiga, dificuldade e diferença de RPE podem reduzir a próxima carga; uma resposta claramente fácil permite somente uma pequena progressão de duração.
- O plano ativo é concluído automaticamente quando não restam sessões planejadas ou em andamento.
- O próximo ciclo começa sem sobrepor o ciclo concluído e preserva todos os dados anteriores.

Segurança aplicada:

- Senhas protegidas com bcrypt.
- Tokens de sessão opacos; somente o hash SHA-256 é armazenado.
- Cookie de sessão `HttpOnly` e também `Secure` quando `APP_ENV=production`.
- CORS configurado por variável de ambiente.
- PostgreSQL local exposto somente em loopback.
- Índice parcial garante no máximo um plano ativo por atleta.
- Alterações de sessão, feedback e adaptação acontecem de forma transacional.

## Frontend implementado

- Dashboard autenticado e responsivo.
- Rotas `/entrar`, `/perfil` e `/plano`.
- Perfil com quatro etapas persistentes: dados básicos, limitações, objetivos e disponibilidade.
- Geração, revisão e ativação do plano de quatro semanas.
- Estado de ciclo concluído com ação para gerar o próximo ciclo.
- Modal de detalhes e estrutura do treino.
- Início, conclusão, cancelamento e feedback da sessão.
- Adaptações automáticas explicadas no dashboard, modal e plano.
- Indicadores de prontidão, carga semanal e explicabilidade.
- Ajuda contextual de RPE com escala de 1 a 10.
- Integração com a API por `VITE_API_URL`.
- Documento configurado com `lang="pt-BR"`.
- PWA com manifesto, ícones, instalação, tela offline e cache apenas de recursos estáticos.
- Rodapé fixo com autoria, versão, LinkedIn, GitHub, e-mail e ação de instalação.
- Contraste, tooltips, tipografia mobile e responsividade revisados.
- Scroll da página bloqueado durante o modal; o modal permanece acima do rodapé.
- Meta viewport explícita e página sem rolagem horizontal no mobile.

## Validações realizadas

- Cadastro, login, cookie de sessão e `/v1/me` testados no banco local.
- Criação e retomada do perfil testadas no PostgreSQL.
- Motor `rules-v1` testado com 12 sessões em quatro semanas sem ultrapassar a disponibilidade.
- Fluxo de integração validado: cadastro temporário, onboarding, geração, ativação e leitura do plano.
- Ciclo de sessão validado: iniciar, concluir com feedback, iniciar outra sessão e cancelar.
- Migração `000004` validada para estados transacionais das sessões.
- Migração `000005` implementada para adaptação por feedback e coberta por testes das decisões.
- Migração `000006` implementada para conclusão automática e correção de planos antigos sem sessões pendentes.
- Geração do próximo ciclo validada sem sobreposição.
- Build do frontend concluído após os ajustes de contraste, modal e responsividade.
- Meta viewport servida confirmada como `width=device-width, initial-scale=1, viewport-fit=cover`.

O Windows App Control pode bloquear executáveis temporários sem assinatura produzidos por `go test`. Por isso, os testes Go são executados no container oficial do Go, sem reduzir a segurança do Windows.

## Como iniciar o ambiente local

1. Abrir o Docker Desktop e aguardar o estado “Running”.
2. Na raiz do repositório, executar `docker compose up -d postgres`.
3. Iniciar a API com `pwsh -NoProfile -File scripts/run-api.ps1`.
4. Em outro terminal, entrar em `frontend/` e executar `npm run dev`.
5. Acessar `http://localhost:3000`.
6. Verificar a API em `http://localhost:8080/health` e a prontidão em `http://localhost:8080/ready` quando necessário.

Valores reais devem continuar somente nos arquivos `.env` locais. O `.env.example` usa valores ilustrativos e deve permanecer sem segredos.

## Pedidos já registrados

- Usar banco PostgreSQL próprio, local no desenvolvimento e na VPS Oracle em produção.
- Footer verde-escuro e minimalista com contatos.
- PWA instalável para celular.
- HTML em português do Brasil.
- Explicação acessível de RPE para atletas iniciantes.
- Interface mobile legível, sem rolagem lateral e com modais que não movimentem o fundo.

## Próximas etapas recomendadas

1. Criar um endpoint autenticado para o histórico de atividades concluídas e canceladas.
2. Criar a rota `/atividades`, ativar o item correspondente na navegação e apresentar data, duração, RPE, dificuldade, fadiga e dor.
3. Testar a instalação do PWA em um celular por uma origem HTTPS acessível pelo aparelho.
4. Preparar produção na VPS Oracle: TLS, firewall, serviço do backend, usuário PostgreSQL de privilégio mínimo, migrações, backups e teste de restauração.

## Estado do Git no momento deste registro

O commit mais recente confirmado localmente e no remoto antes desta atualização documental é:

`fca0a77 fix: corrige responsividade, tipografia e scroll do modal`

A árvore de trabalho estava limpa antes da atualização dos documentos. As alterações documentais devem ser revisadas e commitadas antes de iniciar a próxima etapa funcional.

## Como retomar em uma conversa nova

Em uma nova conversa, informar que o projeto está em `C:\Users\saulo\Documents\App de treino` e pedir para:

1. ler `README.md`, este arquivo, `docs/architecture-decisions.md`, `docs/training-cycle-lifecycle.md` e `docs/training-adaptation-rules.md`;
2. conferir `git status` e os commits recentes;
3. resumir o estado encontrado antes de alterar arquivos;
4. retomar pela implementação do histórico de atividades.

Esses documentos, o código e o histórico do Git fornecem o contexto necessário sem depender da conversa anterior.
