import { Link2, ArrowRight } from "lucide-react";

const traces = [
  {
    service: "Backend API",
    serviceBg: "bg-zinc-100 text-zinc-700",
    method: "GET",
    path: "/api/test-error",
    status: "500",
    statusColor: "text-red-600",
    duration: "72ms",
    badge: "Exception",
  },
  {
    service: "React Frontend",
    serviceBg: "bg-zinc-100 text-zinc-700",
    method: null,
    path: "Error: GET /api/test-error failed: 500 Internal Server Error",
    status: null,
    statusColor: "",
    duration: null,
    badge: "Exception",
  },
];

export function DistributedTraceVisual() {
  return (
    <div className="rounded-xl border border-zinc-200 bg-white overflow-hidden">
      <div className="px-5 py-4 border-b border-zinc-100">
        <div className="flex items-center gap-2 mb-1">
          <Link2 className="w-4 h-4 text-zinc-400" />
          <span className="text-base font-semibold text-zinc-900">
            Distributed Trace
          </span>
        </div>
        <p className="text-sm text-zinc-500">
          This trace spans across multiple services
        </p>
      </div>

      <div className="divide-y divide-zinc-100">
        {traces.map((t, i) => (
          <div
            key={i}
            className="flex items-center gap-4 px-5 py-4"
          >
            <span
              className={`shrink-0 inline-flex items-center px-3 py-1 rounded-full text-xs font-medium ${t.serviceBg}`}
            >
              {t.service}
            </span>

            <div className="flex-1 min-w-0 flex items-center gap-2 text-sm font-mono text-zinc-800 truncate">
              {t.method && (
                <span className="font-semibold">{t.method}</span>
              )}
              <span className="truncate">{t.path}</span>
              {t.status && (
                <span className={`font-semibold ${t.statusColor}`}>
                  {t.status}
                </span>
              )}
              {t.duration && (
                <span className="text-zinc-400">{t.duration}</span>
              )}
            </div>

            <span className="shrink-0 inline-flex items-center px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-red-600 text-white">
              {t.badge}
            </span>

            <span className="shrink-0 flex items-center gap-1 text-sm font-medium text-zinc-500 hover:text-zinc-900 cursor-pointer transition-colors">
              View <ArrowRight className="w-3.5 h-3.5" />
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
