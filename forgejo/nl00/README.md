# Forgejo on nl00

Prepared Compose definition; not deployed. The official rootless 17.0.1 image
is pinned to its verified manifest digest. This is a new-install baseline, not
a claim that Forgejo is already running.

## Configuration

Copy `.env.example` to `.env`, or use its values in Arcane. Set
`FORGEJO_DOMAIN` and `FORGEJO_ROOT_URL` to your intended hostname/origin.

The project name is explicitly `forgejo`. Keep that project name and its
`forgejo-data` volume across updates. The image runs as container UID/GID
1000:1000; these are container IDs, not the host rootless Docker account IDs.

SQLite is used for this single-container starter. The persistent volume at
`/var/lib/gitea` contains repositories, the database, configuration and keys.
The verified image sets `GITEA_APP_INI` to
`/var/lib/gitea/custom/conf/app.ini`.

## Proxy and ports

The stack joins the existing external `arcane-net` network. Traefik Manager
must be configured separately with an HTTPS route to `http://forgejo:3000`.
The nl00 proxy uses the file provider, so this stack intentionally has no
Docker-provider labels. DNS and proxy changes have not been made.

Host HTTP defaults to `127.0.0.1:3030`. Git-over-SSH defaults to
`127.0.0.1:2222` and the image's built-in SSH listener on port 2222.
Select a suitable SSH bind interface before expecting direct remote SSH
clones. HTTPS Git access works through the configured HTTPS proxy route.

## First deployment

1. Set the hostname/origin and review the bind addresses in the environment.
2. Verify the chosen proxy network exists on the target Docker daemon.
3. Import `forgejo/nl00/compose.yaml` into Arcane, or run
   `docker compose --env-file .env -f compose.yaml up -d` from this directory.
4. Access the web installer over a controlled route or SSH tunnel, create the
   first administrator, and complete installation. Confirm the installer has
   set the persistent `INSTALL_LOCK` setting.
5. Registration is disabled and anonymous viewing requires sign-in. Create
   additional users through administration or configure Kanidm/OIDC separately.
6. Create the destination repository and migrate/mirror it before changing
   your local Git remote or configuring Renovate.

For a temporary setup tunnel to the cloud host, forward a local port to its
loopback HTTP port, for example `ssh -L 3030:127.0.0.1:3030 <host>`.
If installing through that tunnel, temporarily set a matching HTTP root URL,
then restore the final HTTPS origin before regular use.

## Backups and upgrades

Back up the whole data volume consistently, including configuration and keys.
For a simple consistent SQLite/volume backup, stop this service while taking
the backup. Do not remove the volume when updating the image. Review Forgejo's
major-version upgrade instructions before changing the pinned major release.

No Forgejo Actions runner, Renovate process, Docker socket mount or repository
migration is included here.

References:
- [Official Docker/rootless installation](https://forgejo.org/docs/v17.0/admin/installation/docker/)
- [Forgejo releases](https://forgejo.org/download/)
