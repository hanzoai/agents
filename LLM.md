# Hanzo Agents

## Overview
Hanzo Agents - Multi-agent orchestration platform

## Tech Stack
- **Language**: TypeScript/JavaScript (apps, packages); Go (`control-plane/`)

## Control plane (Go)
`control-plane/` is the orchestration service. Its RPC is **native ZAP**, not
gRPC/protobuf: schemas are `.zap` (`proto/admin/*.zap`), wire is zapgen output
(`pkg/adminwire`, zero-copy `New`/`Wrap` — no marshal), and the admin surface
serves over `github.com/zap-proto/go/transport` (`internal/server/server_admin.go`).
Auth carries the API key in each request's capability field, checked in constant
time; with no key configured the surface binds loopback only. No gRPC, no
protobuf, no `.proto` remain.

### HTTP surface — zip
The REST/UI surface routes on **`github.com/zap-proto/zip`** (v1.10.0, over
`zap-proto/http`), registered in `internal/server/server.go:setupRoutes`. 130
routes: `/v1`, `/v1/ui`, `/v2/ui`, plus `/health`, `/metrics` and the UI static
tree. `Start` calls `App.Listen("http://:PORT")` — the scheme is explicit
because a **bare address binds ZAP, not HTTP**.

Handlers are still `gin.HandlerFunc` and are fronted, not rewritten, by the one
seam in `internal/server/ginbridge.go`:

- `ginHandler(h)` — wraps a gin handler, copying the matched zip route's params
  onto the gin context so `c.Param(...)` keeps working. The net/http adaptor
  underneath supports buffered, flushed (**SSE**) and hijacked (**WebSocket**)
  responses, so the streaming routes work unchanged.
- `ginChain(mws...)` — runs the gin middleware stack (CORS, access log, timeout,
  API-key auth) unmodified, preserving CORS origin echoing and the auth abort
  bodies byte-for-byte.

Two invariants worth knowing before editing routes:

- The `NoRoute` fallback is `App.All("/+", ...)`, **not** `"/*"` — `"/*"` also
  matches `/` and would shadow the root redirect.
- zip resolves routes by specificity (ServeMux-1.22 semantics), so registration
  order does not matter and ambiguous overlaps panic at boot.

`internal/server/routes_test.go` pins the exact 130-route table; it fails if a
route is added, dropped or renamed.

## Build & Run
```bash
npm install && npm run build
npm test
```

## Structure
```
agents/
  CHANGELOG.md
  LICENSE
  Makefile
  README.md
  VERSION
  apps/
  assets/
  biome.json
  control-plane/
  deployments/
  docs/
  examples/
  package-lock.json
  package.json
  packages/
```

## Key Files
- `README.md` -- Project documentation
- `package.json` -- Dependencies and scripts
- `Makefile` -- Build automation
