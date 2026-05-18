# benchmarks/charts.md — chart reading guide

Every PNG under `benchmarks/results-throughput/` and `benchmarks/results-probe/`
is rendered by [`benchmarks/scripts/chart.py`](scripts/chart.py) from the JSON
files in the same folder. This doc explains what each chart shows, what
question it answers, how to read it, and what to look for. Pair this with
[README.md](README.md) for the bench-mechanics side.

The two output folders are siblings — throughput runs never touch
`results-probe/` and vice-versa. Each folder is wiped on every dispatch of
its scenario, so the committed PNGs always reflect exactly one run.

**File naming.** Per-signal charts: `chart-<name>-<signal>.png` where
`<signal>` ∈ {`spans`, `metrics`, `logs`}. Cross-signal charts have no
signal suffix (`chart-signal-mix.png`, `chart-readprobe-signal-mix.png`).
The headline bar is the special case: `chart-<signal>.png` (no `-headline`
suffix; it's the most-cited artifact and gets the short name).

---

## Reading guide — throughput

Folder: `benchmarks/results-throughput/`

| Question | Chart |
|---|---|
| What's the headline number? | [`chart-<signal>.png`](#t1--chart-signalpng--headline-bars) |
| How does latency grow with batch size at low concurrency? | [`chart-phase1-batch-<signal>.png`](#t2--chart-phase1-batch-signalpng--phase-1-batch-size-scaling) |
| Where does the SUT cliff on collector-shape rate? | [`chart-phase2-rate-<signal>.png`](#t3--chart-phase2-rate-signalpng--phase-2-collector-shape-rate-ramp) |
| Where does the SUT cliff on SDK-fleet-shape rate? | [`chart-phase3-rate-<signal>.png`](#t4--chart-phase3-rate-signalpng--phase-3-sdk-fleet-rate-ramp-conditional) |
| What does each K items/sec cost in P99? | [`chart-pareto-<signal>.png`](#t5--chart-pareto-signalpng--latencythroughput-pareto) |
| Does doubling cores double throughput? | [`chart-tier-scaling-<signal>.png`](#t6--chart-tier-scaling-signalpng--tier-scaling) |
| What worked and what failed, at a glance? | [`chart-cliff-<signal>.png`](#t7--chart-cliff-signalpng--step-status-grid) |
| Is batching still paying off, or is the SUT dropping batches? | [`chart-batch-efficiency-<signal>.png`](#t8--chart-batch-efficiency-signalpng--batch-efficiency) |
| At the same hardware, how do spans vs metrics vs logs compare? | [`chart-signal-mix.png`](#t9--chart-signal-mixpng--cross-signal-comparison) |

## Reading guide — read-probe

Folder: `benchmarks/results-probe/`

| Question | Chart |
|---|---|
| How many rows can you have before the dashboard cliffs? | [`chart-readprobe-headline-<signal>.png`](#p1--chart-readprobe-headline-signalpng--max-fill-level-by-tier-mode) |
| How steeply does read latency grow with table size? | [`chart-readprobe-<signal>.png`](#p2--chart-readprobe-signalpng--read-latency-vs-rows-ingested) |
| How long does it take to fill each level? | [`chart-readprobe-ingest-time-<signal>.png`](#p3--chart-readprobe-ingest-time-signalpng--fill-time-vs-target-rows) |
| Does more hardware buy a bigger queryable dataset? | [`chart-readprobe-tier-scaling-<signal>.png`](#p4--chart-readprobe-tier-scaling-signalpng--read-probe-tier-scaling) |
| Which (tier × mode × fill) combos passed? | [`chart-readprobe-cliff-<signal>.png`](#p5--chart-readprobe-cliff-signalpng--read-probe-step-status-grid) |
| Which read endpoint scales worst on the same hardware? | [`chart-readprobe-signal-mix.png`](#p6--chart-readprobe-signal-mixpng--read-probe-cross-signal) |

---

## Failure-state glossary

The charts use a shared vocabulary for what "failed" means at each step. Most
status colors are reused across the cliff grids.

| Term | Meaning | Where it shows |
|---|---|---|
| **pass** (green) | Step is below the threshold and the loadgen marked it passing. | Cliff grids, `o` markers on lines/scatter. |
| **soft-cliff** (yellow) | Throughput steps: achieved ≤ 70% of attempted with no error spike — the SUT's workers couldn't keep up with the limiter. Read-probe: the read returned 200 but latency was above threshold (default 5000ms). | Cliff grid yellow cell; on line charts, look for the gap between `actual` (solid) and `attempted` (dashed) lines in T3/T8. |
| **hard fail** (red) | Throughput: combined HTTP error rate + OTLP `rejected` items exceeded 5% of attempted. Read-probe: the read returned a non-2xx response or errored out. | Cliff grid red cell; `x` marker on every line/scatter chart. |
| **0 requests completed** / **dead** (dark grey) | The SUT never returned a single OK response in the step's window — usually the container was being restarted by Docker's `restart: on-failure:3`. Shows up as `actualItemsPerSec = 0`. | Cliff grid dark-grey cell; Pareto chart bottom annotation; line charts drop the point (log scale). |
| **not run** (light grey) | The phase or step didn't execute in this matrix entry (e.g. Phase 2 was skipped because Phase 1 cliffed first, or this fill level was never reached). | Cliff grid light-grey cell. |
| **failed @ P1 step N** | Run never produced any sustainable throughput. Annotation on headline bars. | T1 hatched bar, T6 dropped point. |
| **Phase 1 floor** | The Phase 1 P99 latency at the matching batch size — what latency looks like with zero concurrency. Anything above the floor is concurrency tax. | Thin dotted horizontal lines on T3 / T4 middle subplot. |

Numeric thresholds match `benchmarks/loadgen/` defaults: `--ingest-err-threshold 0.05` (5% combined error rate) and `--soft-cliff-ratio 0.7` (70% achieved-vs-attempted).

---

## Throughput charts

### T1 — `chart-<signal>.png` — Headline bars

- **Filename:** `chart-spans.png`, `chart-metrics.png`, `chart-logs.png`
- **Folder:** `results-throughput/`
- **What it shows:** Grouped bars of `maxSustainableItemsPerSec` per (tier, mode).
- **Question answered:** On hardware tier X with DB mode Y, what's the largest sustainable ingest rate for signal Z?
- **How to read it:**
  - X axis: hardware tier. Y axis: max sustainable items/sec.
  - Each bar is annotated with the absolute value and the **phase that produced it** in parentheses (`283,187 (P2)`). P1 = batch-size ramp at 5 req/sec, P2 = collector-shape rate ramp, P3 = SDK-fleet-shape rate ramp.
  - **Hatched grey bars** signal "all phases failed" — the run never produced any sustainable throughput. The annotation tells you which phase + step was the first to fail (`failed @ P1 step 3`).
- **What to look for:**
  - Mode floor — sqlite vs pgch gap shows the cost of the simpler stack.
  - Phase tag — a P1-sourced headline means the SUT saturated on a single client; P2/P3 means it scaled with concurrency.
  - Hatched bars vs `0`-height bars: the old chart rendered failures as zero-height bars that looked identical to "not run." Now they're explicitly flagged.
- **Why it's good:** Headline answer in one image. Disambiguates "failed" from "not measured" — a UX gap the old chart had.

### T2 — `chart-phase1-batch-<signal>.png` — Phase 1 batch-size scaling

- **Folder:** `results-throughput/`
- **What it shows:** Latency P50/P95/P99 vs batch size at fixed 5 req/sec.
- **Question answered:** At fixed 5 req/sec, how does per-request latency grow as you fatten the batch?
- **How to read it:**
  - X axis: batch size, log scaled. Y axis: ingest latency in ms, log scaled.
  - Per (tier, mode) there are three lines: P50 (dotted, alpha 0.5), P95 (dash-dot, alpha 0.7), P99 (solid). Only the P99 line gets a legend entry; lower percentiles are colour-coded by the same colour.
  - `o` markers = step passed. `x` markers = step failed at this batch size.
- **What to look for:**
  - **Cliff knee** — the batch size where the line bends sharply upward. That's where the SUT stops scaling linearly with batch.
  - **Mode spread** — sqlite typically cliffs at ~1K, pgch cruises out to 16K, managed-ch pays a fixed network-RTT floor (visible as a high P50 even at the smallest batch).
- **Why it's good:** Phase 1 saturates one client; the latency curve here is the "no concurrency" baseline that Phase 2/3 charts overlay as the "P1 floor."

### T3 — `chart-phase2-rate-<signal>.png` — Phase 2 collector-shape rate ramp

- **Folder:** `results-throughput/`
- **What it shows:** Three stacked subplots over a shared X axis (request rate, log) for the collector-shape rate ramp (fixed fat batch from Phase 1's winner, default cap 16384).
- **Question answered:** Where does the SUT cliff when a single fat-batch client (like an OTel Collector) ramps its request rate?
- **How to read it:**
  - **Top subplot — throughput.** Solid line = `actualItemsPerSec`. Dashed line = `attemptedItemsPerSec` (= batch × target rate). When solid drops below dashed, the SUT can't keep up — that's the soft-cliff. Log Y.
  - **Middle subplot — latency percentiles.** P50/P95/P99 per run. Thin dotted horizontal lines = **Phase 1 P99 floor at the matching batch size** (the no-concurrency latency). Anything above the floor is concurrency tax. Log Y.
  - **Bottom subplot — failure signal.** Solid line = HTTP error rate (%). Twin Y axis dotted line (if any rejection happened) = OTLP `rejected` count. The horizontal red dashed line at 5% is the configured failure threshold.
  - `o` = passed step, `x` = failed step. Steps are sorted by `requestRate` before plotting because the bisection writer records them in execution order (e.g. 5 → 25 → 15 → 10 → 12).
- **What to look for:**
  - **Soft cliff vs hard fail** — does the SUT degrade gracefully (solid drifts below dashed in the top subplot) or fall off a cliff (errors spike in the bottom subplot at the same rate)?
  - **Concurrency tax** — distance between the latency line and the dotted P1 floor.
  - **Bisection precision** — the X markers should cluster tightly near the cliff. Wide spread means the bisection ran out of steps.
- **Why it's good:** Strictly supersets the legacy `chart-detail-<signal>.png` (which plotted only P99). Adds the achieved-vs-attempted gap, the error breakdown, and the P1 floor — the three signals you need to diagnose a cliff.

### T4 — `chart-phase3-rate-<signal>.png` — Phase 3 SDK-fleet rate ramp (conditional)

- **Folder:** `results-throughput/`
- **What it shows:** Same three-subplot layout as T3, but for the SDK-fleet shape: batch fixed at 100 (typical OTel SDK output), request rates from 10 to 10,000.
- **Question answered:** When each request is tiny and rates are huge, does the SUT cliff in a different place — HTTP/auth/decode path vs raw-insert path?
- **How to read it:** Identical to T3. The X-axis range is much wider (10–10,000 vs ~1–400 in T3) — Phase 3 stresses request handling rather than batch insertion.
- **What to look for:**
  - **Phase ceiling vs Phase 2 ceiling** — Phase 3 typically saturates HTTP/auth before raw insert, so the items/sec ceiling is lower despite higher request rate. If Phase 3's ceiling beats Phase 2's, the SUT is under-utilizing its insert path under fat batches.
  - **Failures** — at thousands of req/sec, decode and auth checks become measurable; spikes there manifest as bottom-subplot errors before the latency lines bend.
- **Why it's good:** Only chart that probes the "thousands of small batches" SDK fleet shape. Different bottleneck profile than the collector shape — surfacing it separately keeps the analysis honest.
- **Note:** This chart is only rendered when at least one run in the folder has a non-empty `phase3` (controlled by the `--phase3-batch-size`/`--phase3-request-rates` loadgen flags).

### T5 — `chart-pareto-<signal>.png` — Latency–throughput Pareto

- **Folder:** `results-throughput/`
- **What it shows:** Scatter of every Phase 2 and Phase 3 step at `(P99 latency, actualItemsPerSec)`. Both axes log. Color per (tier, mode), `o` = passed, `x` = failed.
- **Question answered:** What does each extra K items/sec cost you in P99 latency? What's each (tier, mode) Pareto frontier look like?
- **How to read it:**
  - X axis: P99 latency in ms, log. Y axis: achieved items/sec, log.
  - **Faint dotted polyline per run** connects passed steps sorted by P99 — that's the (tier, mode) frontier.
  - **Up-and-left is better** (more throughput, less latency).
  - **Zero-throughput failures** can't be plotted on a log Y axis — they're listed as a bottom-margin annotation in red (`Zero-throughput failures: ccx23 / sqlite@P2s1; ccx23 / pgch@P2s1`).
- **What to look for:**
  - **Frontier shape** — a flat-then-vertical curve means "lots of room to push more rate before latency rises" (the good case). A steep diagonal means "every K of throughput costs you proportionally in latency" (the bad case).
  - **Mode separation** — sqlite frontiers usually live in the lower-left (low throughput, low latency). pgch frontiers extend right. managed-ch's RTT floor pushes them rightward at the same throughput levels.
  - **Failure clustering** — `x` markers near the frontier top-right tell you what the SUT *attempted* before cliffing.
- **Why it's good:** The only chart that puts the latency-throughput trade-off on a single picture *and* keeps failures visible without crashing on log scale.

### T6 — `chart-tier-scaling-<signal>.png` — Tier scaling

- **Folder:** `results-throughput/`
- **What it shows:** Line: `maxSustainableItemsPerSec` vs hardware tier (ordinal X: ccx13 → ccx43), one line per mode. Tick labels include vCPU/RAM.
- **Question answered:** Does doubling cores actually double throughput? Where does extra hardware stop paying?
- **How to read it:**
  - X axis: tier in linear order (ccx13, ccx23, ccx33, ccx43 — Hetzner doubles vCPU at each step).
  - Y axis: max sustainable items/sec.
  - **Dashed reference line per mode** = perfect linear-vCPU scaling, anchored at the smallest tier present for that mode. If the solid line falls below the dashed, you're paying a scaling penalty.
- **What to look for:**
  - **Linear vs diminishing returns** — sqlite often plateaus quickly (single-process write); pgch usually scales better but with a knee at the largest tier (ClickHouse merges saturate disk I/O).
  - **Missing points** — if a mode has no point at a tier, it failed there. The reference line still projects, giving you a "what perfect scaling would have looked like" anchor.
- **Why it's good:** Distinguishes "you need a bigger box" from "the bottleneck is elsewhere." A flat-line mode says "more cores won't help."

### T7 — `chart-cliff-<signal>.png` — Step-status grid

- **Folder:** `results-throughput/`
- **What it shows:** Gantt-style heatmap. Rows = (tier, mode). Columns = step indices grouped by phase (P1 | P2 | P3 with vertical separators).
- **Question answered:** At a glance, what worked and what failed?
- **How to read it:**
  - Each cell is one step of one phase. Cell color: green=pass, yellow=soft-cliff, red=hard fail, dark grey=0 requests completed, light grey=not run.
  - Inside each cell: `bs=<batch size>  r=<request rate>` (e.g. `bs=16K r=25`).
  - **Right margin:** the headline number for the row, or "failed" in red for all-failed runs.
- **What to look for:**
  - **Red rows** — runs that cliffed almost immediately. The old line charts hid these (zero-height bar or invisible single point); the cliff grid makes them legible.
  - **Yellow + red mix** — bisection territory. Several yellow cells around a red one means the loadgen narrowed the cliff via bisection.
  - **Missing-column patterns** — sqlite often has Phase 1 reds and no Phase 2 (because Phase 1 cliffed first, so Phase 2 was skipped). pgch's failures usually live in Phase 2 because Phase 1 cleared all batches.
- **Why it's good:** The most information-dense chart. Especially useful for runs that mostly fail — line charts make those invisible, but the grid shows every step and why.

### T8 — `chart-batch-efficiency-<signal>.png` — Batch efficiency

- **Folder:** `results-throughput/`
- **What it shows:** Phase 1 throughput vs batch size. Solid = `actualItemsPerSec`, dashed = `attemptedItemsPerSec` (= batch × 5 req/sec).
- **Question answered:** Is the batch size still paying off, or is the SUT rejecting fat batches?
- **How to read it:**
  - X axis: batch size, log. Y axis: items/sec, log.
  - Solid line = actually-delivered items/sec. Dashed line per run = attempted items/sec (the diagonal a perfect SUT would track).
  - **Solid below dashed** = the SUT is rejecting items (HTTP errors or OTLP `PartialSuccess.rejected_log_records` etc.).
  - `x` markers = step failed at that batch size.
- **What to look for:**
  - **The point where the solid line peels away from the dashed** — that's the batch size at which more items per request stop helping.
  - **Failed marker placement** — if `x` is exactly on the dashed line, the SUT accepted the work but errored out at the *next* step. If `x` is below the dashed, the SUT was already rejecting before failing.
- **Why it's good:** Disentangles "the load gen sent X" from "the SUT accepted X." Two systems with identical headlines can look very different here.

### T9 — `chart-signal-mix.png` — Cross-signal comparison

- **Folder:** `results-throughput/`
- **What it shows:** Single grouped-bars figure: one group per (tier, mode), three bars per group — spans / metrics / logs.
- **Question answered:** At the same hardware, how does the cost of each signal compare?
- **How to read it:**
  - X axis: (tier × mode) combinations. Y axis: max sustainable items/sec.
  - Bar colour: signal (spans blue, metrics green, logs red).
  - **The Y unit varies by signal** — spans/sec vs data points/sec vs log records/sec aren't strictly comparable item-to-item. Use relative shape rather than absolute height: which signal does this SUT do best on?
- **What to look for:**
  - **Signal ranking per row** — usually metrics > logs > spans (metrics is the cheapest schema, spans the heaviest), but the ratio shifts by mode.
  - **Missing bars** — a signal absent from a (tier, mode) row means that matrix entry failed.
- **Why it's good:** Reveals SUT specialization. A box that's great at spans but bad at logs probably has a string-store bottleneck.

---

## Read-probe charts

### P1 — `chart-readprobe-headline-<signal>.png` — Max fill level by (tier, mode)

- **Folder:** `results-probe/`
- **What it shows:** Grouped bars of `maxFillLevelPassed` (rows) per (tier, mode). Log Y axis.
- **Question answered:** How many rows can you have in the relevant table before the dashboard read on this endpoint cliffs?
- **How to read it:**
  - X axis: hardware tier. Y axis: max passing fill level (rows, log).
  - Each bar is annotated with the fill level (`1.0M`) and the read latency at that fill (`540ms`).
  - **Hatched grey bars** = no fill level passed (the smallest fill level was already over threshold).
  - The title names the read endpoint being probed (e.g. `/api/metrics/application`).
- **What to look for:**
  - **Order-of-magnitude steps** — fill levels usually 100K → 1M → 10M → 100M. A row that maxes at 100K is in serious trouble; 100M is comfortable.
  - **Mode spread** — sqlite usually maxes out an order of magnitude earlier than pgch.
- **Why it's good:** The read-probe equivalent of T1. One number per (tier, mode), with the latency-at-max so you know how close to the threshold the bar represents.

### P2 — `chart-readprobe-<signal>.png` — Read latency vs rows ingested

- **Folder:** `results-probe/`
- **What it shows:** Line: `readLatencyMs` (Y log) vs `rowsIngested` (X log), one line per (tier, mode). Horizontal threshold line (default 5000ms).
- **Question answered:** How steeply does read latency grow as the table fills up?
- **How to read it:**
  - X axis: rows ingested before the probe, log. Y axis: read latency in ms, log.
  - **Solid horizontal red dashed line** = the threshold from `--read-threshold-ms` (default 5000ms).
  - `o` = step passed (read OK, under threshold). `x` = failed (read errored or over threshold).
  - Title names the read path being probed.
- **What to look for:**
  - **Slope** — log-log slope ≈ 1 means linear scaling; > 1 means superlinear (degrading); < 1 means sublinear (the index is doing its job).
  - **Where the line crosses the threshold** — that's the cliff. Anything to the right of the crossing point is unreliable.
- **Why it's good:** The classic "how does the dashboard hold up?" picture. Slope information is the part you couldn't get from the headline.

### P3 — `chart-readprobe-ingest-time-<signal>.png` — Fill time vs target rows

- **Folder:** `results-probe/`
- **What it shows:** Line: time taken to reach the target fill level (Y log seconds) vs target fill level (X log rows), one line per (tier, mode).
- **Question answered:** How long does it take to FILL each level — i.e. how expensive are the writes at high cardinality?
- **How to read it:**
  - X axis: target fill level (rows, log). Y axis: ingest seconds elapsed (log).
  - Lines are read in execution order — they keep climbing as the test fills the table.
- **What to look for:**
  - **Curve sag** — if the curve bends upward at higher row counts, write cost is rising (compaction / index thrash). A perfectly linear log-log line means write speed is constant regardless of table size.
  - **Mode delta** — ClickHouse usually sags around 10–100M because background merges saturate the disk; SQLite usually fails before sagging at all.
- **Why it's good:** The read-probe view focuses on read latency, but the cost of *getting there* matters too. If filling 100M rows takes hours, that's a different operational story than seconds.

### P4 — `chart-readprobe-tier-scaling-<signal>.png` — Read-probe tier scaling

- **Folder:** `results-probe/`
- **What it shows:** Line: `maxFillLevelPassed` vs tier ordinal, one line per mode. Log Y. Dashed linear-vCPU reference per mode.
- **Question answered:** Does more hardware buy you a bigger queryable dataset on this endpoint?
- **How to read it:** Mirror of T6 but Y is rows instead of items/sec. Same linear-vCPU reference convention.
- **What to look for:**
  - **Plateau** — flat lines mean read scaling is bottlenecked elsewhere (network RTT for managed-ch, single-thread query plan for sqlite).
  - **Below-reference gap** — distance between solid and dashed = how much the read path leaves on the table.
- **Why it's good:** Throughput tier-scaling answers "can I serve more SDK traffic?". This one answers "can I keep the dashboard working at a bigger scale?". Different question, same chart shape.

### P5 — `chart-readprobe-cliff-<signal>.png` — Read-probe step-status grid

- **Folder:** `results-probe/`
- **What it shows:** Gantt-style heatmap. Rows = (tier, mode). Columns = fill levels (union across all runs).
- **Question answered:** Which (tier × mode × fill level) combos passed?
- **How to read it:**
  - Each cell is one probe of one fill level. Cell color: green=passed, yellow=read OK but over threshold, red=read errored, light-grey=not run.
  - Inside each cell: the read latency (`540ms`). Empty when not run.
  - Right margin: `max <fill>` if any fill passed, or "no fill passed" in red.
  - Column headers show the fill levels themselves (`100K`, `1.0M`, etc.).
- **What to look for:**
  - **Yellow vs red** — yellow means "the read came back but slowly," red means "the read errored out." Yellow is often actionable (tune the query plan); red usually means the SUT itself is unwell.
  - **Row patterns** — sqlite rows often have a single green cell and the rest red/yellow; pgch rows usually have a few green cells before yellow.
- **Why it's good:** The cliff equivalent of T7 but for read-probe. Lets you see the failure pattern across the full matrix on one page.

### P6 — `chart-readprobe-signal-mix.png` — Read-probe cross-signal

- **Folder:** `results-probe/`
- **What it shows:** Single grouped-bars figure: one group per (tier, mode), three bars per group — spans / metrics / logs. Y = `maxFillLevelPassed` (log).
- **Question answered:** Which read endpoint scales worst on the same hardware?
- **How to read it:**
  - Same shape as T9 but for read-probe.
  - Legend includes the read path for each signal (`spans (/api/endpoints/grouped)`), since each signal probes a different endpoint.
  - Log Y because fill levels span several orders of magnitude.
- **What to look for:**
  - **Signal worst-case** — usually `spans` (`/api/endpoints/grouped` joins on a high-cardinality table) is the first to cliff.
  - **Mode-invariant cliffs** — if every mode's bar for `spans` is short, the bottleneck is the query plan, not the storage engine.
- **Why it's good:** Three read paths × N (tier, mode) cells on one image. Shows you which API endpoint to optimize first.

---

## Generating these locally

```bash
python3 -m venv .venv-bench
.venv-bench/bin/pip install matplotlib

# Throughput results
.venv-bench/bin/python benchmarks/scripts/chart.py benchmarks/results-throughput

# Read-probe results
.venv-bench/bin/python benchmarks/scripts/chart.py benchmarks/results-probe
```

`chart.py` auto-detects scenarios per JSON file — running it against a folder
that mixes both scenarios will produce the union of charts. The wrong-scenario
render functions are no-ops, so you'll just see the charts that have data.
