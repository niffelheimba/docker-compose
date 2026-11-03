# Forgejo on Ubuntu Internal (nl10 Multi)

Pinned official rootless Forgejo 17.0.1 with SQLite and persistent named volume
`forgejo_forgejo-data`. Arcane Git sync source: `forgejo/nl10/compose.yaml`,
project `forgejo`. The `nl00` starter remains independent and undeployed.

## Access

Web URL: https://git.home.northlake.dev/ through NL10 Traefik, managed by the
NL00 Traefik Manager. Route snapshot: `docs/traefik/nl10/forgejo.yml`.
The route retains the existing private-network restriction and error pages.
The backend remains `http://10.10.60.200:3030`; secure session cookies require
using the HTTPS frontend. Git-over-SSH uses `git.home.northlake.dev:2222` via the same proxy.
Traefik forwards raw TCP to `10.10.60.200:2222`, preserving Forgejo SSH host keys.
The private-interface listener and TCP allowlist match the web access boundary.
Example: `ssh://git@git.home.northlake.dev:2222/nicholas/docker-compose.git`.

## Kanidm

Confidential client ID: `git`. Landing URL: `https://git.home.northlake.dev/`.
Forgejo authentication source: `kanidm`, provider `openidConnect`.
Discovery: `https://idm.northlake.dev/oauth2/openid/git/.well-known/openid-configuration`.
Exact callback: `https://git.home.northlake.dev/user/oauth2/kanidm/callback`.
PKCE and strict redirect matching remain enabled; legacy crypto is not enabled.
Scopes: `openid email profile`. The custom array claim `git_roles` is required
with value `user`; `admin` maps to Forgejo administrator. Kanidm scope/claim
maps use `git_users` and `git_admins`, without exporting unrelated memberships.

Membership direction: `git_users` contains `northlake_users`; `git_admins`
contains `northlake_admins`. Thus baseline users/admins inherit application
roles. Existing individual members are preserved. Do not add the reverse
relationships, which would grant broader baseline roles and risk cycles.

The client secret lives only in Kanidm and Forgejo's persistent auth-source
configuration; never put it in Compose, a synced `.env`, Git or documentation.
Automatic OAuth account provisioning is enabled for authorized Kanidm users;
manual public registration remains disabled. Existing-account linking requires
local authentication (`ACCOUNT_LINKING=login`), never automatic email matching.
The local initial administrator remains available as a recovery account.

## Persistence and recovery

Installation is locked. Initial local admin credential is stored privately on
Ubuntu Internal at `/etc/forgejo/initial-admin.txt`; change it at first login.
Back up repositories, SQLite, configuration and keys in the entire data volume
consistently (stop Forgejo for the simple SQLite method). Preserve that volume
across updates; never use `down -v` as an update procedure. Image rollback after
a database migration needs a compatible pre-upgrade data backup.

Existing GitHub remotes and Arcane syncs stay unchanged. Migrate repositories
and those integrations deliberately later. No Forgejo Actions runner is added.

References: https://forgejo.org/docs/v17.0/admin/installation/docker/ and
https://kanidm.github.io/kanidm/stable/accounts/groups.html.
