# Homelable on nl10

Based on the upstream [compose example](https://github.com/Pouzor/homelable/blob/main/docker-compose.yml), using its [prebuilt images](https://github.com/Pouzor/homelable/blob/main/docker-compose.prebuilt.yml) so this folder deploys without a source checkout. See upstream [installation instructions](https://github.com/Pouzor/homelable/blob/main/INSTALLATION.md).

## Configure and start

Copy `.env.example` to `.env`, then configure:

| Variable | Value |
| --- | --- |
| `SECRET_KEY` | A fresh random secret, at least 32 bytes; `openssl rand -hex 32` works. |
| `OIDC_CLIENT_SECRET` | Retrieve from Kanidm using the command below; retain single quotes around the value. |
| `SCANNER_RANGES` | JSON list of intended LAN CIDRs. The template has no targets. |
| `HOMELABLE_BIND_IP` | Interface reachable from the reverse proxy; defaults to all IPv4 interfaces. |
| `HOMELABLE_PORT` | Unused host port; defaults to `3010` to avoid Karakeep's `3000`. |

The remaining OIDC values are preconfigured for `https://homelable.home.northlake.dev` and client `homelable` at `https://idm.northlake.dev`.

Retrieve the already-created confidential client's secret locally:

```powershell
kanidm -H https://idm.northlake.dev -D idm_admin system oauth2 show-basic-secret homelable
```

From this folder on the Docker host:

```sh
docker compose config --quiet
docker compose pull
docker compose up -d
```

Configure LAN DNS and a Traefik Manager HTTPS route for `homelable.home.northlake.dev` to `http://<Docker-host-LAN-IP>:3010`. Traefik runs on the separate ProxyAuth host; it cannot use this stack's Docker service names. Forward the entire path, including `/api/` and WebSockets. Use Kanidm's built-in OIDC login. The route and DNS must be configured before testing browser login.

The backend and SQLite database have no published ports. Back up the `backend_data` volume, including uploads. Docker bridge networking supports routed scans but cannot discover LAN MAC addresses. No scans are run by this repository setup.

`HOMELABLE_VERSION` defaults to upstream `latest`; set a common published release tag for reproducible upgrades. Back up the volume before upgrading.

## Kanidm access

Membership flows from the existing Northlake groups into the app groups:

```text
homelable_users
  member: northlake_family_users
homelable_admins
  member: northlake_admins
```

This means family users inherit `homelable_users`, and Northlake admins inherit `homelable_admins`. Do not add the app groups as members of the Northlake groups: that reverses inheritance and can grant unrelated Northlake access.

Both app groups have `openid profile email` scope maps on client `homelable`. The landing URL is `https://homelable.home.northlake.dev`; the exact registered callback is `https://homelable.home.northlake.dev/api/v1/auth/oidc/callback`. PKCE remains enabled and strict redirect matching is enabled.

The current upstream app has no separate OIDC user/admin role enforcement. Both groups can sign in with the same application permissions; the groups preserve distinct membership for future role support.

## Optional MCP

Set `MCP_API_KEY` and `MCP_SERVICE_KEY` to two independent random secrets in `.env`, then run `docker compose --profile mcp up -d`. Port `8001` binds to loopback by default. Change `MCP_BIND_IP` only when an intended LAN client needs access, and restrict access to that client. The endpoint is `/mcp/` and clients send the `X-API-Key` header. Leave the profile disabled if unused.
