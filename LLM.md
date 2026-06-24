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
