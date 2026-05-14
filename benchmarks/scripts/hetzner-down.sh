#!/usr/bin/env bash
# Tear down servers + network for a given run-id. Idempotent: missing resources
# are not errors. Designed to be safe to call from a bash trap, including when
# triggered mid-provisioning.
#
# Usage: hetzner-down.sh <run-id>
set -euo pipefail

if [[ $# -lt 1 ]]; then
    echo "usage: $0 <run-id>" >&2
    exit 2
fi

RUN_ID="$1"
SUT_NAME="bench-sut-${RUN_ID}"
LOADGEN_NAME="bench-loadgen-${RUN_ID}"
NET_NAME="bench-net-${RUN_ID}"

drop_server() {
    local name="$1"
    if hcloud server describe "${name}" >/dev/null 2>&1; then
        echo "deleting server ${name}" >&2
        hcloud server delete "${name}" >/dev/null || echo "warn: delete ${name} returned non-zero" >&2
    fi
}

drop_server "${SUT_NAME}"
drop_server "${LOADGEN_NAME}"

if hcloud network describe "${NET_NAME}" >/dev/null 2>&1; then
    echo "deleting network ${NET_NAME}" >&2
    # Network delete requires no attached servers; we just deleted them above.
    hcloud network delete "${NET_NAME}" >/dev/null || echo "warn: delete ${NET_NAME} returned non-zero" >&2
fi

echo "teardown of ${RUN_ID}: done" >&2
