import Link from "next/link";
import Image from "next/image";
import {
  ArrowRight,
  Video,
  ScrollText,
  Network,
  BarChart3,
  Workflow,
  Bug,
  XCircle,
  CheckCircle2,
} from "lucide-react";

import { Eyebrow } from "@/components/eyebrow";
import { Chip } from "@/components/chip";
import { PillarCard } from "@/components/pillar-card";
import { FinalCTA } from "@/components/final-cta";
import { Terminal } from "@/components/terminal";
import { StatsStrip } from "@/components/stats-strip";
import { TwoTrack } from "@/components/two-track";
import { HeroEmailCTA } from "@/components/hero-email-cta";

export default function Home() {
  return (
    <main className="relative">
      {/* HERO — dash0-style centered block: chip, title, subtitle, email form, book-a-demo helper */}
      <section className="hero gridbg">
        <div className="wrap">
          <div className="text-center max-w-3xl mx-auto flex flex-col items-center">
            <Chip variant="ok">OpenTelemetry-native</Chip>
            <h1 className="mt-6">
              Observability that <em>closes the loop.</em>
            </h1>

            <div className="mt-10 w-full">
              <HeroEmailCTA />
            </div>
          </div>
        </div>
      </section>

      {/* PRODUCT SHOWCASE — 6 pillars in the branded "Traceway" box */}
      <section className="pb-24">
        <div className="wrap max-w-5xl mx-auto">
          <div id="dashboard-mount" className="pillar-wrap">
            <div className="pillar-wrap-head">
              <span className="pillar-wrap-brand">
                <Image
                  src="/images/logo.png"
                  alt="Traceway"
                  width={120}
                  height={24}
                  priority
                />
              </span>
              <span className="pillar-wrap-tag">
                <span className="dot" aria-hidden />
                observability · live
              </span>
            </div>
            <div className="pillar-wrap-body">
              <div className="pillars-all">
                <PillarCard
                  icon={ScrollText}
                  title="Logs"
                  description="Structured, trace-linked, sub-second search."
                  href="/product/logs"
                  color="a2"
                />
                <PillarCard
                  icon={Network}
                  title="Traces"
                  description="End-to-end span waterfalls across every service."
                  href="/product/traces"
                  color="a1"
                />
                <PillarCard
                  icon={BarChart3}
                  title="Metrics"
                  description="Host, runtime, custom — any dimension, any chart."
                  href="/product/metrics"
                  color="ok"
                />
                <PillarCard
                  icon={Video}
                  title="Session replay"
                  description="Web DOM capture + exception attach."
                  href="/product/session-replay"
                  color="a2"
                />
                <PillarCard
                  icon={Bug}
                  title="Exceptions / Stack Traces"
                  description="Grouped, normalized, paired with replay."
                  href="/product/stack-traces"
                  color="a3"
                />
                <PillarCard
                  icon={Workflow}
                  title="AI tracing"
                  description="LLM cost, tokens, conversations."
                  href="/product/ai-tracing"
                  color="a4"
                />
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* SMB POSITIONING — 3-way comparison: Enterprise SaaS / DIY OSS / Traceway */}
      <section className="two-track" id="otel">
        <div className="wrap">
          <div className="two-track-head">
            <Eyebrow>For teams without an SRE</Eyebrow>
            <h2>
              Everything you&apos;d build or buy.{" "}
              <em>Stitched by one trace ID.</em>
            </h2>
            <p className="muted mt-4 max-w-[700px]">
              Most teams either pay enterprise tools, or spend months wiring up
              their own. Traceway gives you both — logs, traces, metrics,
              session replay, stack traces, and AI — in one connected system,
              not six.
            </p>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
            {/* Column 1 — Enterprise SaaS route */}
            <div
              className="rounded-[12px] p-7 relative overflow-hidden"
              style={{
                background:
                  "linear-gradient(180deg, color-mix(in oklab, var(--ink-2) 70%, transparent), color-mix(in oklab, var(--ink-1) 50%, transparent))",
                border: "1px solid var(--hair)",
              }}
            >
              <Eyebrow>The enterprise route</Eyebrow>
              <h3
                className="mt-3"
                style={{
                  color: "var(--fg-1)",
                  fontSize: 19,
                  letterSpacing: "-0.015em",
                }}
              >
                Pay Datadog + Sentry + PagerDuty.
              </h3>
              <ul className="mt-5 space-y-3">
                {[
                  "Hire a senior SRE (~$200k+/year)",
                  "Datadog + Sentry + PagerDuty — 3 bills, 3 auth systems",
                  "Metered per event, per host, per seat — surprise overages",
                  "Proprietary SDKs per language, per vendor",
                ].map((item) => (
                  <li
                    key={item}
                    className="flex items-start gap-3 text-[14px]"
                    style={{ color: "var(--fg-2)", lineHeight: 1.5 }}
                  >
                    <XCircle
                      className="mt-0.5 flex-shrink-0"
                      style={{ color: "var(--crit)" }}
                      size={18}
                    />
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
            </div>

            {/* Column 2 — DIY open-source route */}
            <div
              className="rounded-[12px] p-7 relative overflow-hidden"
              style={{
                background:
                  "linear-gradient(180deg, color-mix(in oklab, var(--ink-2) 70%, transparent), color-mix(in oklab, var(--ink-1) 50%, transparent))",
                border: "1px solid var(--hair)",
              }}
            >
              <Eyebrow>The DIY route</Eyebrow>
              <h3
                className="mt-3"
                style={{
                  color: "var(--fg-1)",
                  fontSize: 19,
                  letterSpacing: "-0.015em",
                }}
              >
                Glue 6 open-source tools together.
              </h3>
              <ul className="mt-5 space-y-3">
                {[
                  "Prometheus + Grafana + Loki + Tempo + Alertmanager",
                  "OTel Collector on top + errbit or self-hosted Sentry",
                  "Cardinality, retention, upgrades — weekly ops work",
                  "Session replay + source maps? Good luck wiring it.",
                ].map((item) => (
                  <li
                    key={item}
                    className="flex items-start gap-3 text-[14px]"
                    style={{ color: "var(--fg-2)", lineHeight: 1.5 }}
                  >
                    <XCircle
                      className="mt-0.5 flex-shrink-0"
                      style={{ color: "var(--crit)" }}
                      size={18}
                    />
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
            </div>

            {/* Column 3 — With Traceway (highlighted) */}
            <div
              className="rounded-[12px] p-7 relative overflow-hidden"
              style={{
                background:
                  "radial-gradient(500px 280px at 90% 0%, color-mix(in oklab, var(--ok) 14%, transparent), transparent 60%), linear-gradient(180deg, var(--ink-3), var(--ink-2))",
                border: "1px solid color-mix(in oklab, var(--ok) 30%, var(--hair-2))",
                boxShadow:
                  "0 20px 40px -20px color-mix(in oklab, var(--ok) 30%, transparent)",
              }}
            >
              <Eyebrow>With Traceway</Eyebrow>
              <h3
                className="mt-3"
                style={{
                  color: "var(--fg-0)",
                  fontSize: 19,
                  letterSpacing: "-0.015em",
                }}
              >
                One system. One trace ID. Every surface stitched.
              </h3>
              <ul className="mt-5 space-y-3">
                {[
                  "Logs + traces + metrics + replay + stacks + AI — bundled",
                  "MIT-licensed — no BSL, no open-core asterisks",
                  <>
                    <code>docker compose up -d</code> — 90-second install
                  </>,
                  "Click a log, see its span, see the replay, see the exception",
                ].map((item, i) => (
                  <li
                    key={i}
                    className="flex items-start gap-3 text-[14px]"
                    style={{ color: "var(--fg-1)", lineHeight: 1.5 }}
                  >
                    <CheckCircle2
                      className="mt-0.5 flex-shrink-0"
                      style={{ color: "var(--ok)" }}
                      size={18}
                    />
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* COST — brief, left-aligned */}
      <section className="py-24" id="cost">
        <div className="wrap">
          <div className="max-w-3xl">
            <Eyebrow>Pricing that doesn&apos;t lie to you</Eyebrow>
            <h2 className="mt-4">
              A fraction of the cost. <em>None of the asterisks.</em>
            </h2>
            <p className="muted mt-4 max-w-[640px]">
              Traceway runs on ClickHouse columnar storage — 1M daily events
              compresses to ~2GB/month. Fixed monthly tiers, no per-event
              gouging, no overage invoices at 2am.
            </p>
            <div className="mt-6 flex flex-wrap gap-3">
              <Link href="/cloud" className="btn btn-accent">
                See pricing
                <ArrowRight className="h-4 w-4" />
              </Link>
              <Link href="https://docs.tracewayapp.com" className="btn btn-ghost">
                Self-host for free
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* DEPLOY — stats strip + terminal */}
      <section className="py-24">
        <div className="wrap grid gap-14 md:grid-cols-[1fr_1.1fr] items-center">
          <div>
            <Eyebrow>Your data. Your metal.</Eyebrow>
            <h2 className="mt-4">
              Self-host in <em>90 seconds flat.</em>
            </h2>
            <p className="muted mt-4 max-w-[460px]">
              MIT licensed. No BSL. No &ldquo;open core.&rdquo; Every feature
              Traceway Cloud has, your cluster has. Point an OTLP exporter at
              it and you&apos;re in business.
            </p>
            <StatsStrip
              stats={[
                { num: "<em>0s</em>", label: "Config required" },
                { num: "100%", label: "Feature parity" },
                { num: "MIT", label: "License" },
              ]}
            />
          </div>
          <Terminal
            title="bash · traceway.sh · 80×24"
            lines={[
              {
                ln: "1",
                type: "tx",
                content: (
                  <>
                    <span className="cmd">$</span> git clone github.com/tracewayapp/traceway
                  </>
                ),
              },
              {
                ln: "2",
                type: "tx",
                content: (
                  <>
                    <span className="cmd">$</span> cd traceway &amp;&amp; docker compose up -d
                  </>
                ),
              },
              { ln: "3", type: "mute", content: "# pulling images…" },
              { ln: "4", type: "mute", content: "# starting clickhouse · postgres · collector" },
              { ln: "5", type: "ok", content: "# ✓ dashboard at http://localhost:3000" },
              {
                ln: "6",
                type: "tx",
                content: (
                  <>
                    <span className="cmd">$</span>
                  </>
                ),
              },
            ]}
            showCursor
          />
        </div>
      </section>

      {/* DETECT → RESOLVE */}
      <section className="py-24">
        <div className="wrap">
          <div className="two-track-head">
            <Eyebrow>Why it matters</Eyebrow>
            <h2 className="mt-3">
              Customers don&apos;t complain — they quit.{" "}
              <em>We stop the bleeding.</em>
            </h2>
            <p className="muted mt-4 max-w-[640px]">
              Your users won&apos;t open a ticket when something breaks —
              they&apos;ll close the tab. Traceway catches the error, the
              session replay, and the exact failing span before they bounce.
            </p>
          </div>
          <TwoTrack
            detect={{
              badge: "01 · Detect",
              title: "Surface what actually matters.",
              description:
                "Impact Score ranks every endpoint by five SLIs and bubbles the worst up first. Alerts route to Slack, GitHub, or webhook by threshold — no false-positive fatigue.",
              items: [
                "Impact Score across 5 service-level signals",
                "Per-endpoint slow threshold override",
                "Slack, GitHub, webhook, email routing",
                "Regression detection on new releases",
              ],
            }}
            resolve={{
              badge: "02 · Resolve",
              title: "Walk the full trace. Fix. Ship.",
              description:
                "Click an exception, see the frontend replay, the cross-service trace, the exact span that threw, and the source-mapped stack. Context-switching is the bug.",
              items: [
                "Frontend replay linked to backend errors",
                "Cross-service distributed trace waterfalls",
                "Source-mapped stack traces (webpack, esbuild, Vite)",
                "SHA-256 grouped duplicates into one ranked issue",
              ],
            }}
          />
        </div>
      </section>

      {/* Final CTA */}
      <FinalCTA
        title={
          <>
            Detect. Replay. <em>Resolve.</em>
          </>
        }
        description="Start for free. Self-host whenever you want. Book a demo if you'd like a walkthrough."
        primary={{
          label: "Start for free",
          href: "https://cloud.tracewayapp.com/register",
        }}
        secondary={{
          label: "Book a demo",
          href: "/contact",
        }}
      />
    </main>
  );
}
