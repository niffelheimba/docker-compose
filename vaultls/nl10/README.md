# VaulTLS on Ubuntu Internal (nl10 Multi)

The official v1.4.1 image is pinned to its verified multi-platform manifest digest.
Project/container name: `vaultls`. Persistent volume: `vaultls_vaultls-data`.

Web URL: https://mtls.home.northlake.dev/ . NL10 Traefik on ProxyAuth terminates
HTTPS and forwards to the private backend `http://10.10.60.200:5173`. NL00
Traefik Manager owns the live route; `docs/traefik/nl10/vaultls.yml` is its
repository snapshot. The route retains existing private-network restrictions
and error pages. No proxy restart is needed for this dynamic route.

## Secrets and startup

Generate a 32-byte random, base64-encoded API secret outside the project at
`/etc/vaultls/api-secret`. Keep the directory root-owned mode 0700 and the file
mode 0400. Compose mounts it read-only at `/run/secrets/VAULTLS_API_SECRET`,
which the application supports natively. Never commit its value or a real `.env`.
The upstream image runs nginx and its Rust backend with the upstream user model.

Deploy from this directory with `docker compose up -d`. This stack does not
create or change an Arcane Git sync. Existing stack syncs remain on GitHub.
If importing the project into Arcane later, preserve the project name, volume,
private backend address and external secret path.

The initial setup wizard lets the owner select administrator and CA settings.
After initial deployment, complete it through the HTTPS URL. No administrator,
OIDC provider or certificate authority is provisioned by the Compose definition.
The health check uses the public backend version endpoint and verifies both
nginx and the Rust API are responding; it does not mean setup is complete.

## Data and recovery

Back up the entire persistent volume consistently along with the API secret,
keeping both confidential. The volume can contain CA/private keys, certificates,
database and settings. Stop this service for a simple consistent volume backup.
Do not remove its volume when upgrading or rolling back. Changing API secrets
can invalidate authentication sessions; retain the original during recovery.

Database encryption (`VAULTLS_DB_SECRET`), OIDC, CRL publication, ACME issuance,
and installing CA trust or enforcing client certificates on other services are
separate configuration steps. This stack does not enable those features or
set `VAULTLS_INSECURE=true`. Review release notes and take data backups before
version changes; rolling back an image may require a matching pre-upgrade backup.

Upstream: https://github.com/7ritn/VaulTLS/tree/v1.4.1 .
