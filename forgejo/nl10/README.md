# Forgejo on Ubuntu Internal (nl10 Multi)

Pinned official rootless Forgejo 17.0.1 with SQLite and persistent named volume
`forgejo_forgejo-data`. This is independent of the undeployed `nl00` starter.
Arcane Git sync source: `forgejo/nl10/compose.yaml`, project `forgejo`.

LAN web URL: `http://10.10.60.200:3030/`. Git SSH port: `2222` on the same IP.
No cloud proxy network, public listener, DNS, TLS, OIDC, Git mirror or runner is
created by this stack. Configure an HTTPS reverse proxy and matching root URL
before using this as your regular Git endpoint over untrusted networks.

Installation is locked from the first start; public registration is disabled
and anonymous repository browsing requires sign-in. Provision the first admin
with Forgejo's `admin user create` command inside the running container. Keep
the generated initial password outside Git and require a change at first login.
The deployment session records the protected host credential-file location.

The data volume contains repositories, SQLite, configuration and keys. Preserve
it across updates and take a consistent backup (stop Forgejo while backing up
the whole volume for the simple SQLite method). Removing the container or Git
sync does not remove the volume; never use `down -v` as an update procedure.

Existing GitHub remotes and Arcane syncs stay unchanged. Clone/mirror repositories
and deliberately migrate those integrations later. Image rollback after a
database migration requires a compatible pre-upgrade data backup.

Reference: https://forgejo.org/docs/v17.0/admin/installation/docker/.
