#!/usr/bin/env bash
# Reset the bench database on a managed ClickHouse to an empty state so each
# matrix entry starts from a clean baseline. Reads from the environment:
#   CLICKHOUSE_SERVER       host:nativePort (e.g. cluster.clickhouse.cloud:9440)
#   CLICKHOUSE_USERNAME
#   CLICKHOUSE_PASSWORD
#   CLICKHOUSE_DATABASE     default: traceway
#   BENCH_CH_HTTPS_PORT     default: 8443 (CH Cloud uses 8443; some hosts 8123)
#
# The HTTPS interface lives on the same hostname as the native port, just on a
# different port. We strip the native port off CLICKHOUSE_SERVER to get the
# host, then talk to it over HTTPS.
set -euo pipefail

: "${CLICKHOUSE_SERVER:?required for managed-ch mode}"
: "${CLICKHOUSE_USERNAME:?required for managed-ch mode}"
: "${CLICKHOUSE_PASSWORD:?required for managed-ch mode}"
DB="${CLICKHOUSE_DATABASE:-traceway}"

CH_HOST="${CLICKHOUSE_SERVER%:*}"
CH_HTTPS_PORT="${BENCH_CH_HTTPS_PORT:-8443}"
URL="https://${CH_HOST}:${CH_HTTPS_PORT}/"

# Managed CH (e.g. ClickHouse Cloud) auto-idles when unused. The first request
# wakes it up but can hang or 503 for up to a few minutes while the service
# spins back up. Poll with short per-attempt timeouts until we get a healthy
# response, then run the actual reset.
WAKE_TOTAL_TIMEOUT="${BENCH_CH_WAKE_TIMEOUT:-600}"   # seconds, total budget
WAKE_ATTEMPT_TIMEOUT="${BENCH_CH_WAKE_ATTEMPT_TIMEOUT:-20}"
WAKE_SLEEP="${BENCH_CH_WAKE_SLEEP:-10}"

echo "waking managed CH at ${CH_HOST}:${CH_HTTPS_PORT} (up to ${WAKE_TOTAL_TIMEOUT}s)" >&2
deadline=$(( $(date +%s) + WAKE_TOTAL_TIMEOUT ))
attempt=0
while :; do
    attempt=$(( attempt + 1 ))
    if curl -fsS --max-time "${WAKE_ATTEMPT_TIMEOUT}" \
        -u "${CLICKHOUSE_USERNAME}:${CLICKHOUSE_PASSWORD}" \
        "${URL}" --data-binary "SELECT 1" >/dev/null 2>&1; then
        echo "managed CH responsive after ${attempt} attempt(s)" >&2
        break
    fi
    if [[ $(date +%s) -ge ${deadline} ]]; then
        echo "managed CH did not wake within ${WAKE_TOTAL_TIMEOUT}s after ${attempt} attempts" >&2
        exit 1
    fi
    echo "  attempt ${attempt} not ready yet, sleeping ${WAKE_SLEEP}s" >&2
    sleep "${WAKE_SLEEP}"
done

echo "resetting managed CH database '${DB}' on ${CH_HOST}:${CH_HTTPS_PORT}" >&2
curl -fsSL --max-time 60 -u "${CLICKHOUSE_USERNAME}:${CLICKHOUSE_PASSWORD}" \
    "${URL}" --data-binary "DROP DATABASE IF EXISTS \`${DB}\`"
curl -fsSL --max-time 60 -u "${CLICKHOUSE_USERNAME}:${CLICKHOUSE_PASSWORD}" \
    "${URL}" --data-binary "CREATE DATABASE \`${DB}\`"
echo "managed CH '${DB}' is empty" >&2
