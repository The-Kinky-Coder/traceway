#!/usr/bin/env python3
"""Render per-signal benchmark charts from a directory of loadgen JSONs.

Inputs: a directory containing N JSON files (one per matrix entry), each in
the schema emitted by benchmarks/loadgen/. Files without a "signal" field
(old pre-OTLP-split format) are skipped with a log line.

Outputs depend on which scenarios are present in the directory:
  Throughput scenario:
    - chart-spans.png / chart-metrics.png / chart-logs.png   bar: max items/sec
    - chart-detail-<sig>.png                                  line: ingest p99 vs request rate
  Read-probe scenario:
    - chart-readprobe-<sig>.png                               line: read latency vs fill level
  Always:
    - summary.md                                              tables per signal/scenario
"""

import json
import sys
from pathlib import Path

import matplotlib

matplotlib.use("Agg")
import matplotlib.pyplot as plt

TIER_ORDER = ["ccx13", "ccx23", "ccx33", "ccx43"]
MODE_ORDER = ["sqlite", "pgch"]
SIGNALS = ["spans", "metrics", "logs"]
MODE_COLORS = {"sqlite": "#4f9fff", "pgch": "#ff9f4f"}
ITEM_LABEL = {
    "spans": "spans/sec",
    "metrics": "data points/sec",
    "logs": "log records/sec",
}
ROW_LABEL = {
    "spans": "spans rows",
    "metrics": "metric_points rows",
    "logs": "log_records rows",
}


def load_runs(results_dir: Path) -> list[dict]:
    runs = []
    for p in sorted(results_dir.glob("*.json")):
        with p.open() as fh:
            doc = json.load(fh)
        if "signal" not in doc:
            print(f"skipping {p.name}: missing 'signal' field (pre-OTLP-split format)", file=sys.stderr)
            continue
        if doc["signal"] not in SIGNALS:
            print(f"skipping {p.name}: unknown signal {doc['signal']!r}", file=sys.stderr)
            continue
        doc.setdefault("scenario", "throughput")
        if doc["scenario"] not in ("throughput", "read-probe"):
            print(f"skipping {p.name}: unknown scenario {doc['scenario']!r}", file=sys.stderr)
            continue
        doc["_path"] = p
        runs.append(doc)
    return runs


def tier_rank(t: str) -> int:
    return TIER_ORDER.index(t) if t in TIER_ORDER else len(TIER_ORDER)


def mode_rank(m: str) -> int:
    return MODE_ORDER.index(m) if m in MODE_ORDER else len(MODE_ORDER)


def render_throughput_bar(runs: list[dict], signal: str, out: Path) -> None:
    runs = [r for r in runs if r["signal"] == signal and r["scenario"] == "throughput"]
    if not runs:
        return

    fig, ax = plt.subplots(figsize=(9, 5))
    tiers_present = sorted({r["tier"] for r in runs}, key=tier_rank)
    modes_present = sorted({r["mode"] for r in runs}, key=mode_rank)

    width = 0.8 / max(len(modes_present), 1)
    x = list(range(len(tiers_present)))

    for mi, mode in enumerate(modes_present):
        ys = []
        for tier in tiers_present:
            match = [r for r in runs if r["tier"] == tier and r["mode"] == mode]
            ys.append(match[0].get("maxSustainableItemsPerSec", 0) if match else 0)
        offsets = [xi + (mi - (len(modes_present) - 1) / 2) * width for xi in x]
        bars = ax.bar(offsets, ys, width=width, label=mode, color=MODE_COLORS.get(mode, "#777"))
        for b, y in zip(bars, ys):
            ax.annotate(f"{int(y):,}", xy=(b.get_x() + b.get_width() / 2, y),
                        xytext=(0, 3), textcoords="offset points",
                        ha="center", va="bottom", fontsize=9)

    ax.set_xticks(x)
    ax.set_xticklabels(tiers_present)
    ax.set_ylabel(f"Max sustainable {ITEM_LABEL[signal]}")
    ax.set_title(f"Traceway: max sustainable {signal} ingest by hardware tier\n"
                 "(failure = combined HTTP error + OTLP rejected > 5%)")
    ax.legend(title="DB mode")
    ax.grid(axis="y", linestyle=":", alpha=0.4)
    fig.tight_layout()
    fig.savefig(out, dpi=130)
    plt.close(fig)


def render_throughput_detail(runs: list[dict], signal: str, out: Path) -> None:
    runs = [r for r in runs if r["signal"] == signal and r["scenario"] == "throughput"]
    if not runs:
        return

    fig, ax = plt.subplots(figsize=(10, 6))
    plotted = False

    for run in sorted(runs, key=lambda r: (tier_rank(r["tier"]), mode_rank(r["mode"]))):
        steps = (run.get("phase2") or {}).get("steps") or []
        # Bisection appends steps in execution order, not rate order. Sort
        # by requestRate so the line draws monotonically across the X axis.
        steps = sorted([s for s in steps if s.get("ingest")], key=lambda s: s.get("requestRate", 0))
        xs = [s["requestRate"] for s in steps]
        ys = [s["ingest"]["p99"] for s in steps]
        if not xs:
            continue
        ax.plot(xs, ys, marker="o", label=f"{run['tier']} / {run['mode']}")
        plotted = True

    if not plotted:
        plt.close(fig)
        return

    ax.set_xlabel("Phase 2 request rate (req/sec)")
    ax.set_ylabel("Ingest P99 latency (ms)")
    ax.set_title(f"Ingest latency vs request rate — {signal}")
    ax.set_xscale("log")
    ax.set_yscale("log")
    ax.grid(True, which="both", linestyle=":", alpha=0.4)
    ax.legend(loc="upper left", fontsize=9)
    fig.tight_layout()
    fig.savefig(out, dpi=130)
    plt.close(fig)


def render_readprobe(runs: list[dict], signal: str, out: Path) -> None:
    runs = [r for r in runs if r["signal"] == signal and r["scenario"] == "read-probe"]
    if not runs:
        return

    fig, ax = plt.subplots(figsize=(10, 6))
    plotted = False
    threshold_ms = 5000

    for run in sorted(runs, key=lambda r: (tier_rank(r["tier"]), mode_rank(r["mode"]))):
        rp = run.get("readProbe") or {}
        threshold_ms = rp.get("readThresholdMs", threshold_ms)
        steps = rp.get("steps") or []
        xs = [s["rowsIngested"] for s in steps]
        ys = [s["readLatencyMs"] for s in steps]
        if not xs:
            continue
        ax.plot(xs, ys, marker="o", label=f"{run['tier']} / {run['mode']}")
        plotted = True

    if not plotted:
        plt.close(fig)
        return

    ax.axhline(threshold_ms, linestyle="--", color="#cc3333", alpha=0.7,
               label=f"threshold ({threshold_ms}ms)")
    ax.set_xlabel(f"{ROW_LABEL[signal]} (ingested before probe)")
    ax.set_ylabel(f"Read latency on {readPathForSignal(signal)} (ms)")
    ax.set_title(f"Read latency vs ingested rows — {signal}")
    ax.set_xscale("log")
    ax.set_yscale("log")
    ax.grid(True, which="both", linestyle=":", alpha=0.4)
    ax.legend(loc="upper left", fontsize=9)
    fig.tight_layout()
    fig.savefig(out, dpi=130)
    plt.close(fig)


def readPathForSignal(signal: str) -> str:
    return {
        "spans": "/api/endpoints/grouped",
        "metrics": "/api/metrics/application",
        "logs": "/api/logs",
    }.get(signal, "")


def render_summary(runs: list[dict], out: Path) -> None:
    lines = ["# Traceway hardware benchmark — summary", ""]
    lines.append(f"Runs analyzed: {len(runs)}")
    lines.append("")

    throughput_runs = [r for r in runs if r["scenario"] == "throughput"]
    if throughput_runs:
        lines.append("## Throughput scenario")
        lines.append("")
        lines.append("Failure threshold: combined HTTP error rate + OTLP rejected items > 5% of attempted.")
        lines.append("")
        for signal in SIGNALS:
            sig_runs = [r for r in throughput_runs if r["signal"] == signal]
            if not sig_runs:
                continue
            lines.append(f"### {signal.capitalize()}")
            lines.append("")
            lines.append(f"| Tier | Mode | Max {ITEM_LABEL[signal]} | Phase 1 max batch | Phase 2 max req/sec | Ingest P99 @ max (ms) |")
            lines.append("|------|------|---------|-------------------|---------------------|------------------------|")
            for run in sorted(sig_runs, key=lambda r: (tier_rank(r["tier"]), mode_rank(r["mode"]))):
                max_items = int(run.get("maxSustainableItemsPerSec", 0))
                max_batch = (run.get("phase1") or {}).get("maxBatchSize", 0)
                max_rate = (run.get("phase2") or {}).get("maxRequestRate", 0)
                last_pass = None
                for s in reversed((run.get("phase2") or {}).get("steps") or []):
                    if s.get("passed"):
                        last_pass = s
                        break
                ingest_p99 = int(last_pass["ingest"]["p99"]) if last_pass else 0
                lines.append(f"| {run['tier']} | {run['mode']} | {max_items:,} | {max_batch:,} | {max_rate:g} | {ingest_p99} |")
            lines.append("")

    readprobe_runs = [r for r in runs if r["scenario"] == "read-probe"]
    if readprobe_runs:
        lines.append("## Read-probe scenario")
        lines.append("")
        lines.append("Failure: a read probe exceeded the configured threshold (default 5000ms) or returned an error. "
                     "Max fill level passed is the largest row count at which the read still came back in time.")
        lines.append("")
        for signal in SIGNALS:
            sig_runs = [r for r in readprobe_runs if r["signal"] == signal]
            if not sig_runs:
                continue
            lines.append(f"### {signal.capitalize()} — probing `{readPathForSignal(signal)}`")
            lines.append("")
            lines.append(f"| Tier | Mode | Max {ROW_LABEL[signal]} passed | Read latency @ max (ms) | First failing fill | Read latency @ failing (ms) |")
            lines.append("|------|------|----------------|-------------------|--------------------|------------------------|")
            for run in sorted(sig_runs, key=lambda r: (tier_rank(r["tier"]), mode_rank(r["mode"]))):
                rp = run.get("readProbe") or {}
                steps = rp.get("steps") or []
                max_passed = run.get("maxFillLevelPassed", 0)
                latency_at_max = 0
                first_fail = ""
                latency_at_fail = ""
                for s in steps:
                    if s.get("passed"):
                        latency_at_max = int(s.get("readLatencyMs", 0))
                    elif not first_fail:
                        first_fail = f"{int(s.get('rowsIngested', 0)):,}"
                        latency_at_fail = f"{int(s.get('readLatencyMs', 0))}"
                lines.append(f"| {run['tier']} | {run['mode']} | {int(max_passed):,} | {latency_at_max} | {first_fail or '—'} | {latency_at_fail or '—'} |")
            lines.append("")

    lines.append("Generated by `benchmarks/scripts/chart.py`.")
    out.write_text("\n".join(lines) + "\n")


def main() -> int:
    if len(sys.argv) < 2:
        print("usage: chart.py <results-dir>", file=sys.stderr)
        return 2
    results_dir = Path(sys.argv[1])
    if not results_dir.is_dir():
        print(f"not a directory: {results_dir}", file=sys.stderr)
        return 2

    runs = load_runs(results_dir)
    if not runs:
        print(f"no usable *.json files in {results_dir}", file=sys.stderr)
        return 1

    for signal in SIGNALS:
        render_throughput_bar(runs, signal, results_dir / f"chart-{signal}.png")
        render_throughput_detail(runs, signal, results_dir / f"chart-detail-{signal}.png")
        render_readprobe(runs, signal, results_dir / f"chart-readprobe-{signal}.png")
    render_summary(runs, results_dir / "summary.md")
    print(f"wrote charts and summary.md into {results_dir}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
