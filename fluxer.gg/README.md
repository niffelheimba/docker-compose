# Fluxer on Northlake

Public origin: `https://chat.home.northlake.dev`. This is the official Fluxer
self-hosting stack at commit `7d2c51a57f4bdeeca5a716dfd96dc72187671272`, with
its reverse-proxy settings folded into one compose file for Arcane. The Caddy
routing file is embedded as a Compose config, so no additional host files are
needed. All application images use the upstream `v1` channel by default.

Source: https://github.com/fluxerapp/fluxer/tree/7d2c51a57f4bdeeca5a716dfd96dc72187671272/deploy/self-hosting

## Arcane deployment

1. Create a project using `docker-compose.yaml`.
2. Generate the environment on a Linux Docker host with
   `sh generate-env.sh > .env`, or fill the blanks in `.env.example` yourself.
   Paste the result into Arcane's project Environment editor. Keep `.env` private.
3. `FLUXER_EDGE_BIND=10.10.60.148:18080` publishes the edge on Multi-02's LAN
   address, for Traefik on ProxyAuth (`10.10.50.100`). Change the address if you
   deploy Fluxer on a different host. No external Docker network is required.
4. The live Traefik Manager configuration on ProxyAuth routes host
   `chat.home.northlake.dev` to `http://10.10.60.148:18080`, with TLS using
   the existing certificate resolver and the `private-net` access policy.
   DNS must point to ProxyAuth. Traefik routing is managed at runtime, not by
   Git synchronization; Docker labels will not work with its disabled Docker provider.
5. Deploy. Complete the Fluxer initial setup and keep a local administrator account
   while testing SSO. Open `https://chat.home.northlake.dev/admin` to configure SSO.

Only the edge and LiveKit media ports are published. Databases, object storage,
queues, and application services stay on the private Fluxer network. Compose
defaults to `127.0.0.1:18080` if `FLUXER_EDGE_BIND` is omitted; the environment
template sets the LAN bind needed by ProxyAuth. Restrict inbound port 18080 to
ProxyAuth in the host/network firewall.

Compose must support `configs.content` and healthcheck `start_interval`
(Docker Compose 2.23.1 or newer). Named volumes persist the data. Back up the
volumes and environment together; preserve secrets across updates.

Validation: all 26 services pass `docker compose config`; the resolved Caddy
configuration passes `caddy validate`. The environment generator was checked for
independent secrets, the relay key length, and a matching VAPID key pair. Fluxer
has not been started or deployed by this repository change.

## Environment values

The hostname, HTTPS origin, and ports are already set in `.env.example`.
Check `FLUXER_EDGE_BIND` and the notification email for your deployment.
The generator supplies all required application secrets:

- `POSTGRES_PASSWORD`, `MEILI_MASTER_KEY`, `FLUXER_S3_SECRET_KEY`.
- `FLUXER_SUDO_MODE_SECRET`, `FLUXER_CONNECTION_INITIATION_SECRET`,
  `FLUXER_PROFILE_PSEUDONYM_SECRET`, `FLUXER_GATEWAY_RPC_AUTH_TOKEN`,
  `FLUXER_ERLANG_COOKIE`, `FLUXER_MEDIA_PROXY_SECRET_KEY`.
- `FLUXER_MEDIA_PROXY_UPLOAD_RELAY_SECRET_BASE64` (32 random bytes, Base64).
- `FLUXER_ADMIN_SECRET_KEY_BASE`, `FLUXER_ADMIN_OAUTH_CLIENT_SECRET`.
- `FLUXER_VAPID_PUBLIC_KEY` and `FLUXER_VAPID_PRIVATE_KEY` (a matching P-256 VAPID
  pair, Base64url), and `LIVEKIT_API_SECRET`.

`FLUXER_ADMIN_OAUTH_CLIENT_SECRET` is an internal Fluxer secret. It is separate
from the Kanidm client secret. Native SSO provider settings are stored through
Fluxer's admin interface, rather than made-up OIDC environment variables.
`upstream.env.example` contains the full upstream tuning reference; its proxy
overlay/COMPOSE_FILE instructions do not apply to this combined compose file.

Email is disabled by default. Enable SMTP only after setting a real sender and
server credentials. Voice/video needs client reachability to the Docker host on
TCP 7881 and UDP 7882, in addition to the proxied `/livekit` signaling endpoint.
For LAN-only voice, set `FLUXER_LIVEKIT_NODE_IP` to the host's LAN address and
`FLUXER_LIVEKIT_USE_EXTERNAL_IP=false` if public-IP discovery is unsuitable.

## Kanidm and native SSO

- Client ID: `fluxer` (confidential client, PKCE enabled).
- Issuer: `https://idm.northlake.dev/oauth2/openid/fluxer`.
- Redirect URI: `https://chat.home.northlake.dev/auth/sso/callback`.
- Landing URL: `https://chat.home.northlake.dev/`.
- Requested scopes: `openid profile email groups`.
- Client secret: retrieve privately with
  `kanidm system oauth2 show-basic-secret fluxer -D idm_admin`, then enter it in
  Fluxer's SSO settings. Do not put it in Git.

The group hierarchy grants Northlake members access to Fluxer:

```text
northlake_users contains northlake_family_users, northlake_guest_users, northlake_admins
fluxer_users contains northlake_users and fluxer_admins
fluxer_admins contains northlake_admins
```

`kanidm group add-members CONTAINER MEMBER` puts the second group inside the
first. Therefore `add-members fluxer_admins northlake_admins` grants existing
Northlake administrators Fluxer membership. The OAuth scope map grants
`openid profile email groups` to both `fluxer_users` and `fluxer_admins`.
Fluxer's current SSO implementation does not automatically promote users to
instance administrators from a Kanidm group; assign instance admin privileges
in Fluxer after login. Group nesting and login authorization still apply.
