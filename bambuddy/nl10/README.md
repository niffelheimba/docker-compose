# Bambuddy — home site

Target: NL10-VM-MULTI-01 / `10.10.60.200`, rootless Docker, project `bambuddy`.
URL: https://bambuddy.home.northlake.dev/ through the private-network Traefik router.

## Deployment

Arcane Git Sync uses the existing Forgejo repository, branch `main`, source
`bambuddy/nl10/compose.yaml`, directory sync enabled, deploy mode, five-minute
automatic sync. Only this new project should be deployed.

The required host-only file `/home/service/docker-secrets/bambuddy/oidc.env`
contains `BAMBUDDY_OIDC_CLIENT_SECRET`. Keep its directory mode 0700 and file
mode 0600, owned by service. It is outside the Git checkout and synced project.
Never place secrets in Arcane's source repository, `.env.example`, or this file.
Docker administrators can inspect container environment values.

Initialize authentication while the HTTP bind is loopback-only by setting
`BAMBUDDY_BIND_IP=127.0.0.1` in the new sync's non-secret environment override.
Use `/api/v1/auth/setup` to enable authentication and create a random local
recovery administrator; save that password only in the private host directory.
Then change the override to `10.10.60.200` and redeploy this project. Install
`docs/traefik/nl10/bambuddy.yml` into ProxyAuth's watched dynamic directory.
No Traefik restart is needed.

## KanIDM

Confidential client: `bambuddy`; landing URL as above; strict exact redirect:
`https://bambuddy.home.northlake.dev/api/v1/auth/oidc/callback`.
Issuer: `https://idm.northlake.dev/oauth2/openid/bambuddy`.
Keep PKCE and ES256 enabled. Request `openid email profile`; require verified
email, automatically provision users, and disable automatic account linking.
The `groups` claim maps `bambuddy_admins` (Nicholas) to Bambuddy's
`Administrators` group, and `northlake_family_users` to `Operators`.
Both KanIDM groups have the required scope grant; the claim uses array join.
New accounts otherwise default to Viewers. KanIDM scope grants control who can
sign in. Family access was explicitly requested on 2026-11-26.
The separate local recovery account is retained.

Back up both persistent data and `/app/data/.mfa_encryption_key`; Bambuddy
encrypts the OIDC secret in its database with this generated key. Backups contain
sensitive application data and must remain private.

## X2D and LAN operation

Upstream lists X2D support since 0.2.3b4, including dual-nozzle print-command
handling and RTSP camera port 322. Enable LAN Only and Developer Mode on the
printer, then add its IP, serial and LAN access code in Bambuddy. Do not commit
the access code. No Bambu Cloud account is required for LAN printer management.

This shared rootless host uses bridge networking and publishes only the web
port. Add printers by IP; multicast discovery is not provided. Virtual printers
are deliberately not enabled: existing applications occupy ports 3000/3001,
and rootless host networking does not expose the physical host network.
Do not enable VP in the UI until a separate address/network plan is prepared.
Upstream VP needs bind ports 3000/3002, MQTT 8883, FTP 990, transfer 6000,
camera 322 (and 2024-2026 for A1/P1), plus passive FTP ranges. Non-proxy VPs
use ten ports each starting at 50000; proxy mode needs 50000-50100.
NET_BIND_SERVICE is required for low ports. A dedicated VM/address is the
straightforward option; no host-wide port/security settings are changed here.

## Optional server-side slicing

The `slicer` Compose profile adds OrcaSlicer API on the private project network,
with no host-published API port. It is off by default. Set
`COMPOSE_PROFILES=slicer` in Arcane and redeploy; then enable Use Slicer API in
Bambuddy with `http://orca-slicer-api:3000`. The upstream sidecar is amd64-only.
The app is pinned to stable 1.3.2 and its amd64 manifest; the optional sidecar
follows upstream latest and should be digest-pinned when enabled.

For X2D, confirm the chosen slicer actually offers the correct X2D/nozzle
profiles before printing. Bambu Studio API is an upstream alternative, separate
from this optional Orca profile. Desktop slicing and uploading `.gcode.3mf`
works without any sidecar. No physical print is part of deployment validation.

## Validation and recovery

Check `/health`, Docker health, `/api/v1/auth/status`, OIDC provider visibility,
discovery, exact callback, PKCE S256 and the redirect to KanIDM. Complete a real
browser sign-in to validate token exchange and role mapping; an authorization
redirect alone does not prove that login completed.

To withdraw this new service, disable its Arcane sync, stop only the Bambuddy
project, and remove only its Traefik router file. Retain the data volumes and
host-only recovery credentials. Revert the deployment commit if source rollback
is needed; do not remove volumes or unrelated projects.

Verified upstream references (2026-11-26):
- https://github.com/maziggy/bambuddy/releases/tag/v1.3.2
- https://wiki.bambuddy.cool/getting-started/docker/
- https://wiki.bambuddy.cool/features/authentication/
- https://wiki.bambuddy.cool/reference/printers/
- https://wiki.bambuddy.cool/features/virtual-printer/
- https://github.com/maziggy/bambuddy/blob/v1.3.2/slicer-api/docker-compose.yml
