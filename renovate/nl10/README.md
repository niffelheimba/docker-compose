# Renovate on Ubuntu Internal (nl10 Multi)

Pinned official Renovate CLI, run hourly in a single non-overlapping loop.
No AI, Docker socket, published ports, automatic merge or deployment access.
The current target is `niffelheimba/docker-compose` on GitHub. Arcane Git syncs
remain on GitHub; this stack does not migrate repositories to Forgejo.

## Git sync

Use `renovate/nl10/compose.yaml`, project `renovate`, and sync the whole directory.
`config.js` and `run.sh` must accompany the Compose file. Credentials live outside
the synced project at `/etc/renovate/github-token`; `.env.example` has no secrets.

## Credential handoff

An empty host token file is intentional at initial deployment: the container
stays running and logs `WAITING FOR CREDENTIAL`, but does not run update jobs.
Supply a dedicated GitHub bot credential later, restricted to the target repo
with the permissions Renovate needs to read/write code, issues and pull requests.
Keep the parent directory root-owned, mode 0700. The root-owned token file can
be mode 0444 inside that protected directory, so the non-root container can read
its bind-mounted Compose secret. Do not paste credentials into Compose or Git.

Write the token into the existing file (do not replace the inode), or recreate
only this service after an atomic file replacement so its secret mount refreshes.
The next hourly iteration will pick up an in-place update. An immediate check
can be requested by restarting only the Renovate service. Validate a first run
with `RENOVATE_DRY_RUN=full` before allowing normal PR creation.

The repository's root `renovate.json` limits processing to Compose, pins digests,
requires dashboard approval before update PRs, disables automerge, and groups
Kanidm replicas. Archived/stopped/unverified variants are excluded. An update
PR does not itself deploy anything; merging can invoke existing Arcane syncs.

## Later Forgejo migration

After copying the repository and creating a Forgejo bot credential, set
`RENOVATE_PLATFORM=forgejo`, `RENOVATE_ENDPOINT=https://your-forgejo/api/v1/`,
and `RENOVATE_REPOSITORIES=owner/repository`. Update the mounted credential,
validate a dry run, and recreate this service. Root repository policies can move
with the repository. Arcane sync URLs are a separate, deliberate migration.

## Verification and recovery

Check container logs for waiting, completed checks or errors; a running waiting
container does not prove successful authentication. Stop this project to disable
the updater. Its temporary clone/cache is ephemeral; there is no application
data volume to restore. Keep the host token file out of backups published to Git.

References: https://docs.renovatebot.com/getting-started/running/ and
https://docs.renovatebot.com/modules/platform/forgejo/.
