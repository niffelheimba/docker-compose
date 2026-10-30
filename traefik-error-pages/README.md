# Northlake error pages

A small OAuth2 Proxy-inspired page: light grey background, white card, large
status code, short explanation and Northlake footer. One embedded HTML template;
no JavaScript, fonts, images, CDN, tracking or external dependencies. Edit
`error.html` and rebuild to change the design.

The server uses only Go's standard library. The final `scratch` image contains
one statically linked executable: no shell, OS packages, CA bundle or build tools.
It needs no outbound connections, mounted files, credentials or writable storage.
Responses are pre-rendered at startup and preserve the actual 4xx/5xx HTTP status.
The build stage needs access to the pinned Go base image; rebuild with a patched
Go version when security updates are available. The image size must be measured
after building; no container size or runtime resource usage has been verified yet.
Cross-compilation with the local Go 1.27.0 toolchain succeeded for Linux AMD64
(9,076,860 bytes) and ARM64 (8,454,268 bytes). The Docker build uses Go 1.27.3;
these local binary sizes are indicative, not a measurement of that final image.

## Compose integration

Both Traefik Compose files build the same source from `../traefik-error-pages`.
Keep this sibling folder alongside each stack on its Docker build host (or build
and distribute the image separately before deployment). An Arcane sync that only
copies the stack folder must also copy the shared source folder; do not deploy a
stack with a missing build context. Builds are native to the host architecture,
including ARM64 on NL00 and AMD64 on NL10. No image has been published.

Only Traefik and `error-pages` join each stack's new `internal` bridge network.
The renderer has no published ports, host mounts, Docker socket, Linux
capabilities or access to application networks. It runs as UID/GID 65532 with
`no-new-privileges`, a read-only root filesystem, 64 MiB memory, 32 PIDs and a
0.25 CPU limit. These limits are initial values to verify on each actual host.
The 32 MiB Go memory setting is a soft GC target, not a process memory limit.

The server sets restrictive CSP, no-store, no-referrer, nosniff and frame-denial
headers. It never includes the original host, URL, headers or backend error text
in its HTML and does not record request logs. The error middleware explicitly
forwards no original request headers (`errorRequestHeaders: []`, supported by the
checked-in Traefik v3.7.1); cookies and authorization are not needed here.

## Routes and Manager ownership

The two `config/northlake-errors.yml` files are **seed/import files**. Live dynamic
configuration belongs to Manager (NL00 directly; NL10 through its agent).
Deploying Compose does NOT import these files. Never copy the complete Git
`config/` tree over a running Manager-owned configuration directory.

Each seed defines:

- `northlake-error-pages`: HTTP backend at `http://error-pages:8080`.
- `northlake-errors`: replaces 403 and 500–599 responses on opted-in routes.
- `northlake-errors-with-404`: optional variant also replacing application 404s.
- `northlake-not-found`: priority-1 fallback for unmatched Northlake hostnames.
  It replaces the request path with `/errors/404` and returns a genuine 404.
  This also prevents requests to `/healthz` or `/errors/403` on unknown hosts
  from reaching those renderer endpoints directly.

NL00's fallback covers the apex, single-label `*.northlake.dev`, the cloud apex
and single-label `*.cloud.northlake.dev`. NL10 covers the same base zone and
single-label `*.home.northlake.dev`. Unrelated domains are excluded. Existing
routes with higher priority win, so existing applications retain their own
missing-path behavior unless the optional 404 middleware is attached.

The existing seed routes prepend `northlake-errors@file` before IP restrictions:

- NL00: status-admin, semaphore, arcane.
- NL10: homeassistant-home, watch, pve1, pve2, truenas, amp, grafana.

Apply the same attachment in Manager to enable these on live routes. Do not
replace existing middleware lists. Other browser routes can opt in after checking
whether they return error bodies that their clients rely on. Frigate's existing
401/403 OAuth flow is left alone. S3, Loki ingestion and other machine-facing
routes, including OpenBao, do not receive the HTML middleware. An attachment is
per route: API calls sharing an opted-in browser route also get HTML on matched
errors, so check mixed-use applications before enabling it live. Authentication and allowlists stay in
place; formatting a denial does not grant access.

## DNS and HTTPS prerequisites

An incorrect hostname such as `asdasd.home.northlake.dev` must resolve to NL10's
Traefik address. Traefik cannot render a page for an NXDOMAIN response.
Configure/verify wildcard DNS in the appropriate internal DNS zone; don't expose
private services or change public DNS merely to obtain the fallback.

The fallback routers specify explicit ACME domains for `*.northlake.dev` plus
`*.home.northlake.dev` on NL10 or `*.cloud.northlake.dev` on NL00. Importing these
routers requests certificates through the existing Cloudflare DNS challenge.
Confirm credentials permit those zones and inspect existing wildcard certificates
first to avoid needless issuance or rate limits. A wildcard covers one label:
`*.northlake.dev` does not cover `asdasd.home.northlake.dev`. The new fallback does
not cover deeper labels such as `typo.actualbudget.home.northlake.dev`.

TLS must succeed before a browser can display the HTML. Do not work around failed
issuance by disabling certificate checks. No DNS/certificate issuance has been
performed by this implementation session.

## Deployment and validation

1. Back up/export each instance's current dynamic configuration in Manager.
   Record current route middleware ordering. Check the build context is present.
2. Validate each Compose file with its deployed environment. Build and start only
   `error-pages` first: `docker compose up -d --build error-pages`.
3. Recreate only Traefik to attach its new network:
   `docker compose up -d --no-deps traefik`. This can briefly interrupt requests;
   do one host at a time. Keep Manager and the NL10 agent running.
4. Import/create the definitions from that host's `config/northlake-errors.yml`
   using Manager's configuration editor, retaining `http.routers`,
   `http.middlewares` and `http.services`. If using the middleware-only form,
   paste only the individual middleware body. Inspect Traefik's API/logs for
   configuration errors and confirm certificate issuance before testing HTTPS.
5. Attach `northlake-errors@file` first on the selected browser routes, ahead of
   the IP restriction. Preserve all other middlewares. Do not attach both variants
   to the same route. Do not put the errors middleware on the renderer service
   itself or its fallback router (that router already serves the 404 directly).
6. Verify an existing allowed route still works; a disallowed client gets the
   styled 403 and remains blocked; a controlled unavailable browser backend gets
   the matching 502/503/504; Frigate login and machine-facing APIs are unchanged.
7. Verify `https://asdasd.home.northlake.dev/`, an arbitrary path and `/healthz`
   on that unknown hostname all return the styled 404 with a trusted certificate.
   Test an unknown cloud hostname on NL00 as well. Use `curl --resolve` when
   checking routing before DNS, but keep normal TLS verification enabled.
8. Check container restart count, memory/CPU usage and lack of host listeners.
   Verify both actual HTTP status and body; the page must not turn errors into 200.

Rollback: restore the prior route middleware lists and remove the fallback router
and imported error definitions in Manager. Restore the previous Compose file and
recreate Traefik to detach the added network, then stop/remove only `error-pages`.
Keep existing ACME storage intact. Do not use `down` on the complete proxy stack
or delete certificate volumes.

## Local checks

From this folder: `go test ./...` and `go vet ./...`.
The tests check status/body agreement, no-store/CSP headers, HEAD/health/method
handling, and rejection of malformed status paths without reflecting untrusted
request data. Container build and end-to-end proxy checks require Docker.

For a loopback-only local preview, set `LISTEN_ADDR=127.0.0.1:18080` before running
the native executable and open `/errors/403` or `/errors/404`. Default container
listening address is `:8080`.

References: [Traefik Errors middleware](https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/errorpages/),
[OAuth2 Proxy's error template](https://github.com/oauth2-proxy/oauth2-proxy/blob/master/pkg/app/pagewriter/error.html).
