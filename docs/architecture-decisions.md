# Decisões de arquitetura

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

## Próxima implementação planejada

1. Criar PostgreSQL local e configuração de ambiente de exemplo.
2. Definir o esquema inicial e as migrações.
3. Criar a API REST em Go com acesso ao banco.
4. Implementar autenticação segura no backend.
5. Conectar o frontend exclusivamente pela API HTTPS.
6. Preparar implantação na VPS, TLS, firewall, backups e restauração testada.

