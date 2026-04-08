# Add Traceway to a Project

Add OpenTelemetry tracing to an existing project so it reports endpoints, spans, and errors to a Traceway instance.

## Step 1: Identify the Framework

Detect the framework by reading `package.json` (Node.js), `go.mod` (Go), `composer.json` (PHP), or asking the user. Then follow the correct guide below.

## Step 2: Follow the Framework-Specific Guide

### Hono (Node.js)
Follow `skills/add-traceway-to-hono-project.md` in this repo. Uses `@hono/otel` middleware — do NOT use `@opentelemetry/instrumentation-http` (it doesn't work with Hono's ESM imports on Node 22+).

### Express (Node.js)
- Install: `@opentelemetry/sdk-node @opentelemetry/auto-instrumentations-node @opentelemetry/exporter-trace-otlp-http @opentelemetry/exporter-metrics-otlp-http`
- Create `instrumentation.js` at project root with `NodeSDK` + `getNodeAutoInstrumentations()` (keep `instrumentation-http` enabled — Express is CJS so it works)
- No app code changes needed — auto-instrumentation captures routes, status codes, errors
- Start with `node --import ./instrumentation.js server.js`
- Full docs: `docs/pages/client/node-sdk/index.mdx`

### NestJS (Node.js)
- Install: same OTel packages as Express
- Create `instrumentation.ts` at project root with `NodeSDK` + `getNodeAutoInstrumentations()` (keep `instrumentation-http` enabled — NestJS uses Express/Fastify internally, both CJS)
- No app code changes needed — auto-instrumentation handles everything
- Start with `node --require ./instrumentation.js dist/main.js` (must load before NestJS boots)
- Full docs: `docs/pages/client/nestjs/index.mdx`

### Next.js (Node.js)
Follow `skills/add-traceway-to-nextjs-project.md` in this repo. Requires a `withRoute()` wrapper for every API route handler (Next.js doesn't use Express, so `http.route` is never set automatically). Supports Prisma auto-instrumentation for database queries.

### Gin / Chi / Fiber / FastHTTP / stdlib (Go)
- Install the framework-specific middleware: `go get go.tracewayapp.com/tracewaygin` (or `tracewaychi`, `tracewayfiber`, `tracewayfasthttp`, `tracewayhttp`)
- Add middleware: `r.Use(tracewaygin.New("token@http://traceway:8082/api/report"))`
- Reports via Traceway's native protocol (`/api/report`), not OTel
- Full docs: `docs/pages/client/gin-middleware/index.mdx` (or the corresponding framework directory)

### Symfony (PHP)
- Install: `composer require traceway/opentelemetry-symfony open-telemetry/exporter-otlp php-http/guzzle7-adapter`
- Configure via `.env` with `OTEL_*` variables
- Add `\OpenTelemetry\SDK\SdkAutoloader::autoload()` to `public/index.php`
- Full docs: `docs/pages/client/symfony/index.mdx`

### React / Vue / Svelte / jQuery (Frontend)
- Install the framework-specific Traceway SDK: `npm install @tracewayapp/react` (or `@tracewayapp/vue`, `@tracewayapp/svelte`, `@tracewayapp/jquery`)
- These are client-side SDKs that report to `/api/report`, not OTel
- Full docs: `docs/pages/client/react/index.mdx` (or the corresponding framework directory)

### Cloudflare Workers
- Uses Cloudflare's built-in OTLP export, not the Node SDK
- Full docs: `docs/pages/client/cloudflare/index.mdx`

### Any Other Language (Generic OTel)
- Use any OpenTelemetry SDK for the language
- Export via OTLP/HTTP to `https://<traceway-instance>/api/otel/v1/traces` and `/v1/metrics`
- Set `Authorization: Bearer <project-token>` header
- Full docs: `docs/pages/client/otel/index.mdx`

## Common Across All Node.js Frameworks

- **Traceway URL**: `https://<instance>/api/otel/v1/traces` and `/v1/metrics`
- **Auth header**: `Authorization: Bearer <project-token>`
- **Environment variables**: `TRACEWAY_URL` and `TRACEWAY_TOKEN` (or standard `OTEL_*` vars)
- **What maps to what in Traceway**:
  - Root SERVER span → **Endpoint**
  - Root CONSUMER span → **Task**
  - Child span → **Span**
  - Exception event on any span → **Issue**
- **Auto-instrumented child spans** (CJS packages only): `pg`, `mysql2`, `mongodb`, `ioredis`, `redis`, outgoing `fetch()` via `instrumentation-undici`
- **Not auto-instrumented**: SQLite (`better-sqlite3`), custom business logic — use `tracer.startActiveSpan()` manually
