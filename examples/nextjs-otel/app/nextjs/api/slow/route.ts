import { withRoute } from "@/lib/with-route";
import { trace } from "@opentelemetry/api";

const tracer = trace.getTracer("nextjs-otel-example");

export const GET = withRoute("/nextjs/api/slow", async () => {
  await tracer.startActiveSpan("slow-operation", async (span) => {
    await new Promise((resolve) => setTimeout(resolve, 300));
    span.end();
  });
  return Response.json({ message: "Slow response" });
});
