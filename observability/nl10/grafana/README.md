# Grafana improvements — 2026-09-09

Applied live through the Grafana MCP:
- Docker & Host Logs: repaired LogQL, meaningful zero count, error causes and camera names, cleaner original logs, all-host default, and Traefik link.
- Traefik Statistics: operational error/warning counts, canceled ForwardAuth requests, middleware breakdown, and logs. HTTP requests, status codes, router rates, and p95 latency are ready but await access-log ingestion.
- Both existing alerts: Loki excluded by container and service labels, with Grafana's pre-existing exclusion retained. Corrected invalid negative regex syntax and word-boundary escaping. Preserved Discord routing.
- Warning threshold: more than 4 matching lines per host/container/component/cause over 5 minutes, held for 2 minutes. Sparse log matches resolve as OK. Actual query execution failures still surface as errors.
- Critical rule: immediate fatal/panic/OOM/segmentation-fault/crash-loop patterns. A camera FFmpeg worker crash now goes through the warning rule.
- Added readable causes, counts, affected camera/component, and filtered dashboard links to alert annotations.
- Original dashboards/rules are in grafana-backups/2026-09-09.

## Remaining deployment: HTTP access logs

Only Loki is configured in Grafana; Prometheus metrics are not available. Live Loki currently has Docker and journal streams, without Traefik JSON request access logs.

The two Traefik Compose files in this checkout now omit:
```
--accesslog.filepath=/var/log/traefik/access.log
```
Traefik then emits JSON access logs to stdout, which the existing Alloy Docker source collects. These Compose edits have NOT been deployed. Apply the same one-line change to the actual active Traefik stacks in Arcane if their configuration differs from this checkout; do not replace a live stack with an older local Compose definition. Redeploy only the Traefik service in each relevant stack. This stops new writes to the old access.log file; existing files are retained.

After deployment, make a normal request through Traefik and check:
```logql
{source="docker",container="traefik"} | json | __error__="" | DownstreamStatus!=""
```
The HTTP dashboard panels will populate from that point onward. No historical request statistics can be recovered from stdout; collecting the existing files with Alloy is an alternative if historical access logs are needed.

## Frigate findings

Recent logs showed Garage_Mobile RTSP connection timeouts, invalid camera input, and watchdog messages reporting no valid recording segments for 120 seconds. Earlier logs showed Living_Room_Mobile FFmpeg worker crashes. These affect camera capture/recording while the Frigate container can remain running. Frigate also emits all FFmpeg output at error severity, including timestamp warnings; only the observed timestamp-warning patterns were excluded from error views, and they remain visible in the full logs.

## Verification

Both alert rules evaluated with health=ok after the changes. All panel query families were exercised against Loki. Operational Traefik counters returned live data; HTTP queries parsed successfully but cannot be validated against real request data until collection is deployed. Dashboard rendering was inspected in Grafana.

References:
- https://docs.frigate.video/configuration/advanced/
- https://grafana.com/docs/loki/latest/query/log_queries/
- https://doc.traefik.io/traefik/observability/access-logs/

