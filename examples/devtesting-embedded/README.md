# devtesting-embedded

Distributed tracing demo with a Go backend and jQuery frontend. Shows two ways to integrate `@tracewayapp/jquery`:

## Quick Start (CDN page only)

```bash
go run .
```

Open http://localhost:8080/cdn — works immediately, no build step. jQuery and Traceway are loaded from CDN.

## Full Setup (both pages)

```bash
cd frontend && npm install && npm run build && cd ..
go run .
```

## Pages

| URL | Description | Build required? |
|-----|-------------|-----------------|
| http://localhost:8080 | Node-built version — `@tracewayapp/jquery` bundled via Vite into `static/app.js` | Yes (`npm install && npm run build` in `frontend/`) |
| http://localhost:8080/cdn | CDN version — loads `@tracewayapp/jquery@1.0.3` IIFE bundle from jsdelivr | No |
| http://localhost:8082 | Traceway dashboard | No |

Both pages have the same functionality: error button, success button, and a log showing trace IDs. Login to the dashboard with `admin@localhost.com` / `admin`.

## How it works

1. Go serves both HTML pages and the API endpoints on port 8080
2. Traceway backend runs on port 8082 with SQLite storage
3. Click "Call Backend Error Endpoint" to trigger a 500 error
4. The `traceway-trace-id` header links the jQuery frontend exception to the Go backend error
5. Both appear in the Traceway dashboard under their respective projects, connected by the distributed trace
