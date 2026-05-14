#!/usr/bin/env bash
# Preflight checks for the benchmark tooling. Fails loud and early so we don't
# half-provision a Hetzner box and then discover matplotlib is missing.
#
# Required env vars: HCLOUD_TOKEN, BENCHMARK_SSH_KEY (path to private key file)
# Required commands: hcloud, jq, python3, ssh, scp, curl, go
# Required python:   matplotlib (in whatever python3 we'll invoke chart.py with)
# Required Hetzner:  an SSH key uploaded as "benchmark-key"

set -euo pipefail

fail=0
warn() { printf 'WARN: %s\n' "$*" >&2; }
err()  { printf 'FAIL: %s\n' "$*" >&2; fail=1; }

# --- Commands on PATH -------------------------------------------------------

for cmd in hcloud jq python3 ssh scp curl go; do
    if ! command -v "${cmd}" >/dev/null 2>&1; then
        err "missing command: ${cmd}"
    fi
done

# --- Python: matplotlib for chart rendering ---------------------------------

if command -v python3 >/dev/null 2>&1; then
    if ! python3 -c "import matplotlib" >/dev/null 2>&1; then
        err "python3 cannot import matplotlib. Install in a venv: python3 -m venv .venv && .venv/bin/pip install matplotlib; then export PATH=\$PWD/.venv/bin:\$PATH"
    fi
fi

# --- Hetzner credentials ----------------------------------------------------

if [[ -z "${HCLOUD_TOKEN:-}" ]]; then
    err "HCLOUD_TOKEN not set. Export your Hetzner Cloud API token."
else
    if ! hcloud server-type list >/dev/null 2>&1; then
        err "HCLOUD_TOKEN is set but 'hcloud server-type list' failed — token invalid or revoked?"
    fi
fi

# --- SSH key ---------------------------------------------------------------

if [[ -z "${BENCHMARK_SSH_KEY:-}" ]]; then
    warn "BENCHMARK_SSH_KEY not set; will skip key-on-disk checks. Hetzner-side key 'benchmark-key' is still required."
else
    if [[ ! -f "${BENCHMARK_SSH_KEY}" ]]; then
        err "BENCHMARK_SSH_KEY=${BENCHMARK_SSH_KEY} does not exist"
    elif [[ "$(stat -f '%Lp' "${BENCHMARK_SSH_KEY}" 2>/dev/null || stat -c '%a' "${BENCHMARK_SSH_KEY}" 2>/dev/null)" != "600" ]]; then
        warn "BENCHMARK_SSH_KEY permissions are not 600; ssh may refuse to use it"
    fi
fi

if command -v hcloud >/dev/null 2>&1 && [[ -n "${HCLOUD_TOKEN:-}" ]]; then
    if ! hcloud ssh-key describe benchmark-key >/dev/null 2>&1; then
        err "no SSH key named 'benchmark-key' in your Hetzner project. Upload the matching public key via: hcloud ssh-key create --name benchmark-key --public-key-from-file ~/.ssh/hetzner_benchmark.pub"
    fi
fi

if [[ "${fail}" -ne 0 ]]; then
    echo >&2
    echo "preflight failed. fix the issues above before running." >&2
    exit 1
fi
echo "preflight: OK" >&2
