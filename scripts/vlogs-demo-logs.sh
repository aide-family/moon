#!/bin/sh
set -eu

VLOGS_URL="${VLOGS_URL:-http://victorialogs:9428}"
VLOGS_DEMO_INTERVAL="${VLOGS_DEMO_INTERVAL:-5}"
sequence=0

printf 'waiting for VictoriaLogs health at %s/health\n' "$VLOGS_URL"
until curl -fsS "$VLOGS_URL/health" >/dev/null; do
  sleep 1
done

printf 'sending synthetic alioth-demo logs every %ss\n' "$VLOGS_DEMO_INTERVAL"
while true; do
  api_payload="$(printf '{"_msg":"synthetic error: request timeout id=%s","app":"api","env":"local","level":"error","source":"alioth-demo"}' "$sequence")"
  web_payload="$(printf '{"_msg":"synthetic info: request handled id=%s","app":"web","env":"local","level":"info","source":"alioth-demo"}' "$sequence")"

  # _stream_fields 把 app/env/level/source 设成流字段，查询时可用 _stream:{...} 过滤。
  curl -fsS --retry 3 --retry-connrefused \
    -H 'Content-Type: application/stream+json' \
    --data-binary "$api_payload" \
    "$VLOGS_URL/insert/jsonline?_stream_fields=app,env,level,source" >/dev/null
  curl -fsS --retry 3 --retry-connrefused \
    -H 'Content-Type: application/stream+json' \
    --data-binary "$web_payload" \
    "$VLOGS_URL/insert/jsonline?_stream_fields=app,env,level,source" >/dev/null

  sequence="$((sequence + 1))"
  sleep "$VLOGS_DEMO_INTERVAL"
done
