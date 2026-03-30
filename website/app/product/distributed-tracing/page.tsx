import Image from "next/image";
import Link from "next/link";

import { Badge } from "@/components/ui/badge";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import {
  ArrowRight,
  GitBranch,
  Video,
  FileCode,
  TrendingUp,
} from "lucide-react";

export default function DistributedTracingPage() {
  return (
    <main className="min-h-screen bg-white text-zinc-950 font-sans selection:bg-zinc-100 selection:text-zinc-900">
      {/* Hero Section */}
      <section className="relative pt-16 pb-20 overflow-hidden">
        <div className="absolute inset-0 -z-10 h-full w-full bg-white bg-[radial-gradient(#e5e7eb_1px,transparent_1px)] [background-size:16px_16px] [mask-image:radial-gradient(ellipse_50%_50%_at_50%_50%,#000_70%,transparent_100%)]"></div>
        <div className="container mx-auto px-4 text-center">
          <Badge
            variant="secondary"
            className="mb-4 bg-teal-50 text-teal-700 hover:bg-teal-100 px-2.5 py-0.5 border border-teal-100 text-xs font-normal rounded-full"
          >
            Distributed Tracing
          </Badge>
          <h1 className="text-4xl md:text-6xl font-bold tracking-tight mb-6 text-zinc-900">
            See the user behind <br />
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-teal-600 to-cyan-600">
              every backend error
            </span>
          </h1>
          <p className="text-zinc-600 text-lg md:text-xl max-w-2xl mx-auto mb-10 leading-relaxed font-medium">
            When a backend service throws an exception, Traceway shows you the
            user&apos;s session replay, the cross-service trace, and the exact
            span that failed. No log-digging. No guessing what happened.
          </p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
            <Link
              href="https://docs.tracewayapp.com"
              className="inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium transition-all cursor-pointer h-10 px-6 bg-zinc-900 text-white hover:bg-zinc-800 shadow-lg shadow-zinc-900/20"
            >
              Get Started <ArrowRight className="ml-2 h-4 w-4" />
            </Link>
            <Link
              href="http://cloud.tracewayapp.com/register"
              className="inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium transition-all cursor-pointer h-10 px-6 border border-zinc-200 bg-white hover:bg-zinc-50 text-zinc-900 shadow-sm"
            >
              Try Traceway Cloud
            </Link>
          </div>
        </div>
      </section>

      {/* Feature 1: Cross-Service Visibility */}
      <section className="py-24 bg-white border-y border-zinc-100">
        <div className="container mx-auto px-4 max-w-5xl">
          <div className="flex flex-col md:flex-row items-center gap-12 lg:gap-20">
            <div className="flex-1 space-y-6">
              <div className="w-12 h-12 bg-teal-50 rounded-2xl flex items-center justify-center">
                <GitBranch className="w-6 h-6 text-teal-600" />
              </div>
              <h2 className="text-2xl md:text-3xl font-bold text-zinc-900 tracking-tight">
                Trace requests across every service
              </h2>
              <p className="text-zinc-600 text-lg leading-relaxed">
                Follow a single user action from the browser click through your
                API gateway, backend services, and database calls. Traceway
                propagates trace context so every span connects into one
                distributed trace.
              </p>
              <ul className="space-y-3 pt-2">
                <li className="flex items-center gap-3 text-zinc-700">
                  <div className="w-1.5 h-1.5 rounded-full bg-teal-500"></div>
                  W3C Trace Context propagation
                </li>
                <li className="flex items-center gap-3 text-zinc-700">
                  <div className="w-1.5 h-1.5 rounded-full bg-teal-500"></div>
                  Cross-service span waterfall
                </li>
                <li className="flex items-center gap-3 text-zinc-700">
                  <div className="w-1.5 h-1.5 rounded-full bg-teal-500"></div>
                  Automatic context linking
                </li>
                <li className="flex items-center gap-3 text-zinc-700">
                  <div className="w-1.5 h-1.5 rounded-full bg-teal-500"></div>
                  Works with any OpenTelemetry-instrumented service
                </li>
              </ul>
            </div>
            <div className="flex-1 w-full relative">
              <div className="absolute inset-0 bg-gradient-to-tr from-teal-100/50 to-transparent rounded-3xl transform rotate-3 scale-105 -z-10"></div>
              <div className="relative rounded-xl overflow-hidden border border-zinc-200 bg-white">
                <Image
                  src="/images/screenshot-3.png"
                  alt="Distributed Trace Waterfall"
                  width={800}
                  height={600}
                  className="w-full h-auto"
                />
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Feature 2: Session Replay + Backend Errors */}
      <section className="py-24 bg-zinc-50/50">
        <div className="container mx-auto px-4 max-w-5xl">
          <div className="flex flex-col md:flex-row-reverse items-center gap-12 lg:gap-20">
            <div className="flex-1 space-y-6">
              <div className="w-12 h-12 bg-purple-50 rounded-2xl flex items-center justify-center">
                <Video className="w-6 h-6 text-purple-600" />
              </div>
              <h2 className="text-2xl md:text-3xl font-bold text-zinc-900 tracking-tight">
                See what the user did when the backend broke
              </h2>
              <p className="text-zinc-600 text-lg leading-relaxed">
                Traceway connects frontend session replays to backend exceptions.
                When your payment service returns a 500, you don&apos;t just see
                the stack trace &mdash; you see the user clicking
                &ldquo;Checkout&rdquo;, filling in their card, and hitting
                submit. The replay is linked to the distributed trace
                automatically.
              </p>
              <ul className="space-y-3 pt-2">
                <li className="flex items-center gap-3 text-zinc-700">
                  <div className="w-1.5 h-1.5 rounded-full bg-purple-500"></div>
                  Frontend replay linked to backend errors
                </li>
                <li className="flex items-center gap-3 text-zinc-700">
                  <div className="w-1.5 h-1.5 rounded-full bg-purple-500"></div>
                  Automatic correlation via trace ID
                </li>
                <li className="flex items-center gap-3 text-zinc-700">
                  <div className="w-1.5 h-1.5 rounded-full bg-purple-500"></div>
                  No manual reproduction needed
                </li>
                <li className="flex items-center gap-3 text-zinc-700">
                  <div className="w-1.5 h-1.5 rounded-full bg-purple-500"></div>
                  Works across browser and server
                </li>
              </ul>
            </div>
            <div className="flex-1 w-full relative">
              <div className="absolute inset-0 bg-gradient-to-tl from-purple-100/50 to-transparent rounded-3xl transform -rotate-3 scale-105 -z-10"></div>
              <div className="relative rounded-xl overflow-hidden border border-zinc-200 bg-white">
                <Image
                  src="/images/session-replay.png"
                  alt="Session Replay linked to Backend Error"
                  width={800}
                  height={600}
                  className="w-full h-auto"
                />
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Two-Card Section */}
      <section className="py-24 bg-white border-y border-zinc-100">
        <div className="container mx-auto px-4 max-w-5xl">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            <div className="rounded-2xl border border-zinc-200 bg-white p-8 space-y-5">
              <div className="w-12 h-12 bg-orange-50 rounded-2xl flex items-center justify-center">
                <FileCode className="w-6 h-6 text-orange-600" />
              </div>
              <h2 className="text-xl md:text-2xl font-bold text-zinc-900 tracking-tight">
                Source map stack trace resolution
              </h2>
              <p className="text-zinc-600 leading-relaxed">
                Minified JavaScript stack traces are resolved to original source
                files and line numbers automatically. Upload your source maps and
                every frontend error shows readable, actionable traces.
              </p>
              <ul className="space-y-3 pt-1">
                <li className="flex items-center gap-3 text-zinc-700">
                  <div className="w-1.5 h-1.5 rounded-full bg-orange-500"></div>
                  Automatic source map resolution
                </li>
                <li className="flex items-center gap-3 text-zinc-700">
                  <div className="w-1.5 h-1.5 rounded-full bg-orange-500"></div>
                  Original file names and line numbers
                </li>
                <li className="flex items-center gap-3 text-zinc-700">
                  <div className="w-1.5 h-1.5 rounded-full bg-orange-500"></div>
                  Works with webpack, esbuild, and Vite
                </li>
              </ul>
            </div>

            <div className="rounded-2xl border border-zinc-200 bg-white p-8 space-y-5">
              <div className="w-12 h-12 bg-blue-50 rounded-2xl flex items-center justify-center">
                <TrendingUp className="w-6 h-6 text-blue-600" />
              </div>
              <h2 className="text-xl md:text-2xl font-bold text-zinc-900 tracking-tight">
                Impact Score for distributed services
              </h2>
              <p className="text-zinc-600 leading-relaxed">
                The Impact Score extends across service boundaries. When an
                upstream service degrades, its impact propagates to every
                downstream consumer &mdash; so you fix the root cause, not the
                symptoms.
              </p>
              <ul className="space-y-3 pt-1">
                <li className="flex items-center gap-3 text-zinc-700">
                  <div className="w-1.5 h-1.5 rounded-full bg-blue-500"></div>
                  Cross-service impact propagation
                </li>
                <li className="flex items-center gap-3 text-zinc-700">
                  <div className="w-1.5 h-1.5 rounded-full bg-blue-500"></div>
                  Root cause identification
                </li>
                <li className="flex items-center gap-3 text-zinc-700">
                  <div className="w-1.5 h-1.5 rounded-full bg-blue-500"></div>
                  Automatic priority ranking
                </li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* FAQ Section */}
      <section className="py-24 bg-zinc-50 border-t border-zinc-100">
        <div className="container mx-auto px-4 max-w-3xl">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold mb-4 text-zinc-900 tracking-tight">
              Frequently Asked Questions
            </h2>
            <p className="text-zinc-600 text-lg">
              Common questions about distributed tracing with Traceway.
            </p>
          </div>

          <Accordion type="single" collapsible className="w-full">
            <AccordionItem value="item-1" className="border-b-zinc-200">
              <AccordionTrigger className="text-zinc-900 hover:text-zinc-700 hover:no-underline text-left">
                How does distributed tracing connect to session replay?
              </AccordionTrigger>
              <AccordionContent className="text-zinc-600 leading-relaxed">
                Traceway&apos;s frontend SDK generates a trace ID for each user
                interaction and passes it to your backend via the traceparent
                header. When the backend reports a span or exception with that
                trace ID, Traceway links the frontend session replay to the
                backend trace automatically.
              </AccordionContent>
            </AccordionItem>
            <AccordionItem value="item-2" className="border-b-zinc-200">
              <AccordionTrigger className="text-zinc-900 hover:text-zinc-700 hover:no-underline text-left">
                What protocols does Traceway use for trace propagation?
              </AccordionTrigger>
              <AccordionContent className="text-zinc-600 leading-relaxed">
                Traceway supports W3C Trace Context (traceparent/tracestate
                headers) and is compatible with any OpenTelemetry-instrumented
                service. If your services already propagate trace context,
                Traceway picks it up automatically.
              </AccordionContent>
            </AccordionItem>
            <AccordionItem value="item-3" className="border-b-zinc-200">
              <AccordionTrigger className="text-zinc-900 hover:text-zinc-700 hover:no-underline text-left">
                Do I need to instrument both frontend and backend?
              </AccordionTrigger>
              <AccordionContent className="text-zinc-600 leading-relaxed">
                For the full experience (session replay linked to backend
                traces), yes. The frontend SDK captures user interactions and the
                backend middleware captures server-side spans. Both connect via
                the shared trace ID. However, backend-only distributed tracing
                works independently.
              </AccordionContent>
            </AccordionItem>
            <AccordionItem value="item-4" className="border-b-zinc-200">
              <AccordionTrigger className="text-zinc-900 hover:text-zinc-700 hover:no-underline text-left">
                How does source map resolution work?
              </AccordionTrigger>
              <AccordionContent className="text-zinc-600 leading-relaxed">
                Upload your source maps to Traceway or configure your CI to do
                it automatically. When a frontend exception is reported with a
                minified stack trace, Traceway resolves it to the original file
                paths and line numbers before displaying it.
              </AccordionContent>
            </AccordionItem>
          </Accordion>
        </div>
      </section>
    </main>
  );
}
