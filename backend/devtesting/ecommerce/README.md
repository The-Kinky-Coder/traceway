# Ecommerce Demo (YC Pitch)

A small ecommerce app — Go + Gin backend, React + Vite frontend, SQLite — with three intentional bugs. Used as the live demo target for the YC pitch: ship the bugs in vanilla form, then add Traceway on stage and watch each pillar light up.

**There is no Traceway integration in either project.** That gets added during the demo.

## Layout

```
ecommerce/
├── server/        Go + Gin + SQLite, port 8080
└── web/           React + Vite + TS, port 5173 (proxies /api → :8080)
```

## Run

Two terminals.

**Terminal 1 — server**
```bash
cd server
go mod tidy
go run .
```

**Terminal 2 — web**
```bash
cd web
npm install
npm run dev
```

Open http://localhost:5173.

## The three bugs

| # | Where | What | Pillar lit when Traceway is added |
| - | ----- | ---- | --------------------------------- |
| 1 | Backend | Visiting product `id=42` panics (nil-map assignment); `gin.Recovery()` returns 500 | **Exceptions** |
| 2 | Backend | `POST /api/checkout` takes ~4s — `lookupInventory` sleeps 3500ms | **Distributed Tracing + Logs** |
| 3 | Frontend | Cart's "Apply Promo" button reads `undefined.toUpperCase()` when the input is empty | **Session Replay** |

### How to trigger each

1. **Bug 1** — Products page → click any tile and look for the one labeled with id 42, or hit `curl http://localhost:8080/api/products/42`. Server logs a panic stack trace, returns 500.
2. **Bug 2** — Add a couple of items to the cart, click Checkout, click Place Order. Wait ~4s. Server stdout shows four `INFO`/`WARN` lines including `slow inventory lookup detected`.
3. **Bug 3** — Cart page → click **Apply Promo** without typing a code. In dev mode, React's red overlay appears. In a production build (`npm run build && npm run preview`), the cart subtree blanks.

## Where Traceway plugs in (the live-demo cheat sheet)

Three edits, total ~6 lines, no other changes needed.

### 1. Backend — `server/main.go`

```diff
 import (
     "github.com/gin-contrib/cors"
     "github.com/gin-gonic/gin"
+    tracewaygin "go.tracewayapp.com/tracewaygin"
 )
```

```diff
     r := gin.New()
     r.Use(gin.Logger(), gin.Recovery())
+    r.Use(tracewaygin.New("demo-token@http://localhost:8082/api/report"))
     r.Use(cors.New(...))
```

That's it for the server — every endpoint, panic, and request now reports.

### 2. Bug 2 — wrap the slow steps in spans (optional but visually striking)

In `server/handlers.go`, inside `checkout`, wrap each step:

```go
span := traceway.StartSpan(c.Request.Context(), "lookup_inventory")
lookupInventory(c.Request.Context())
span.End()
```

Replace `log.Printf("WARN ...")` with `traceway.CaptureMessageWithContext(c.Request.Context(), "WARN ...")` so the log breadcrumb attaches to the trace.

### 3. Frontend — `web/src/main.tsx`

```diff
 import { createRoot } from "react-dom/client";
+import { TracewayProvider, TracewayErrorBoundary } from "@tracewayapp/react";
 import App from "./App";
```

```diff
 createRoot(document.getElementById("root")!).render(
-    <StrictMode><App /></StrictMode>
+    <StrictMode>
+        <TracewayProvider connectionString="demo-token@http://localhost:8082/api/report">
+            <TracewayErrorBoundary fallback={<div>Something went wrong.</div>}>
+                <App />
+            </TracewayErrorBoundary>
+        </TracewayProvider>
+    </StrictMode>
 );
```

The error boundary catches Bug 3 automatically; the SDK auto-injects `traceway-trace-id` on every `fetch`, so trace IDs link the browser session replay → backend span → backend exception.

## What lights up after wiring

| Bug | Dashboard path |
| --- | --- |
| 1   | **Exceptions** → `assignment to entry in nil map` panic, request URL `/api/products/42`, full Go stack |
| 2   | **Performance / Endpoints** → `POST /api/checkout` → trace waterfall reveals `lookup_inventory` 3500ms span; trace breadcrumbs include the WARN line |
| 3   | **Exceptions** → `TypeError: Cannot read properties of undefined (reading 'toUpperCase')` from `Cart.tsx`, with rrweb session replay attached |
