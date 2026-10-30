# Traefik deployment

Both locations use Traefik's file provider. Docker discovery is disabled, no
Docker socket is mounted, and application Compose files do not contain Traefik
labels.

## NL10

NL10 retains its prior IP-based routing. Services on `10.10.60.200` must keep
their published ports available to the Traefik host:

- Karakeep: 3000
- Frigate authenticated UI: 8971
- Actual Budget: 5006, 5007, and 5008

Kanidm remains reachable on `10.10.50.100:8443`. oauth2-proxy shares the
existing `nl10-traefik_traefik-net` network with Traefik.

## NL00

NL00's former Caddy routes are declared in
`traefik (nl00)/config/dynamic_config.yml`. Traefik joins the same external
networks previously used by Caddy. OpenBao and Semaphore explicitly join
`caddy-net` so their Docker DNS names remain reachable.

## Published application ports

Uptime Kuma and Vaultwarden on NL00 have no host-published ports. Traefik
reaches them through `uptime-net` and `vaultwarden-net` using the existing
container addresses. Redeploy those application stacks to remove their old
host listeners; the Traefik routes require no changes.

Keep NL10's published backend ports where Traefik connects from another
host. Removing those mappings would also remove the proxy's access. Frigate's
go2rtc port 1984 remains published pending confirmation of its consumers.

The client-to-server UniFi restriction covers destination ports 80 and 443,
with the admin devices and proxy HTTPS route allowed first. It does not block
direct access to applications on other published ports, or traffic within
one VLAN. Restrict remote backend listeners to their intended proxy or
integration clients separately when stronger isolation is required.

## Traefik Manager ownership

Arcane/Git remains authoritative for Compose infrastructure, images, networks,
ports, volumes, environment wiring, and the NL10 Manager Agent service. The live
file-provider configuration is no longer mounted from the Git-synchronised
project directory:

- NL00: `/home/docker-secure/docker-state/traefik/dynamic`
- NL10: `/home/service/docker-state/traefik/dynamic`

Before redeploying either stack, copy the complete current `./config` directory
to its corresponding runtime directory. For NL10 this includes `dynamic.yml`,
`certs/`, and the local untracked secrets configuration if present. NL00's
KanIDM CA is an existing independent bind mount at `/etc/traefik/certs`, so it
does not belong in the Manager-owned runtime directory. After this one-time
seed, Arcane/Git must not sync `./config` into either runtime directory;
Traefik Manager writes the runtime config and makes backups before changes.

Frigate's design is unchanged: oauth2-proxy remains its own Compose service
with KanIDM/OIDC secrets outside Traefik Manager. The seeded NL10 configuration
retains the ForwardAuth, error middleware, `/oauth2/` router, and service
definitions, which Traefik Manager can import/adopt.

## Deployment

Custom error pages and unknown-host fallbacks are prepared in
`traefik-error-pages/README.md`. Both Compose stacks build the shared shell-free
renderer. Import each `config/northlake-errors.yml` through Manager separately;
Git seed-route changes do not update live routes. Verify wildcard DNS and TLS
before expecting a browser to display the fallback page.

1. Seed the two runtime config directories as described above, before changing
   the bind mounts.
2. Add the new `TMA_*` values to each deployed Arcane environment. Set the
   bind addresses to private/Tailscale addresses, not `0.0.0.0`.
3. Run `docker compose config` in each changed project.
4. Deploy NL00's `traefik-manager` first. It connects directly to the local
   Traefik API and runtime configuration; no NL00 agent or API key is needed.
5. Add the NL10 agent using its private/Tailscale URL on port 8090, set
   `TMA_NL10_API_KEY`, and restrict its host firewall so only NL00 can reach it.
6. Deploy/restart the agents and Traefik stacks. Import/adopt the seeded
   file-provider config in Traefik Manager; do not recreate Frigate auth.
