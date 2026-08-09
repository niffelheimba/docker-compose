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

## Deployment

1. Copy each Traefik `.env.example` to `.env` and set `CF_DNS_API_TOKEN`.
2. Run `docker compose config` in each changed project.
3. Recreate applications whose ports or network attachments changed.
4. Restart Traefik and check its logs for file-provider or ACME errors.
