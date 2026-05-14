# Running benchmarks locally

Step-by-step recipe for running the loadgen against a backend on your laptop —
no Hetzner, no GitHub Actions. Useful for: iterating on the loadgen itself,
checking that a backend change didn't regress single-machine throughput, or
just understanding the tooling before spending real money on Hetzner.

> Laptop numbers are not comparable to Hetzner numbers. Use this for tooling
> validation and relative regression checks, not for the published chart.

---

## Prerequisites

Install once:

```bash
# macOS
brew install docker jq go

# Confirm Docker is running
docker info >/dev/null && echo "docker: ok"
```

You also need Python 3 with matplotlib. The cleanest path on macOS (system
Python is PEP-668 locked and refuses `pip install`):

```bash
# Pick one of:

# Option A — venv in /tmp (wiped on restart, fine for ad-hoc smoke tests)
python3 -m venv /tmp/bench-venv && /tmp/bench-venv/bin/pip install -q matplotlib

# Option B — venv inside the repo (persists across restarts; gitignored)
python3 -m venv .venv-bench && .venv-bench/bin/pip install -q matplotlib
```

The rest of this doc assumes option B, so swap `/tmp/bench-venv` for
`.venv-bench` if you went with option A.

---

## Pick a backend mode

Two compose stacks, each pinned to one of the project's existing Dockerfiles:

| Mode | Compose file | Image | What it tests |
|------|--------------|-------|---------------|
| `sqlite` | `benchmarks/compose/docker-compose.sqlite.yml` | `Dockerfile.sqlite` | Single-binary backend with embedded SQLite. Fast to bring up, lower ceiling. |
| `pgch` | `benchmarks/compose/docker-compose.pgch.yml` | `Dockerfile.minimal` + clickhouse + postgres | Full prod-shape stack. Slower first build, much higher ceiling. |

Both expose port **8087** on the host (override with `BENCH_PORT=8088 docker
compose -f ... up`).

First-time builds:
- `sqlite` mode: ~3–6 min (npm install + Go build).
- `pgch` mode: ~6–10 min (above + ClickHouse + Postgres image pulls).

Subsequent runs reuse the Docker layer cache and start in seconds.

---

## The 6-step run

These are the exact commands. Open a terminal at the repo root:

```bash
cd /Users/dusanstanojevic/Documents/workspace/traceway
```

### 1. Bring up the backend on a clean volume

```bash
# Always tear down with -v first. Reusing a volume across runs means the next
# loadgen run starts with old data, AND can hit "database disk image is
# malformed" if the previous container was killed mid-write.
docker compose -f benchmarks/compose/docker-compose.sqlite.yml down -v 2>/dev/null
docker compose -f benchmarks/compose/docker-compose.sqlite.yml up -d --build
```

For `pgch` mode swap the filename to `docker-compose.pgch.yml` here and in the
remaining steps.

### 2. Wait for `/health` to return 200

```bash
until curl -sf http://localhost:8087/health >/dev/null; do echo "waiting..."; sleep 2; done
echo "ready"
```

If this loops for more than ~3 min, something is wrong. Inspect:

```bash
docker compose -f benchmarks/compose/docker-compose.sqlite.yml logs --tail=80 traceway
```

### 3. Register a fresh user + project; capture creds

```bash
eval "$(benchmarks/scripts/seed-project.sh http://localhost:8087 \
  | jq -r '"export JWT=\(.jwt) PROJECT_TOKEN=\(.projectToken) PROJECT_ID=\(.projectId)"')"
echo "PROJECT_ID=$PROJECT_ID"
```

`seed-project.sh` hits `/api/register`. On self-hosted backends only one
organization is allowed, so this only works on a **fresh** DB — step 1's `down
-v` is what guarantees that.

### 4. Run the loadgen

Pick numbers appropriate for the mode you booted in step 1:

**SQLite (laptop):**

```bash
mkdir -p /tmp/bench-localhost && rm -f /tmp/bench-localhost/*.json
benchmarks/loadgen/loadgen \
  --target http://localhost:8087 \
  --token "$PROJECT_TOKEN" --jwt "$JWT" --project-id "$PROJECT_ID" \
  --duration 5m --step-duration 45s \
  --initial-eps 100 --max-eps 800 \
  --report-out /tmp/bench-localhost/local-sqlite.json \
  --tier local --mode sqlite
```

**pgch (laptop):**

```bash
benchmarks/loadgen/loadgen \
  --target http://localhost:8087 \
  --token "$PROJECT_TOKEN" --jwt "$JWT" --project-id "$PROJECT_ID" \
  --duration 5m --step-duration 45s \
  --initial-eps 500 --max-eps 4000 \
  --report-out /tmp/bench-localhost/local-pgch.json \
  --tier local --mode pgch
```

On stderr you'll see one line per step like:

```
step 1: target=300 actual=306 ingest_p99=92ms ingest_err=0.00% passed=true
step 2: target=600 actual=600 ingest_p99=17ms ingest_err=0.00% passed=false read endpoint "endpoints_grouped" p99 = 3268ms > 3000ms threshold
wrote /tmp/bench-localhost/local-sqlite.json: max sustainable EPS = 306 (2 step(s))
```

The run finishes when either (a) a step fails the breaking-point check, (b)
`--max-eps` is exceeded, or (c) `--duration` expires.

### 5. Render the chart + summary

```bash
.venv-bench/bin/python benchmarks/scripts/chart.py /tmp/bench-localhost/
open /tmp/bench-localhost/chart.png
open /tmp/bench-localhost/chart-detail.png
cat /tmp/bench-localhost/summary.md
```

A one-entry result renders fine; it just won't have multiple bars/lines to
compare against.

### 6. Tear down

```bash
docker compose -f benchmarks/compose/docker-compose.sqlite.yml down -v
```

The `-v` deletes the named volume so the next run is guaranteed clean.

---

## Reading the output

Each step record in the JSON looks like:

```json
{
  "step": 1,
  "targetEps":  300,            // Synthetic event units the ramp asked for. NOT requests/sec — see "What an EPS is" below.
  "actualEps":  306.6,          // What the loadgen actually pushed on the wire (same unit).
  "ingest": {                    // Latency/errors for /api/report + OTLP combined
    "p50": 3.8, "p95": 12.5, "p99": 92.3,
    "ok": 1380, "errors": 0, "errRate": 0
  },
  "reads": {                     // Per dashboard read endpoint
    "endpoints_grouped":  { "p50": 1120, "p95": 1462, "p99": 1509, "ok": 36,  ... },
    "dashboard_overview": { ... },
    "exceptions_grouped": { ... },
    "logs":               { ... },
    "session_recording":  { ... }
  },
  "passed": true,
  "failReason": ""               // populated when passed=false
}
```

**A step passes** when *all three* hold:
1. Ingest error rate ≤ 5% (5xx, timeouts).
2. Every read endpoint p99 < 3000 ms.
3. SUT was reachable for the whole step.

`maxSustainableEps` in the JSON is the `actualEps` of the **last passing
step** — that's the headline number for this configuration.

If `target` and `actual` diverge a lot on a passing step, the backend is
already saturated even though it's not erroring yet. The next step usually
fails.

If reads degrade before ingest, that's expected: aggregation reads
(`endpoints_grouped`, `dashboard_overview`) compete with write locks/IO and
will be the first to cliff on SQLite. Point lookups (`logs`,
`session_recording`, `exceptions_grouped`) typically stay fast right up until
the SUT gives up.

---

## What an EPS is (and isn't)

`EPS` in the loadgen is **synthetic event units**, not requests/sec and not
DB rows/sec. The conversion:

```
requests/sec = EPS / 10        # because tracesPerFrame = 10
```

So `--max-eps 800` actually means "push up to 80 HTTP requests/sec." The
factor-of-10 fudge keeps the rate-limiter math clean (each `/api/report`
frame carries 10 traces; the loadgen counts EPS as traces).

### What goes in each request

Each "frame" the loadgen sends is one HTTP POST, routed by `--report-share`
(default 0.7 → 70 % go to `/api/report`, 30 % to OTLP).

**`/api/report` frame** — gzipped JSON, built by `buildFrame()` in
`benchmarks/loadgen/report.go`:

| Field | Count per request | Notes |
|---|---|---|
| `traces` | **10 (fixed)** | One row per `transactions` |
| `spans` (nested in traces) | **20–50** (each trace gets 2–5) | Random duration 1–200 ms, name from `{db.query, http.call, cache.get, render, json.encode}` |
| `metrics` | **5 (fixed)** | `bench.metric.{cpu,mem,qps,lat,errs}`, random value 0–100 |
| `sessions` | 0 / 1 — 5 % with recording, 16 % session-only | Random uuid + StartedAt |
| `sessionRecordings` | 0 / 1 — 5 % chance | ~100 KB synthetic rrweb-shaped blob |
| `stackTraces` | 0 / 1 — 2 % chance | Hardcoded short stack trace |

Trace fields: random endpoint from a fixed 13-item list (`GET /api/users`,
`POST /api/orders`, `GET /api/products/:id`, …), duration 10–1000 ms, random
status code, body size 200–8200 bytes.

**OTLP frame** — protobuf, built by `sendOTLPTraces()` in
`benchmarks/loadgen/otlp.go`:

| Field | Count per request | Notes |
|---|---|---|
| `spans` | **20 (fixed, `otlpSpansPerRequest`)** | Random 16-byte traceId + 8-byte spanId, http.method=GET, status code, route from the same 13-endpoint pool |

No metrics, sessions, or exceptions on this path.

### Row growth per second at 400 EPS

At 400 EPS the loadgen fires **40 requests/sec total**: ~28 to `/api/report`,
~12 to OTLP. That fans out into:

| Table | Rows/sec |
|---|---|
| `transactions` | ~280 (28 × 10) |
| `spans` (from `/api/report`) | ~980 (28 × ~35) |
| `spans` (from OTLP) | ~240 (12 × 20) |
| `metric_points` | ~140 (28 × 5) |
| `sessions` | ~6 (28 × 21 %) |
| `session_recordings` | ~1.4 (28 × 5 %) |
| `exception_stack_traces` | ~0.56 (28 × 2 %) |

So **"400 EPS" ≈ 1,500 rows/sec** across the database, not 400. Keep this
mapping in mind when comparing the chart's headline number to your real
workload.

### Data variety

Everything is uniform random within bounded ranges — deliberately flat:

- 13 distinct endpoint paths → `GROUP BY endpoint` aggregations always
  produce ~13 groups
- 5 span names, 5 metric names
- All UUIDs are fresh per record (no hash clustering)
- Durations: uniform 10–1000 ms → percentile queries hit a smooth
  distribution
- No diurnal traffic, no power-law endpoint skew, no correlated bursts

This isolates the backend's raw throughput from traffic-shape noise. It is
**not** representative of real production telemetry; don't read the numbers
as "this is what users will experience" — they're "this is the floor the
backend can sustain on synthetic, evenly-distributed load."

---

## Tuning the run

| Flag | Default | When to change |
|------|---------|----------------|
| `--initial-eps` | 100 | Raise close to the expected ceiling so you don't waste the first few steps far below saturation. |
| `--max-eps` | 50000 | Lower it to bound runtime; raise to find the ceiling on a beefier box. |
| `--step-duration` | 2m | Shorten for smoke checks (45s), keep at 2m for real measurements. |
| `--duration` | 30m | Hard cap on total wall time. |
| `--report-share` | 0.7 | Fraction of ingest sent via `/api/report` vs OTLP. Lower it to test OTLP-heavy workloads. |
| `--read-rps` | 1 | Read traffic per endpoint per second. Raise to model a busier dashboard. |
| `--ingest-err-threshold` | 0.05 | Allowed ingest error rate before a step fails. |
| `--read-p99-threshold-ms` | 3000 | The dashboard latency budget. |

The ramp doubles target each step (100→200→400→800→…), so the cliff
resolution is coarse by design — for a precise number, run a second pass with
`--initial-eps` set near the known cliff and shorter steps.

---

## Common pitfalls

- **`database disk image is malformed (11)`** in the container logs.
  Previous run was killed mid-write and the SQLite WAL didn't recover. Run
  step 1 again with `down -v`.

- **`actual` is ~10× lower than `target` on every step.**
  You're running an older loadgen binary. Rebuild:
  ```bash
  (cd benchmarks/loadgen && go build -o loadgen .)
  ```

- **`seed-project.sh` returns `409 Conflict: organization already exists`.**
  The DB isn't fresh — go back to step 1 with `down -v`.

- **Port 8087 already in use.**
  Pick another: `BENCH_PORT=8088 docker compose -f ... up -d --build`, then
  use `http://localhost:8088` everywhere.

- **`/tmp/bench-localhost/` gone after a restart.**
  macOS empties `/tmp` on reboot. Use `~/bench-localhost/` or a path inside
  the repo if you want results to survive.

- **Loadgen run was killed; some Docker resources still up.**
  The compose stack is yours to manage — it won't auto-tear-down. Run step 6.

---

## Going beyond local

Once the local run works end-to-end, move to the Hetzner ladder:

```bash
# Validate Hetzner credentials, no money spent
./benchmarks/scripts/run-local.sh --dry-run

# ~5 min, ~€0.02 — one tier, one mode, on real Hetzner
./benchmarks/scripts/run-local.sh --smoke

# Full matrix — 4 tiers x 2 modes, ~1 h, ~€1.20
./benchmarks/scripts/run-local.sh
```

Prereqs and detail in [README.md](./README.md). The same `run-matrix-entry.sh`
script powers both the laptop path and the GitHub Actions workflow.
