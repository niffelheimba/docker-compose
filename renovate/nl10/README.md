# Renovate on Ubuntu Internal (nl10 Multi)

The pinned official Renovate CLI runs hourly against the Forgejo pilot repository
`nicholas/docker-compose`. It uses the dedicated restricted, nonadmin `renovate`
account, with write collaboration on this repository only. Its credential has
repository/issue write and user/organization read scopes; it is not a GitHub token.

There is no Docker socket, exposed port, automatic merge or deployment access.
Root `renovate.json` limits processing to Compose, maintains digest pins, creates
patch/minor and digest update PRs automatically, and requires Dependency Dashboard
approval before major update PRs. PR and branch creation have no hourly or
concurrency caps. Container updates are grouped by Compose file so an application
and its database can be reviewed together. Major upgrades have a separate group
and approval checkbox for each file. Kanidm server replicas remain one shared
group across sites (including both Docker Hub image spellings). Automerge is
disabled.
Archived/stopped/unverified variants remain excluded. Approvals happen on
Forgejo's Dependency Dashboard issue. Merging an update here does not deploy it.

## Deployment boundary

Arcane's existing Git syncs still point to GitHub and have not been migrated.
The deployed Renovate project uses a local `.env` override for the Forgejo
platform, API endpoint, repository and token-file location. This is separate
from changing an Arcane Git sync's repository/source configuration. The Forgejo
copy records the intended settings for a later deliberate source migration.
No existing local Git checkout remotes have been repointed.

After that later migration, source path is `renovate/nl10/compose.yaml`, project
`renovate`, with whole-directory sync for `config.js` and `run.sh`. Never include
host credentials in the Git sync. `RENOVATE_ENDPOINT` must end in `/api/v1/`.

## Credential and verification

The token is mounted read-only from `/etc/renovate/forgejo-token`. Keep its parent
directory root-owned mode 0700; the root-owned token file may be mode 0444 inside
that protected directory so the non-root container can read its secret mount.
Keep generated bot recovery credentials in a separate root-only file. Never put
tokens in Compose, repository configuration, chat, or a committed `.env`.

Write token rotations into the existing file or recreate only this service after
replacing the inode. An empty token causes an explicit waiting log, not an update
check. Validate changes with an authenticated `RENOVATE_DRY_RUN=full` run before
restarting the hourly worker for normal operation. Do not overlap manual runs
with a scheduled check. Logs and exit status distinguish successful checks from
the container merely being alive. GitHub API rate limits may affect optional
upstream release/changelog lookup without a separate GitHub read credential.

Stop this project to disable updates. Temporary clone/cache data is ephemeral;
there is no application data volume to restore. The one-time copies of GitHub
repositories are standalone Forgejo repositories, with no upstream sync/mirror.

References: https://docs.renovatebot.com/modules/platform/forgejo/ and
https://docs.renovatebot.com/getting-started/running/.
