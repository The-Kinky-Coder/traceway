# benchmarks/

Hardware-vs-breaking-point benchmark for the Traceway backend. Provisions
Hetzner Cloud servers, hammers the backend with realistic ingest while
measuring read latency, and produces a chart of sustainable throughput per
hardware tier and DB mode.

## What it answers

> On hardware tier **X** with DB config **Y**, you can sustain **N** traces +
> spans + session-replays per second while reads remain responsive.

Two DB modes are tested:
- **sqlite** — single-binary Traceway with embedded SQLite (`Dockerfile.sqlite`).
- **pgch** — full ClickHouse + Postgres stack (`Dockerfile.minimal`).

Four hardware tiers, all Hetzner CCX (dedicated vCPU) so neighbor noise doesn't
pollute the read-latency signal:

| Tier  | vCPU | RAM   | Disk        |
|-------|------|-------|-------------|
| CCX13 | 2    | 8 GB  | 80 GB NVMe  |
| CCX23 | 4    | 16 GB | 160 GB NVMe |
| CCX33 | 8    | 32 GB | 240 GB NVMe |
| CCX43 | 16   | 64 GB | 360 GB NVMe |

## How a run works

Per matrix entry (one tier × one mode):

1. `hetzner-up.sh` provisions a SUT box + a CAX11 loadgen box on a private network (override with `LOADGEN_TIER`).
2. `sut-bootstrap.sh` rsyncs the repo, installs Docker, brings up the right
   `docker-compose.<mode>.yml`, waits for `/health`.
3. `seed-project.sh` registers a fresh user + project, captures the JWT and
   project token.
4. `loadgen-bootstrap.sh` cross-compiles the loadgen for linux/amd64, pushes
   it to the loadgen box, runs it against the SUT's *private* IP.
5. The loadgen ramps target ingest in doubling steps (100 -> 200 -> 400 -> ...)
   while a constant read mix probes the dashboard endpoints. Each step holds
   for `--step-duration` (default 2 min).
6. A step "fails" when ingest error rate > 5%, OR any read endpoint P99 > 3000ms.
   Max sustainable EPS is the actual rate of the last passing step.
7. `hetzner-down.sh` deletes everything via a bash `trap` — even on Ctrl-C.

After all matrix entries finish, `chart.py` renders `chart.png` (bar) +
`chart-detail.png` (line: read P99 vs ingest EPS) + `summary.md`.

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

One tier (ccx13), one mode (sqlite), short steps. If this works the full
matrix works.

### Full matrix (~1 h, ~€1.20)

```bash
./benchmarks/scripts/run-local.sh
```

4 tiers × 2 modes, default `--duration 30m`. Output goes to
`benchmarks/results/<YYYY-MM-DD>-local/` — the `-local` suffix exists so
dev runs never get confused with CI runs and never get accidentally committed.

### Other useful invocations

```bash
# Validate environment without provisioning anything (free)
./benchmarks/scripts/run-local.sh --dry-run

# Re-run just one tier with default duration
./benchmarks/scripts/run-local.sh --tier ccx23 --mode pgch

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

## Layout

```
benchmarks/
  compose/                       # SUT-side docker compose, one per mode
    docker-compose.sqlite.yml
    docker-compose.pgch.yml
  loadgen/                       # The Go binary that generates load
    main.go                      # CLI + orchestration
    ingest.go                    # Stream A — stepped ingest worker pool
    report.go                    # /api/report JSON+gzip sender
    otlp.go                      # OTLP/HTTP protobuf sender
    reads.go                     # Stream B — constant read mix
    ramp.go                      # Step controller + breaking-point logic
    stats.go                     # Latency tracker (percentiles via sort)
    util.go, log.go              # Misc helpers
  scripts/
    run-local.sh                 # ★ Laptop entry point
    run-matrix-entry.sh          # One matrix cycle (used by local + CI)
    preflight.sh                 # Validates env before any provisioning
    hetzner-up.sh                # hcloud server create
    hetzner-down.sh              # hcloud server delete (idempotent)
    sut-bootstrap.sh             # Installs Docker, brings up backend
    loadgen-bootstrap.sh         # Cross-compiles + runs loadgen
    seed-project.sh              # /api/register -> JWT + project token JSON
    chart.py                     # matplotlib renderer
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
- **All steps pass on the highest tier.** Increase `--max-eps` (default 50000)
  if you want to find the ceiling on big boxes. The default is meant to keep
  one-tier runs bounded; on CCX43 with PG/CH the ceiling can be well above 10k EPS.
