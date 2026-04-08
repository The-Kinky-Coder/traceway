import { withRoute } from "@/lib/with-route";

// Outgoing fetch() is auto-instrumented by @opentelemetry/instrumentation-undici
export const GET = withRoute("/nextjs/api/external", async () => {
  const res = await fetch("https://jsonplaceholder.typicode.com/posts/1");
  const post = await res.json();
  return Response.json(post);
});
