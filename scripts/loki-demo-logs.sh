#!/bin/sh
set -eu

LOKI_URL="${LOKI_URL:-http://loki:3100}"
LOKI_DEMO_INTERVAL="${LOKI_DEMO_INTERVAL:-5}"
sequence=0

printf 'waiting for Loki readiness at %s/ready\n' "$LOKI_URL"
until curl -fsS "$LOKI_URL/ready" >/dev/null; do
  sleep 1
done

printf 'sending synthetic alioth-demo logs every %ss\n' "$LOKI_DEMO_INTERVAL"
while true; do
  epoch_seconds="$(date +%s)"
  timestamp_ns="$((epoch_seconds * 1000000000 + sequence))"
  payload="$(printf '{"streams":[{"stream":{"app":"api","env":"local","level":"error","source":"alioth-demo"},"values":[["%s","synthetic error: request timeout id=%s"]]},{"stream":{"app":"web","env":"local","level":"info","source":"alioth-demo"},"values":[["%s","synthetic info: request handled id=%s"]]}]}' \
    "$timestamp_ns" "$sequence" "$timestamp_ns" "$sequence")"

  curl -fsS --retry 3 --retry-connrefused \
    -H 'Content-Type: application/json' \
    --data-binary "$payload" \
    "$LOKI_URL/loki/api/v1/push" >/dev/null

  sequence="$((sequence + 1))"
  sleep "$LOKI_DEMO_INTERVAL"
done
