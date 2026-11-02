# Northlake Docker Compose

Deployable stacks are organized by application, then location. `nl00` is cloud;
`nl10` is the home site. A further subfolder identifies multiple instances.

## Layout

| Application | Variants |
| --- | --- |
| Actual Budget | `actual-budget/nl10/{family,jacob,nicholas}/` |
| Alloy | `alloy/nl00/`, `alloy/nl10/multi/`, `alloy/nl10/proxyauth/` |
| Kanidm | `kanidm/nl00/`, `kanidm/nl10/` |
| Traefik | `traefik/nl00/`, `traefik/nl10/` |
| Forgejo starter | `forgejo/nl00/` (not deployed) |
| Arcane manager | `arcane/nl00/` |
| Cloudflare DDNS, Frigate, Homelable, Karakeep, OAuth2 Proxy | each application’s `nl10/` directory |
| Grafana and Loki stack | `observability/nl10/` |
| OpenBao, Semaphore, Uptime Kuma, Vaultwarden | each application’s `nl00/` directory |
| Parseable | `parseable/nl10/` (no verified current deployment) |
| Shared error-page source | `traefik-error-pages/` |
| Stopped Fluxer / historical templates | `fluxer.gg/`, `archive/` (unchanged) |

Verified Compose variants declare their deployed project names explicitly, so
leaf directories such as `nl00` do not become a shared default project name.

Keep each deployment's local `.env`, volumes and project name when changing a
repository path. Relative bind mounts remain relative to that variant's Compose
directory. Moving Git source does not authorize moving live data.

## Image pins

The 2026-11-01 audit pins verified services to their exact deployed release and
manifest digest. See [the inventory](docs/image-inventory-2026-11-01.json) for
references and explicit coverage gaps. Digest pins preserve the actual running
artifact even when a publisher rebuilds a version tag.

- Karakeep is a nightly image; Homelable backend/frontend are main-branch builds.
  Their deployed digests are frozen without claiming they are stable releases.
- Alloy deliberately differs by host: cloud and home Multi run 1.21.0; home
  ProxyAuth runs 1.20.1. Select the correct variant instead of upgrading it
  implicitly through a shared file.
- Archived, undeployed and stopped stacks are not assigned guessed versions.
  Their floating defaults are intentionally preserved. Homelable's optional MCP
  profile is also unverified.
- Local `northlake/error-pages:1.0.0` builds retain their exact project version;
  the source and Dockerfiles remain alongside each Traefik variant.

For Forgejo/Renovate, point the repository URL at its new remote and use its
Docker Compose manager with digest updates. Keep both Kanidm replicas in the
same version update and roll them out deliberately. Review major upgrades and
migrations before deployment. No Renovate container or schedule is created here.

## Arcane path migration

Configured GitOps source paths move to:

| Project | Repository Compose path |
| --- | --- |
| Vaultwarden | `vaultwarden/nl00/docker-compose.yaml` |
| KanIDM | `kanidm/nl00/docker-compose.yaml` |
| OpenBao | `openbao/nl00/docker-compose.yaml` |
| Semaphore | `semaphore/nl00/compose.yaml` |
| Traefik | `traefik/nl00/compose.yaml` |
| Alloy (cloud manager) | `alloy/nl00/compose.yaml` |

The existing runtime project destinations and secrets stay in place. The
inspected GitOps definitions have automatic redeployment and image pulling
disabled. Other manually managed projects should select their corresponding
new source path when next imported or migrated.

## Rootless Docker on nl00

The reproducible settings are stored under `rootless-docker/nl00/`.

- Install `daemon.json` as `/home/docker-secure/.config/docker/daemon.json`.
- Install `docker.service.d/override.conf` as
  `/home/docker-secure/.config/systemd/user/docker.service.d/override.conf`.
- Review availability impact before running `systemctl --user daemon-reload`
  and restarting Docker.

The `slirp4netns` port driver is intentional. On this host, RootlessKit's
`builtin` driver presented Docker gateway addresses to Traefik despite disabling
the userland proxy. `slirp4netns` preserves real client addresses so private-route
allowlists remain effective. HTTP/3 and UDP 443 remain disabled.
