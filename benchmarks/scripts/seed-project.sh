#!/usr/bin/env bash
# Register a fresh user + org + project against a Traceway backend.
#
# Emits a single JSON line on stdout (everything else on stderr) so callers can
# capture cleanly:
#   eval "$(seed-project.sh http://localhost:8087 | jq -r '@sh "JWT=\(.jwt) PROJECT_TOKEN=\(.projectToken) PROJECT_ID=\(.projectId)"')"
#
# Output schema:
#   { "jwt": "<JWT for dashboard read endpoints>",
#     "projectToken": "<bearer token for /api/report + OTLP ingestion>",
#     "projectId":    "<UUID; pass as ?projectId= on read endpoints>" }
#
# Usage: seed-project.sh <base-url>
#   <base-url>  Backend root, e.g. http://localhost:8087 or http://10.0.0.2
set -euo pipefail

if [[ $# -lt 1 ]]; then
    echo "usage: $0 <base-url>" >&2
    exit 2
fi

BASE_URL="${1%/}"
EMAIL="bench-$(date +%s)-$$@example.com"
PASSWORD="bench-pw-not-secret-12345"

response=$(curl -sS --fail-with-body -X POST "${BASE_URL}/api/register" \
    -H "Content-Type: application/json" \
    -d "$(cat <<JSON
{
  "email": "${EMAIL}",
  "name": "bench",
  "password": "${PASSWORD}",
  "organizationName": "bench-org",
  "timezone": "UTC",
  "projectName": "bench-project",
  "framework": "custom"
}
JSON
)")

echo "registered ${EMAIL}" >&2

out=$(printf '%s' "${response}" | jq -c '{jwt: .token, projectToken: .project.token, projectId: .project.id}')
if [[ "$(printf '%s' "${out}" | jq -r '.projectToken')" == "null" ]]; then
    echo "ERROR: no project token in response: ${response}" >&2
    exit 1
fi

printf '%s\n' "${out}"
