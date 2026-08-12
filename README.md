# Northlake Docker Compose
## PVE1
- N/A

## PVE2
- Portainer Server
- Pangolin / Traefik
- Actual Budget
- Frigate
- Semaphore
- Karakeep

## OCI
- Portainer Agent
- Caddy
- Checkmate
- KanIDM

## Rootless Docker on nl00

The reproducible rootless Docker settings are stored under
`rootless-docker (nl00)/`.

- Install `daemon.json` as `/home/docker-secure/.config/docker/daemon.json`.
- Install `docker.service.d/override.conf` as
  `/home/docker-secure/.config/systemd/user/docker.service.d/override.conf`.
- Run `systemctl --user daemon-reload` and restart `docker.service` after
  reviewing the availability impact.

The `slirp4netns` port driver is intentional. On this host, RootlessKit's
`builtin` port driver continued presenting Docker gateway addresses to
Traefik even with Docker's userland proxy disabled. The `slirp4netns` port
driver preserves real TCP client addresses so the private-route IP allowlist
can remain fail-closed. HTTP/3 and UDP 443 remain disabled.
