# Decisões de arquitetura

Última revisão: 31 de agosto de 2026.

## ADR-001 — Banco de dados próprio

**Status:** Aceita

O Cadência usará PostgreSQL controlado pelo proprietário do projeto.

- Desenvolvimento e testes: PostgreSQL local, preferencialmente executado com Docker Compose.
- Produção: PostgreSQL hospedado na VPS Oracle do proprietário.
- O ambiente de desenvolvimento e o de produção usarão o mesmo mecanismo de banco e as mesmas migrações.
- A aplicação não dependerá do banco gerenciado da hospedagem do frontend.
- Endereços, usuários, senhas e certificados serão fornecidos por variáveis de ambiente e nunca serão gravados no repositório.
- A conexão de produção deverá usar TLS, usuário exclusivo da aplicação, privilégios mínimos, firewall restritivo e backups automatizados.
- O backend em Go será o único componente autorizado a acessar o banco. O navegador nunca receberá credenciais do PostgreSQL.

## Estratégia de ambientes

### Local

O repositório terá um `compose.yaml` com PostgreSQL para permitir subir um ambiente reproduzível. Os dados locais ficarão em volume separado e não serão enviados ao repositório.

### Produção

O backend em Go será executado na VPS ou em outro servidor autorizado a alcançar o PostgreSQL da VPS. A configuração usará uma variável `DATABASE_URL` específica do ambiente.

### Migrações

Todas as alterações de estrutura serão versionadas em arquivos SQL. O mesmo conjunto de migrações será aplicado primeiro localmente e depois na VPS, com backup antes de mudanças de produção.

## Fluxo de acesso

```text
Frontend React
      |
      | HTTPS / REST
      v
Backend Go
      |
      | conexão PostgreSQL protegida
      v
PostgreSQL local (desenvolvimento)
ou PostgreSQL na VPS Oracle (produção)
```

## Estado da implementação

Concluído localmente:

1. PostgreSQL 17 em Docker Compose, exposto somente em loopback.
2. Migrações SQL versionadas até `000006`.
3. API REST em Go com verificações de saúde e prontidão.
4. Autenticação com bcrypt, sessões opacas, hash do token e cookie `HttpOnly`.
5. Frontend conectado exclusivamente à API.
6. Onboarding persistente, motor `rules-v1`, planos de quatro semanas e aprovação transacional.
7. Ciclo de vida das sessões, feedback pós-treino e adaptação conservadora das próximas cargas.
8. Encerramento automático do plano e geração do próximo ciclo sem apagar o histórico.
9. PWA, interface responsiva e validações locais de integração.

## Próximas implementações

1. Expor e apresentar o histórico de atividades concluídas e canceladas.
2. Testar instalação e comportamento do PWA em um celular por origem HTTPS.
3. Preparar a produção na VPS Oracle com TLS, firewall, serviço do backend, usuário PostgreSQL de privilégio mínimo, backups automatizados e teste de restauração.

Uma implantação só será considerada concluída depois de validar HTTPS, CORS, cookies seguros, migrações, backup e restauração.

