#!/usr/bin/env bash
# Run the hardware benchmark from your laptop, end-to-end, against real Hetzner.
# The GitHub Action is a thin wrapper around run-matrix-entry.sh; this script
# is the developer-facing equivalent — same orchestration, run from anywhere.
#
# Required env:
#   HCLOUD_TOKEN          Hetzner Cloud API token
#   BENCHMARK_SSH_KEY     Path to the private key matching the Hetzner-side
#                         SSH key named 'benchmark-key'.
#
# Common usage:
#   run-local.sh                        # full matrix (4 tiers x 2 modes)
#   run-local.sh --smoke                # 1 tier, 1 mode, short steps (~5 min)
#   run-local.sh --tier ccx13 --mode sqlite
#   run-local.sh --dry-run              # validate env + print plan, no provisioning
#   run-local.sh --commit               # drop the "-local" suffix on the output dir
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

TIERS="ccx13,ccx23,ccx33,ccx43"
MODES="sqlite,pgch"
DURATION="30m"
SMOKE=0
DRY_RUN=0
COMMIT=0

usage() {
    sed -n '1,16p' "$0"; exit 2
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --tier)    TIERS="$2"; shift 2 ;;
        --mode)    MODES="$2"; shift 2 ;;
        --duration) DURATION="$2"; shift 2 ;;
        --smoke)   SMOKE=1; shift ;;
        --dry-run) DRY_RUN=1; shift ;;
        --commit)  COMMIT=1; shift ;;
        -h|--help) usage ;;
        *) echo "unknown flag: $1" >&2; usage ;;
    esac
done

if [[ "${SMOKE}" -eq 1 ]]; then
    TIERS="ccx13"
    MODES="sqlite"
    DURATION="3m"
    echo "smoke mode: tier=ccx13 mode=sqlite duration=${DURATION} stepped 100->400 eps" >&2
fi

# Preflight first — fail fast on missing tooling.
"${SCRIPT_DIR}/preflight.sh"

date_tag="$(date -u +%Y-%m-%d)"
if [[ "${COMMIT}" -eq 1 ]]; then
    OUT_DIR="${REPO_ROOT}/benchmarks/results/${date_tag}"
else
    OUT_DIR="${REPO_ROOT}/benchmarks/results/${date_tag}-local"
fi
mkdir -p "${OUT_DIR}"
echo "results dir: ${OUT_DIR}" >&2

# Plan: explode tiers x modes.
plan=()
IFS=',' read -ra TIER_ARR <<<"${TIERS}"
IFS=',' read -ra MODE_ARR <<<"${MODES}"
for t in "${TIER_ARR[@]}"; do
    for m in "${MODE_ARR[@]}"; do
        plan+=("${t}|${m}")
    done
done

echo "plan (${#plan[@]} entries):" >&2
for e in "${plan[@]}"; do
    echo "  - ${e//|/ x }" >&2
done

if [[ "${DRY_RUN}" -eq 1 ]]; then
    echo "dry-run: would call run-matrix-entry.sh ${#plan[@]} time(s) with duration=${DURATION}; no servers will be created." >&2
    exit 0
fi

# Sequential execution. Local runs prioritize simplicity and low concurrent
# Hetzner spend over wall-clock speed; the GH workflow parallelizes with
# strategy.matrix instead.
smoke_arg=""
[[ "${SMOKE}" -eq 1 ]] && smoke_arg="smoke"

failures=()
for e in "${plan[@]}"; do
    tier="${e%%|*}"; mode="${e##*|}"
    if ! "${SCRIPT_DIR}/run-matrix-entry.sh" "${tier}" "${mode}" "${DURATION}" "${OUT_DIR}" "${smoke_arg}"; then
        echo "FAIL: ${tier}/${mode}" >&2
        failures+=("${tier}/${mode}")
    fi
done

# Render charts once everything (or at least something) is done.
if compgen -G "${OUT_DIR}/*.json" >/dev/null; then
    echo "rendering charts" >&2
    python3 "${SCRIPT_DIR}/chart.py" "${OUT_DIR}"
else
    echo "no JSON files were produced; skipping chart render" >&2
fi

if [[ ${#failures[@]} -gt 0 ]]; then
    echo "FINISHED with failures: ${failures[*]}" >&2
    exit 1
fi
echo "FINISHED. Results in ${OUT_DIR}" >&2
