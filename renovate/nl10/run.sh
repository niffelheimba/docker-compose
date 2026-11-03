#!/bin/sh
# The official CLI is a one-shot job. Keep a single, non-overlapping hourly loop.
set -u
child=''
stop() {
  if [ -n "$child" ]; then
    kill -TERM "$child" 2>/dev/null || true
    wait "$child" 2>/dev/null || true
  fi
  exit 0
}
trap stop TERM INT
case "${RENOVATE_INTERVAL_SECONDS:-3600}" in
  ''|*[!0-9]*) echo 'RENOVATE_INTERVAL_SECONDS must be a positive integer' >&2; exit 1 ;;
esac
if [ "${RENOVATE_INTERVAL_SECONDS:-3600}" -lt 60 ]; then
  echo 'RENOVATE_INTERVAL_SECONDS must be at least 60' >&2
  exit 1
fi
while :; do
  if [ -s /run/secrets/renovate_token ]; then
    printf '%s Renovate update check starting\n' "$(date -u +%FT%TZ)"
    renovate &
    child=$!
    wait "$child"
    status=$?
    child=''
    printf '%s Renovate check finished; exit=%s\n' "$(date -u +%FT%TZ)" "$status"
  else
    printf '%s WAITING FOR CREDENTIAL: populate the host token file; no registry/repository update job has run.\n' "$(date -u +%FT%TZ)"
  fi
  sleep "${RENOVATE_INTERVAL_SECONDS:-3600}" &
  child=$!
  wait "$child"
  child=''
done
