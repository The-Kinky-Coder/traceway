#!/usr/bin/env bash
# Bring up the Traceway backend on a freshly-provisioned Hetzner box.
#
# 1. Wait for SSH.
# 2. Install Docker if missing.
# 3. rsync the repo to /opt/traceway (sans heavy junk).
# 4. `docker compose -f benchmarks/compose/docker-compose.<mode>.yml up -d --build`
# 5. Poll /health until 200.
#
# Usage: sut-bootstrap.sh <sut-public-ip> <mode>
#   <mode>  sqlite | pgch
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
# shellcheck source=_ssh.sh
source "${SCRIPT_DIR}/_ssh.sh"

if [[ $# -lt 2 ]]; then
    echo "usage: $0 <sut-public-ip> <mode>" >&2
    exit 2
fi
SUT_IP="$1"
MODE="$2"

case "${MODE}" in
    sqlite|pgch) ;;
    *) echo "mode must be sqlite or pgch, got: ${MODE}" >&2; exit 2 ;;
esac

echo "waiting for ssh on ${SUT_IP}" >&2
wait_for_ssh "${SUT_IP}"

echo "installing docker on ${SUT_IP}" >&2
bench_ssh "${SUT_IP}" 'bash -s' <<'REMOTE'
set -euo pipefail
if ! command -v docker >/dev/null 2>&1; then
    curl -fsSL https://get.docker.com | sh
fi
# compose plugin lives at /usr/libexec/docker/cli-plugins after get.docker.com.
docker compose version >/dev/null || { echo "docker compose plugin missing" >&2; exit 1; }
mkdir -p /opt/traceway
REMOTE

echo "rsyncing source to ${SUT_IP}:/opt/traceway" >&2
bench_rsync \
    --exclude '.git' \
    --exclude 'node_modules' \
    --exclude 'frontend/build' \
    --exclude 'frontend/.svelte-kit' \
    --exclude 'benchmarks/loadgen/loadgen' \
    --exclude 'benchmarks/results' \
    --exclude 'traceway.db*' \
    --exclude 'traceway_telemetry.db*' \
    --exclude 'backend/storage' \
    "${REPO_ROOT}/" "root@${SUT_IP}:/opt/traceway/"

echo "bringing up compose stack (mode=${MODE}) on ${SUT_IP}" >&2
bench_ssh "${SUT_IP}" "cd /opt/traceway && BENCH_PORT=80 docker compose -f benchmarks/compose/docker-compose.${MODE}.yml up -d --build"

echo "polling /health on ${SUT_IP}" >&2
deadline=$(( $(date +%s) + 600 ))   # cold compose build can hit 10 min on small tiers
while [[ $(date +%s) -lt ${deadline} ]]; do
    if curl -sf --max-time 5 "http://${SUT_IP}/health" >/dev/null 2>&1; then
        echo "SUT ${SUT_IP} healthy (mode=${MODE})" >&2
        exit 0
    fi
    sleep 5
done
echo "SUT ${SUT_IP} never reported healthy" >&2
bench_ssh "${SUT_IP}" "cd /opt/traceway && docker compose -f benchmarks/compose/docker-compose.${MODE}.yml logs --tail=80" >&2 || true
exit 1
