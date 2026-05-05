# Server

Vanilla Go + Gin + SQLite. No Traceway. Port 8080.

```bash
go mod tidy
go run .
```

## Endpoints

```
GET    /healthz
GET    /api/products
GET    /api/products/:id     (id=42 PANICS — Bug 1)
POST   /api/cart             { productId, qty }
GET    /api/cart
POST   /api/promo            { code }
POST   /api/checkout         (~4 seconds — Bug 2)
```

## Files

- `main.go` — Gin entry, middleware, route registration, startup banner
- `handlers.go` — endpoint handlers + the `checkout` step helpers
- `db.go` — SQLite open + seed (10 products, including id=42; cart table)

`ecommerce.db` is created in the working directory on first boot. Re-seed is idempotent.
