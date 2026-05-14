#!/usr/bin/env python3
"""Render hardware-vs-breaking-point charts from a directory of loadgen JSONs.

Inputs: a directory containing N JSON files (one per matrix entry), each in
the schema emitted by benchmarks/loadgen/. Order-independent.

Outputs (written into the same directory):
  - chart.png         Bar: max sustainable EPS per tier, two bars per tier
                      (sqlite vs pgch). The headline.
  - chart-detail.png  Lines: read P99 (ms) vs ingest EPS, one line per
                      (tier, mode). Shows how reads degrade as ingest climbs.
  - summary.md        Markdown table for human consumption / docs source.
"""

import json
import sys
from collections import defaultdict
from pathlib import Path

import matplotlib

matplotlib.use("Agg")
import matplotlib.pyplot as plt

# Stable left-to-right ordering of tiers regardless of how the matrix runs.
TIER_ORDER = ["ccx13", "ccx23", "ccx33", "ccx43"]
MODE_ORDER = ["sqlite", "pgch"]
MODE_COLORS = {"sqlite": "#4f9fff", "pgch": "#ff9f4f"}
READ_P99_BUDGET_MS = 3000


def load_runs(results_dir: Path) -> list[dict]:
    runs = []
    for p in sorted(results_dir.glob("*.json")):
        with p.open() as fh:
            doc = json.load(fh)
        doc["_path"] = p
        runs.append(doc)
    return runs


def tier_rank(t: str) -> int:
    return TIER_ORDER.index(t) if t in TIER_ORDER else len(TIER_ORDER)


def mode_rank(m: str) -> int:
    return MODE_ORDER.index(m) if m in MODE_ORDER else len(MODE_ORDER)


def render_bar(runs: list[dict], out: Path) -> None:
    fig, ax = plt.subplots(figsize=(9, 5))
    tiers_present = sorted({r["tier"] for r in runs}, key=tier_rank)
    modes_present = sorted({r["mode"] for r in runs}, key=mode_rank)

    width = 0.8 / max(len(modes_present), 1)
    x = list(range(len(tiers_present)))

    for mi, mode in enumerate(modes_present):
        ys = []
        for tier in tiers_present:
            match = [r for r in runs if r["tier"] == tier and r["mode"] == mode]
            ys.append(match[0]["maxSustainableEps"] if match else 0)
        offsets = [xi + (mi - (len(modes_present) - 1) / 2) * width for xi in x]
        bars = ax.bar(offsets, ys, width=width, label=mode, color=MODE_COLORS.get(mode, "#777"))
        for b, y in zip(bars, ys):
            ax.annotate(f"{int(y):,}", xy=(b.get_x() + b.get_width() / 2, y),
                        xytext=(0, 3), textcoords="offset points",
                        ha="center", va="bottom", fontsize=9)

    ax.set_xticks(x)
    ax.set_xticklabels(tiers_present)
    ax.set_ylabel("Max sustainable traces/sec")
    ax.set_title("Traceway: max sustainable ingest by hardware tier\n"
                 f"(failure = ingest err > 5% OR any read p99 > {READ_P99_BUDGET_MS}ms)")
    ax.legend(title="DB mode")
    ax.grid(axis="y", linestyle=":", alpha=0.4)
    fig.tight_layout()
    fig.savefig(out, dpi=130)
    plt.close(fig)


def render_detail(runs: list[dict], out: Path) -> None:
    fig, ax = plt.subplots(figsize=(10, 6))

    for run in sorted(runs, key=lambda r: (tier_rank(r["tier"]), mode_rank(r["mode"]))):
        xs, ys = [], []
        for s in run.get("steps", []):
            if not s.get("reads"):
                continue
            xs.append(s["actualEps"])
            # Worst read p99 across all read endpoints — that's what defines the cliff.
            ys.append(max((rd.get("p99", 0) for rd in s["reads"].values()), default=0))
        if xs:
            label = f"{run['tier']} / {run['mode']}"
            ax.plot(xs, ys, marker="o", label=label)

    ax.axhline(READ_P99_BUDGET_MS, linestyle="--", color="#cc3333", alpha=0.6,
               label=f"P99 budget ({READ_P99_BUDGET_MS}ms)")
    ax.set_xlabel("Actual ingest rate (traces/sec)")
    ax.set_ylabel("Worst read endpoint P99 (ms)")
    ax.set_title("Read latency degradation under increasing ingest load")
    ax.set_xscale("log")
    ax.set_yscale("log")
    ax.grid(True, which="both", linestyle=":", alpha=0.4)
    ax.legend(loc="upper left", fontsize=9)
    fig.tight_layout()
    fig.savefig(out, dpi=130)
    plt.close(fig)


def render_summary(runs: list[dict], out: Path) -> None:
    lines = ["# Traceway hardware benchmark — summary", ""]
    lines.append(f"Runs analyzed: {len(runs)}")
    lines.append("")
    lines.append("| Tier | Mode | Max EPS | Ingest p99 @max (ms) | Worst read p99 @max (ms) | Steps |")
    lines.append("|------|------|---------|----------------------|---------------------------|-------|")

    for run in sorted(runs, key=lambda r: (tier_rank(r["tier"]), mode_rank(r["mode"]))):
        max_eps = int(run.get("maxSustainableEps", 0))
        last_pass = None
        for s in reversed(run.get("steps", [])):
            if s.get("passed"):
                last_pass = s
                break
        ingest_p99 = int(last_pass["ingest"]["p99"]) if last_pass else 0
        worst_read = 0
        if last_pass and last_pass.get("reads"):
            worst_read = int(max(rd.get("p99", 0) for rd in last_pass["reads"].values()))
        lines.append(f"| {run['tier']} | {run['mode']} | {max_eps:,} | {ingest_p99} | {worst_read} | {len(run.get('steps', []))} |")

    lines.append("")
    lines.append("Generated by `benchmarks/scripts/chart.py`. "
                 "Failure thresholds: ingest error rate > 5%, OR any read endpoint p99 > "
                 f"{READ_P99_BUDGET_MS}ms.")
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
        print(f"no *.json files in {results_dir}", file=sys.stderr)
        return 1

    render_bar(runs, results_dir / "chart.png")
    render_detail(runs, results_dir / "chart-detail.png")
    render_summary(runs, results_dir / "summary.md")
    print(f"wrote {results_dir}/chart.png, chart-detail.png, summary.md")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
