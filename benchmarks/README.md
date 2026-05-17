# benchmarks/

Hardware-vs-throughput benchmark for the Traceway backend. Provisions
Hetzner Cloud servers, runs OTLP ingest in realistic batch shapes (one signal
per matrix entry), and produces per-signal charts of sustainable throughput
per hardware tier and DB mode.

## What it answers

> On hardware tier **X** with DB config **Y**, you can sustain **N** spans/sec,
> **M** metric data points/sec, and **P** log records/sec via OTLP under
> collector-shaped batch traffic (gzipped protobuf, batches up to 8192).

Three signals are tested in separate matrix entries:
- **spans** → `POST /api/otel/v1/traces` (`ExportTraceServiceRequest`)
- **metrics** → `POST /api/otel/v1/metrics` (`ExportMetricsServiceRequest`, Gauge data points)
- **logs** → `POST /api/otel/v1/logs` (`ExportLogsServiceRequest`)

Three DB modes are supported:
- **sqlite** — single-binary Traceway with embedded SQLite (`Dockerfile.sqlite`).
- **pgch** — full ClickHouse + Postgres stack, all in Docker on the SUT (`Dockerfile.minimal`).
- **managed-ch** — `Dockerfile.minimal` pointed at an externally-hosted ClickHouse (ClickHouse Cloud, Aiven, Altinity, etc.) via env vars. Postgres still runs locally in the SUT's Docker. See [Running against managed ClickHouse](#running-against-managed-clickhouse).

Four hardware tiers, all Hetzner CCX (dedicated vCPU) so neighbor noise doesn't
pollute the latency signal:

| Tier  | vCPU | RAM   | Disk        |
|-------|------|-------|-------------|
| CCX13 | 2    | 8 GB  | 80 GB NVMe  |
| CCX23 | 4    | 16 GB | 160 GB NVMe |
| CCX33 | 8    | 32 GB | 240 GB NVMe |
| CCX43 | 16   | 64 GB | 360 GB NVMe |

## Not measured here

This benchmark is **ingest-only** — it does not probe dashboard read latency.
A separate future benchmark will pre-load N events (100k, 1M, 10M, 100M) and
measure read endpoint P99 at each fill level. Don't read the numbers here as
"the dashboard stays usable at N items/sec" — they're "the SUT keeps swallowing
items at N items/sec without erroring or rejecting."

## How a run works

Per matrix entry (one tier × one mode × one signal):

1. `hetzner-up.sh` provisions a SUT box + a CAX11 loadgen box on a private network (override with `LOADGEN_TIER`).
2. `sut-bootstrap.sh` rsyncs the repo, installs Docker, brings up the right
   `docker-compose.<mode>.yml`, waits for `/health`.
3. `seed-project.sh` registers a fresh user + project, captures the project
   bearer token.
4. `loadgen-bootstrap.sh` cross-compiles the loadgen, pushes it to the
   loadgen box, runs it with `--signal <spans|metrics|logs>` against the
   SUT's *private* IP.
5. The loadgen runs a two-phase ramp:
   - **Phase 1 — batch-size ramp.** Single client at a fixed 5 req/sec.
     Batch sizes step through `256,1024,4096,8192,16384`. Each step holds for
     `--step-duration` (default 2 min). Stops at the first failing step.
   - **Inter-phase cooldown.** Phase 1's last step typically runs the SUT
     near saturation; the loadgen sleeps `--inter-phase-cooldown-seconds`
     (default 30 s) before starting Phase 2 so the SUT can drain its
     queues (CH merges, PG WAL, HTTP handler pool). Without this pause,
     pgch runs reliably produced "0 OK / 0 errors" Phase 2 because new
     requests sat on the SUT-side TCP backlog while the SUT was still
     digesting Phase 1's final wave.
   - **Phase 2 — request-rate ramp.** Batch size fixed at
     `min(Phase 1 winner, --phase2-batch-cap=16384)`. Request rates step
     through `1,5,25,100,400`. When the coarse ramp finds the first failing
     step, up to `--phase2-bisect-max-steps` (default 3) bisection steps run
     between the last passing rate and the failing rate to pin the cliff
     within `--phase2-bisect-tolerance` (default 20%). So `5→25` (a 5× jump)
     gets refined into something like `5, 15, 10, 12` until the gap is
     <20% of the last passing rate.
6. A step "fails" when **either**:
   - combined error rate (HTTP failures + OTLP `PartialSuccess` rejected items)
     exceeds `--ingest-err-threshold` (default 5%), **or**
   - the *achieved* request rate falls below `--soft-cliff-ratio` × target rate
     (default 70%) — meaning the workers can't keep up with the limiter, the
     SUT has cliffed on latency, and we'd be erroring out one step later anyway.

   The headline `maxSustainableItemsPerSec` is the highest `actualItemsPerSec`
   recorded across passing Phase 2 steps — measured from real OK responses,
   not the formula `batchSize × targetRate` (which over-reports when workers
   saturate before the limiter does). If Phase 2 produces zero passing steps
   (rare, but happens when Phase 1 leaves the SUT in an unrecoverable state
   even after the cooldown), the headline falls back to the best Phase 1
   `actualItemsPerSec` so the run still reports a meaningful number.
7. `hetzner-down.sh` deletes everything via a bash `trap` — even on Ctrl-C.

After all matrix entries finish, `chart.py` renders three bar charts
(`chart-spans.png`, `chart-metrics.png`, `chart-logs.png`), three detail
charts (`chart-detail-<signal>.png`), and a combined `summary.md`.

## Running from your laptop

### One-time setup

1. **Install tooling**: `hcloud`, `jq`, `python3`, Go 1.25+, `rsync`. On macOS:
   `brew install hcloud jq rsync go`.
2. **Install matplotlib in a venv** (system Python is usually PEP 668 locked):
   ```bash
   python3 -m venv .venv
   source .venv/bin/activate
   pip install matplotlib
   ```
3. **Generate an SSH key** specifically for benchmarks and upload its public
   half to Hetzner under the name `benchmark-key`:
   ```bash
   ssh-keygen -t ed25519 -f ~/.ssh/hetzner_benchmark -C benchmark-key
   chmod 600 ~/.ssh/hetzner_benchmark
   hcloud ssh-key create --name benchmark-key --public-key-from-file ~/.ssh/hetzner_benchmark.pub
   ```
4. **Export creds** (use `direnv` or a sourced `.envrc.local`, never commit):
   ```bash
   export HCLOUD_TOKEN=...
   export BENCHMARK_SSH_KEY=~/.ssh/hetzner_benchmark
   ```

### Smoke (cheap — ~5 min, ~€0.02)

```bash
./benchmarks/scripts/run-local.sh --smoke
```

One tier (ccx13), one mode (sqlite), one signal (spans), short steps. If this
works, the full matrix works.

### Full matrix (~3 h, ~€3.60)

```bash
./benchmarks/scripts/run-local.sh
```

4 tiers × 2 modes × 3 signals, default `--duration 30m`. Output goes to
`benchmarks/results/<YYYY-MM-DD>-local/` — the `-local` suffix exists so
dev runs never get confused with CI runs and never get accidentally committed.

### Other useful invocations

```bash
# Validate environment without provisioning anything (free)
./benchmarks/scripts/run-local.sh --dry-run

# Re-run just one tier/mode across all signals
./benchmarks/scripts/run-local.sh --tier ccx23 --mode pgch

# One signal only across all tiers/modes
./benchmarks/scripts/run-local.sh --signal spans

# A single matrix cell
./benchmarks/scripts/run-local.sh --tier ccx13 --mode sqlite --signal logs

# Override the per-entry runtime
./benchmarks/scripts/run-local.sh --tier ccx13 --duration 10m

# Drop the "-local" suffix so the result dir matches the CI naming convention
# (use this only when you want to commit a one-off re-benchmark)
./benchmarks/scripts/run-local.sh --tier ccx23 --commit
```

## Running from GitHub Actions

`.github/workflows/benchmark-hardware.yml`, `workflow_dispatch` only. Inputs
mirror the local flags. The workflow YAML is a thin wrapper around the same
`run-matrix-entry.sh` script the local path uses — if it works locally, it
works in CI.

Required GitHub secrets:
- `HCLOUD_TOKEN`
- `BENCHMARK_SSH_PRIVATE_KEY` — the private key matching the Hetzner-side
  `benchmark-key`.

After the matrix completes, an `aggregate` job downloads all artifacts, runs
`chart.py`, and commits `benchmarks/results/<YYYY-MM-DD>/` to `main` via a bot
commit. No PR — it's a generated artifact.

## Running against managed ClickHouse

Setting `modes=managed-ch` in the workflow dispatch (or `--mode managed-ch`
locally) points the SUT's Traceway container at an externally-hosted
ClickHouse. Postgres still runs locally in the SUT's Docker — this benchmark
is about ClickHouse characteristics.

Required GitHub repository secrets (Settings → Secrets and variables → Actions):

| Secret | Example | Notes |
|---|---|---|
| `BENCH_CH_SERVER` | `your-cluster.us-east-1.aws.clickhouse.cloud:9440` | Native TCP endpoint with TLS port (usually `9440`) |
| `BENCH_CH_USERNAME` | `default` | A dedicated bench user is wiser than `default` |
| `BENCH_CH_PASSWORD` | `••••••` | |
| `BENCH_CH_DATABASE` | `traceway` | Optional, defaults to `traceway` |
| `BENCH_CH_HTTPS_PORT` | `8443` | Optional, defaults to `8443` (CH Cloud); some hosts use `8123` for plain HTTP |

The bench user needs `DROP DATABASE` and `CREATE DATABASE` privileges — between
every matrix entry the orchestrator runs `reset-managed-ch.sh`, which wipes and
recreates the bench database via the HTTPS interface so each (tier × signal ×
scenario) cell starts on an empty cluster. ~5–10s of overhead per entry. If
you're running on a shared cluster, **point the bench at a dedicated database**
or you will lose other data.

Locally: export the same vars (`CLICKHOUSE_SERVER`, `CLICKHOUSE_USERNAME`,
`CLICKHOUSE_PASSWORD`, optional `CLICKHOUSE_DATABASE`, `BENCH_CH_HTTPS_PORT`)
before invoking `run-local.sh --mode managed-ch`. Preflight will fail early if
they're missing.

### Caveats specific to managed CH

- **Network RTT dominates.** Hetzner `nbg1` to ClickHouse Cloud `us-east-1` is
  ~100ms each way; that's a wall the cluster can't climb. Match regions
  (`nbg1`/`fsn1`/`hel1` → CH Cloud EU; `ash`/`hil` → CH Cloud US) for numbers
  comparable to local-CH `pgch`.
- **Bandwidth matters too.** At 8192-span gzipped OTLP batches × 400 req/sec
  you push ~30–80 MB/s outbound from the SUT. Hetzner's egress is generous;
  the managed cluster's ingress quota is the more likely cap.
- **Read-probe is the more interesting scenario on managed CH.** Throughput
  often gets bottlenecked on the SUT→cluster pipe before the cluster's actual
  ingest path. Read-probe surfaces the cluster's read scaling, which is what
  you actually buy from a managed offering.

## Layout

```
benchmarks/
  compose/                       # SUT-side docker compose, one per mode
    docker-compose.sqlite.yml
    docker-compose.pgch.yml
    docker-compose.managed-ch.yml  # External CH; Postgres still local
  loadgen/                       # The Go binary that generates load (OTLP-only)
    main.go                      # CLI + orchestration
    ingest.go                    # Worker pool driving the selected signal sender
    otlp_common.go               # Shared OTLP helpers + signalSender interface
    otlp_spans.go                # ExportTraceServiceRequest builder
    otlp_metrics.go              # ExportMetricsServiceRequest builder (Gauge)
    otlp_logs.go                 # ExportLogsServiceRequest builder
    ramp.go                      # Two-phase ramp (batch size, then request rate)
    stats.go                     # Latency tracker (percentiles via sort)
    util.go, log.go              # Misc helpers
  scripts/
    run-local.sh                 # ★ Laptop entry point
    run-matrix-entry.sh          # One matrix cycle (used by local + CI)
    preflight.sh                 # Validates env before any provisioning
    hetzner-up.sh                # hcloud server create
    hetzner-down.sh              # hcloud server delete (idempotent)
    sut-bootstrap.sh             # Installs Docker, brings up backend
    reset-managed-ch.sh          # Wipes the managed CH DB between matrix entries
    loadgen-bootstrap.sh         # Cross-compiles + runs loadgen
    seed-project.sh              # /api/register -> JWT + project token JSON
    chart.py                     # matplotlib renderer (per-signal charts)
    _ssh.sh                      # Shared ssh/rsync helpers
  results/                       # Committed bench results live here
```

## Failure modes & debugging

- **A run leaves Hetzner servers behind.** `hetzner-down.sh` is idempotent and
  callable directly: `./benchmarks/scripts/hetzner-down.sh <run-id>`. Run IDs
  appear in the orchestrator's stderr at the start of each matrix entry. Also
  visible in `hcloud server list` — anything tagged `bench=true` is safe to
  delete.
- **`preflight.sh` complains about `benchmark-key`.** The Hetzner-side SSH key
  resource is named `benchmark-key` regardless of what you named the file
  locally. Re-upload via `hcloud ssh-key create --name benchmark-key
  --public-key-from-file <your-pub-key>`.
- **Docker compose build is slow on tiny tiers.** First-time `docker compose
  up --build` on a CCX13 takes 5–8 minutes (npm install + Go build).
  `sut-bootstrap.sh` waits up to 10 minutes for `/health` before giving up.
  Subsequent runs reuse the Docker layer cache via the persistent volumes.
- **All steps pass on the highest tier.** Raise the upper end of
  `--phase2-request-rates` (default `1,5,25,100,400`) — e.g. add `1000,2000` —
  if you want to find the ceiling on big boxes. The default is meant to keep
  one-tier runs bounded.
