# Infraestrutura

Esta pasta contém a configuração versionada de produção do Cadência na VPS Oracle. A composição Docker, o Cloudflare Tunnel dedicado, as migrações controladas e as rotinas de backup já estão implantados e validados.

Nenhuma credencial de produção deve ser armazenada no repositório. O arquivo `.env.production` existe somente na VPS e contém os segredos do PostgreSQL, Resend e Cloudflare.

Consulte `cadencia/README.md` para o primeiro deploy, atualizações, backup e restauração. A situação operacional atual e as pendências de firewall estão em `docs/project-status.md`.
