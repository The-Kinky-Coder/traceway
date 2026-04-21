import Link from "next/link";
import { ArrowRight, BarChart3 } from "lucide-react";

import { Chip } from "@/components/chip";
import { SectionHead } from "@/components/section-head";
import { FeatureRow } from "@/components/feature-row";
import { FaqList } from "@/components/faq-list";
import { FinalCTA } from "@/components/final-cta";
import { AuroraBackground } from "@/components/aurora-background";

export default function MetricsPage() {
  return (
    <main className="relative">
      <section className="hero hero-product gridbg relative">
        <AuroraBackground variant="hero" />
        <div className="wrap relative z-10">
          <Chip variant="ok">
            <BarChart3 className="h-3 w-3 inline mr-1" />
            Metrics
          </Chip>
          <h1 className="mt-6">
            Measure what matters, <em>without the bill shock.</em>
          </h1>
          <p className="hero-sub">
            Application metrics via OpenTelemetry, automatic server metrics,
            and flexible widget dashboards — all included, with no per-metric
            billing and no surprise overages.
          </p>
          <div className="hero-cta-row">
            <Link href="https://docs.tracewayapp.com" className="btn btn-accent">
              Get Started <ArrowRight className="h-4 w-4" />
            </Link>
            <Link href="https://cloud.tracewayapp.com/register" className="btn btn-ghost">
              Try Traceway Cloud
            </Link>
          </div>
        </div>
      </section>

      {/* Application metrics */}
      <section className="wrap">
        <FeatureRow
          eyebrow="Application metrics"
          title={
            <>
              Application metrics <em>via OpenTelemetry</em>
            </>
          }
          description="Emit Counter, Gauge, and Histogram metrics through the OpenTelemetry SDK you already have. Traceway ingests OTLP natively, preserves units and dimensional tags, and bills nothing per metric."
          bullets={[
            "OTLP/HTTP and OTLP/gRPC ingestion",
            "Counter / Gauge / Histogram preserved natively",
            "Dimensional tags become facet filters",
            "No per-metric billing",
          ]}
          image={{ src: "/images/screenshot-4.png", alt: "Application metrics dashboard" }}
        />
      </section>

      {/* Server metrics */}
      <section className="wrap">
        <FeatureRow
          reverse
          eyebrow="Server metrics"
          title="CPU, memory, goroutines, GC — automatic"
          description="The Traceway SDK emits runtime metrics every 10 seconds without any configuration. See host health alongside application metrics in a single view."
          bullets={[
            "CPU usage percentage",
            "Memory (allocated, heap, used%)",
            "Goroutine count + heap object count",
            "GC cycles and pause time",
            "Zero-config, always on",
          ]}
          image={{ src: "/images/screenshot-4.png", alt: "Server metrics dashboard" }}
        />
      </section>

      {/* Widget groups */}
      <section className="wrap">
        <FeatureRow
          eyebrow="Widget groups"
          title={
            <>
              Dashboards that match <em>your team&apos;s mental model</em>
            </>
          }
          description="Pick metrics, pick charts, group them into widget pages. No query language required; filters, tag breakdowns, and rollups are all declarative."
          bullets={[
            "Drag-to-add charts",
            "Group widgets by feature, service, or team",
            "Per-metric filters and rollups",
            "Set default dashboards per organization",
          ]}
          image={{ src: "/images/screenshot-3.png", alt: "Widget groups dashboard" }}
        />
      </section>

      <FinalCTA
        title={
          <>
            Ship metrics <em>in 5 minutes</em>
          </>
        }
        description="Application + server metrics. Included on every plan. No per-metric billing."
        primary={{
          label: "Read the Metrics docs",
          href: "https://docs.tracewayapp.com",
        }}
        secondary={{
          label: "Try Live Demo",
          href: "https://cloud.tracewayapp.com/login?email=demo@tracewayapp.com&password=demoaccount!",
        }}
      />

      <section className="wrap pt-10 pb-24">
        <div className="max-w-3xl mx-auto">
          <SectionHead align="center" eyebrow="FAQ" title="Questions about metrics" />
          <div className="mt-4">
            <FaqList
              items={[
                {
                  q: "How do I emit an application metric?",
                  a: (
                    <>
                      <p>
                        Point any OpenTelemetry SDK at{" "}
                        <code>/api/otel/v1/metrics</code> — OTLP/HTTP and
                        OTLP/gRPC are both supported natively. Counter, Gauge,
                        and Histogram metric types are ingested as-is, and
                        dimensional tags become facet filters in the dashboard.
                      </p>
                      <p>
                        If you&apos;re instrumenting from scratch, the
                        OpenTelemetry metrics SDK is the recommended path for
                        every language we support.
                      </p>
                    </>
                  ),
                },
                {
                  q: "What server metrics are collected automatically?",
                  a: "The Traceway SDK automatically emits CPU usage, memory usage (allocated, heap, used%), goroutine count, heap object count, GC cycle count, and GC pause time every 10 seconds. No configuration required — they show up in the Metrics dashboard out of the box.",
                },
                {
                  q: "Do custom metrics count toward my event limit?",
                  a: "No. Metrics are included at no additional event cost — only issues, HTTP requests, and background tasks count toward your event limit. This means you can emit thousands of custom metrics without worrying about billing.",
                },
                {
                  q: "Can I query metrics by tag or dimension?",
                  a: "Yes. Every tag becomes a facet you can filter on; widget groups let you build per-dimension chart panels. For example, a `plan` tag on a signups metric lets you chart signups broken down by plan, region, or tenant.",
                },
              ]}
            />
          </div>
        </div>
      </section>
    </main>
  );
}
