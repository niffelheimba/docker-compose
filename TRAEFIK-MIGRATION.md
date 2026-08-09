# Traefik migration

Each location runs one Traefik container on its existing host address. Public
and private routers share HTTPS port 443; private routers require the
`private-net@file` middleware.

## Prerequisites

1. Copy the applicable Traefik `.env.example` to `.env` and set the Cloudflare
   DNS API token. The host bind address, ACME email, and Docker socket already
   have location-specific defaults and need values only when overriding them.
2. Ensure the rootless user can write the ACME and log state directories in the
   Traefik Compose file.
3. Start the Traefik Compose project once to create the named `traefik-public`
   and `traefik-private` networks. Application projects reference these shared
   names as external networks.

## Semaphore rename

The stack moved from `semaphore (nl10)` to `semaphore (nl00)`. Copy its local
`.env` to the new directory before deployment. The Compose file pins the MySQL
volume to `semaphore_semaphore-mysql` so the existing database volume is reused.
Confirm this with `docker volume inspect semaphore_semaphore-mysql` before
starting the renamed stack.

## Cutover

1. Run `docker compose config` in every changed stack directory.
2. Stop the old Caddy container on NL00 or old Traefik container on NL10 so it
   releases ports 80 and 443.
3. Start the new Traefik stack; this creates the shared proxy networks.
4. Recreate every changed application stack so Docker applies its labels and
   network attachments.
5. Verify public and private routes, then remove obsolete proxy containers.

Private DNS is not an access control. Every private Docker router must include
`private-net@file`, and private application HTTP ports must not be published on
the host.
