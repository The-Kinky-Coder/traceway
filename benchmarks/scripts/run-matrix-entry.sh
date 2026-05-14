#!/usr/bin/env bash
# Run ONE (tier, mode) cycle end-to-end: provision Hetzner -> bootstrap SUT ->
# run loadgen -> fetch JSON -> tear down. Called by both run-local.sh and the
# GitHub workflow; treat this as the single source of truth for matrix-entry
# orchestration.
#
# Usage: run-matrix-entry.sh <tier> <mode> <duration> <out-dir> [smoke]
#   <tier>      ccx13 | ccx23 | ccx33 | ccx43
#   <mode>      sqlite | pgch
#   <duration>  Loadgen total runtime, e.g. 30m, 3m
#   <out-dir>   Directory to write <tier>-<mode>.json into
#   [smoke]     "smoke" to enable short-step overrides (--initial-eps 100
#               --max-eps 400 --step-duration 1m). Optional.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

if [[ $# -lt 4 ]]; then
    echo "usage: $0 <tier> <mode> <duration> <out-dir> [smoke]" >&2
    exit 2
fi
TIER="$1"; MODE="$2"; DURATION="$3"; OUT_DIR="$4"; SMOKE="${5:-}"
LOCATION="${BENCH_LOCATION:-nbg1}"

RUN_ID="$(date -u +%Y%m%d-%H%M%S)-${TIER}-${MODE}-$RANDOM"
echo "=== run-matrix-entry tier=${TIER} mode=${MODE} duration=${DURATION} run_id=${RUN_ID} ===" >&2

mkdir -p "${OUT_DIR}"

# Always tear down — even on failure, even on Ctrl-C. The trap is set BEFORE
# any hcloud create call so a failure mid-provision still cleans up.
cleanup() {
    local rc=$?
    echo "--- teardown for ${RUN_ID} (exit=${rc}) ---" >&2
    "${SCRIPT_DIR}/hetzner-down.sh" "${RUN_ID}" || true
    exit "${rc}"
}
trap cleanup EXIT INT TERM

# 1. Provision.
INFRA_JSON=$("${SCRIPT_DIR}/hetzner-up.sh" "${TIER}" "${RUN_ID}" "${LOCATION}")
echo "infra: ${INFRA_JSON}" >&2
SUT_PUBLIC_IP=$(printf '%s' "${INFRA_JSON}" | jq -r '.sutPublicIp')
SUT_PRIVATE_IP=$(printf '%s' "${INFRA_JSON}" | jq -r '.sutPrivateIp')
LOADGEN_PUBLIC_IP=$(printf '%s' "${INFRA_JSON}" | jq -r '.loadgenPublicIp')

# 2. Bring up the backend on the SUT.
"${SCRIPT_DIR}/sut-bootstrap.sh" "${SUT_PUBLIC_IP}" "${MODE}"

# 3. Run the loadgen, pulling JSON back into OUT_DIR.
extra_args=()
if [[ "${SMOKE}" == "smoke" ]]; then
    extra_args+=( --initial-eps 100 --max-eps 400 --step-duration 1m )
fi

OUT_PATH="${OUT_DIR}/${TIER}-${MODE}.json"
"${SCRIPT_DIR}/loadgen-bootstrap.sh" \
    "${LOADGEN_PUBLIC_IP}" \
    "${SUT_PRIVATE_IP}" \
    "${SUT_PUBLIC_IP}" \
    "${DURATION}" \
    "${TIER}" \
    "${MODE}" \
    "${OUT_PATH}" \
    "${extra_args[@]}"

# Trap handles teardown — no explicit call needed.
echo "matrix entry ${TIER}-${MODE} complete -> ${OUT_PATH}" >&2
