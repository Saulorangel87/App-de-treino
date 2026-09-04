# Infraestrutura

Esta pasta contém a configuração versionada de produção do Cadência na VPS Oracle. A composição Docker, o Cloudflare Tunnel dedicado, as migrações controladas e as rotinas de backup já estão implantados e validados.

Nenhuma credencial de produção deve ser armazenada no repositório. O arquivo `.env.production` existe somente na VPS e contém os segredos do PostgreSQL, Resend e Cloudflare.

O último deploy oficial documentado é o commit `33de28a`, publicado somente na VPS Oracle; para essa alteração de interface, apenas o container `cadencia-frontend-1` foi reconstruído. Os commits locais posteriores, incluindo o catálogo `000015` e o piloto de estrada, não são produção. Consulte `cadencia/README.md` para o primeiro deploy, atualizações, backup e restauração. A situação operacional atual, o inventário de serviços e as pendências de firewall estão em `docs/project-status.md`. O ambiente Sites não faz parte do fluxo de produção do Cadência.
