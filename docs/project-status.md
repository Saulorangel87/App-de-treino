# Estado atual do projeto Cadência

Última atualização: 30 de agosto de 2026.

Este documento é o ponto de retomada do projeto. Ele registra o que já foi decidido, implementado e testado, além das próximas etapas. Não devem ser incluídas aqui senhas, tokens ou outras credenciais.

## Objetivo

Construir uma aplicação inteligente de planejamento adaptativo de treinos de ciclismo. O sistema terá cadastro do atleta, avaliação de objetivos, disponibilidade, limitações, geração de plano, acompanhamento de sessões e ajustes com base no feedback e na recuperação.

## Repositório e publicação

- Repositório GitHub: <https://github.com/Saulorangel87/App-de-treino>
- Branch utilizada: `master`
- Frontend publicado para demonstração: <https://cadencia-treino-inteligente.sauloleonardo1987.chatgpt.site>
- A versão publicada ainda representa principalmente o dashboard visual. Autenticação e perfil estão sendo testados localmente porque dependem da API local.

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
- `api/openapi.yaml`: contrato da API.
- `infrastructure/`: futura configuração de produção na VPS Oracle.
- `scripts/`: scripts auxiliares para desenvolvimento local.
- `docs/`: decisões de arquitetura e este registro de continuidade.

## Banco de dados local

- Container: `cadencia-postgres`
- Imagem: PostgreSQL 17 Alpine
- Banco: `cadencia_dev`
- Usuário local: `cadencia`
- Endereço local confirmado: `127.0.0.1:5433`
- A porta 5433 foi escolhida porque a porta 5432 já estava ocupada por outro processo PostgreSQL no Windows.
- O pgAdmin foi configurado e as tabelas foram visualizadas com sucesso.
- A senha local existe apenas no `.env` ignorado e não deve ser copiada para documentos ou commits.

O esquema contém as tabelas de usuários, perfis, objetivos, disponibilidade, recuperação, limitações, planos, treinos, sessões, feedback e sessões de autenticação.

## Backend implementado

A API Go possui:

- `GET /health`: estado básico do processo.
- `GET /ready`: verifica se a aplicação consegue acessar o banco.
- `POST /v1/auth/register`: cria usuário e inicia uma sessão.
- `POST /v1/auth/login`: autentica o usuário.
- `POST /v1/auth/logout`: encerra a sessão atual.
- `GET /v1/me`: retorna o usuário autenticado.
- `GET /v1/profile`: consulta o perfil básico do atleta.
- `PUT /v1/profile`: cria ou atualiza o perfil básico.
- `GET /v1/onboarding`: consulta limitações, objetivos e disponibilidade.
- `PUT /v1/onboarding/limitations`: substitui as limitações ativas.
- `PUT /v1/onboarding/goals`: salva os objetivos priorizados.
- `PUT /v1/onboarding/availability`: salva os sete dias da semana.

Segurança já aplicada:

- Senhas protegidas com bcrypt.
- Tokens de sessão opacos; apenas o hash SHA-256 é armazenado no banco.
- Cookie de sessão `HttpOnly`.
- Em produção, o cookie também exige HTTPS.
- CORS configurado por variável de ambiente.
- PostgreSQL local exposto somente no endereço de loopback.

## Frontend implementado

- Dashboard inicial responsivo.
- Modal de detalhes do treino.
- Indicadores de prontidão, carga semanal e explicabilidade.
- Rota `/entrar` para cadastro e login.
- Rota `/perfil` com quatro etapas persistentes do cadastro do atleta.
- Integração do frontend com a API por `VITE_API_URL`.
- O documento HTML está configurado com `lang="pt-BR"`.
- Acesso ao perfil pelo dashboard.

### Situação do fluxo de perfil

O onboarding salva cada etapa no PostgreSQL e permite avançar e voltar sem perder o que já foi gravado.

Etapas planejadas:

1. Dados básicos e experiência — implementada.
2. Lesões, dores e limitações — implementada.
3. Objetivos esportivos — implementada.
4. Disponibilidade semanal e revisão — implementada.

## Avisos observados no console

- O perfil inexistente agora retorna `200` com valor nulo, evitando o antigo erro `404` no console durante o primeiro acesso.
- Avisos de recursos carregados por `preload` e não usados imediatamente: relacionados ao ambiente de desenvolvimento/empacotamento; não bloqueiam o funcionamento.
- Os campos do perfil agora usam estado controlado desde a inicialização para eliminar o aviso do Base UI sobre mudança de estado não controlado para controlado.
- Mensagem para instalar React DevTools: somente uma recomendação do React durante o desenvolvimento.

## Validações já realizadas

- Cadastro de usuário, cookie de sessão e consulta de `/v1/me` testados contra o banco real local.
- Criação e leitura do perfil testadas contra o banco real local.
- Usuário temporário de integração removido após o teste.
- Compilação e análise do backend concluídas.
- Testes Go executados com sucesso em container Linux.
- Build do frontend concluído com as rotas `/`, `/entrar` e `/perfil`.

Observação: o Windows App Control bloqueia executáveis temporários sem assinatura produzidos pelo `go test`. Por isso, os testes são executados no container oficial do Go, sem reduzir a segurança do Windows.

## Como iniciar o ambiente local

1. Abrir o Docker Desktop e aguardar o estado “Running”.
2. Na raiz do repositório, executar `docker compose up -d postgres`.
3. Iniciar a API com `pwsh -NoProfile -File scripts/run-api.ps1`.
4. Em outro terminal, entrar em `frontend/` e executar `npm run dev`.
5. Acessar a URL local apresentada pelo frontend.

Os valores reais devem continuar somente nos arquivos `.env` locais. O `.env.example` pode usar valores ilustrativos e precisa permanecer sem segredos.

## Pedidos já registrados

- Footer verde-escuro e minimalista semelhante à referência enviada — implementado com autoria, versão, GitHub, LinkedIn, e-mail e ação de instalação.
- PWA instalável — implementado com manifesto, ícones de 192/512 px, ícone maskable, suporte a iOS, service worker e tela offline básica.
- Preservar o idioma do HTML como português do Brasil.

## Próximas etapas recomendadas

1. Testar a instalação do PWA em um celular usando uma origem HTTPS acessível pelo aparelho.
2. Conectar os dados concluídos do onboarding à geração inicial do plano de treino.
3. Preparar produção na VPS: TLS, firewall, usuário PostgreSQL de privilégio mínimo, backups e teste de restauração.

## Estado do Git no momento deste registro

As alterações de autenticação e perfil ainda aparecem como modificadas ou novas no diretório de trabalho. Antes de iniciar uma nova grande etapa, revisar `git status`, testar e criar um commit. Sugestão de mensagem:

`Implementa autenticação e perfil básico do atleta`

## Como retomar em uma conversa nova

Em uma nova conversa, informar que o projeto está em `C:\Users\saulo\Documents\App de treino` e pedir para ler este arquivo e `docs/architecture-decisions.md` antes de continuar. Esses dois documentos, junto com o código e o histórico do Git, fornecem o contexto necessário sem depender do histórico completo da conversa anterior.
