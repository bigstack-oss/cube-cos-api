# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

All build tasks use [Task](https://taskfile.dev/) (`task` CLI). `GOWORK=off` is required for all Go commands.

```bash
# Vet (lint)
task vet                  # go vet ./...

# Test
task test                 # go test -v ./...

# Run a single test
CGO_ENABLED=1 GOWORK=off go test -v ./internal/... -run TestName

# Build binary (macOS: uses Docker; Linux: native)
task build

# Run locally (requires config file)
task runLocalDevApi       # uses configs/cube-cos-api.yaml

# Generate API docs (required before build; converts submodule YAML → JSON)
task generateApiDocs      # yq api/cube-cos-openapi/docs.yaml → api/docs.json

# Full pre-PR check
task check                # generateApiDocs + vet + go mod tidy
```

The `api/cube-cos-openapi/` directory is a git submodule. Run `git submodule update --init` if `api/docs.json` is missing or stale. `api/docs.json` is embedded at compile time via `//go:embed`.

## Architecture

CubeCOS API is a per-node HTTP service. Every CubeCOS cluster node runs one instance; nodes discover peers via MDNS. The node's **role** (`control`, `compute`, `storage`, `control-converged`, `moderator`, `edge-core`) determines which API endpoints are registered at startup.

### Layer overview

```
cmd/main.go
  └── internal/runtime        ← HTTP server, Gin router, middleware, dependency init
        └── internal/apis     ← Role-based handler registration
              └── internal/apis/v1/handlers/*  ← HTTP handlers (thin; delegate to helpers)
                    └── internal/cubecos       ← Business logic, CLI wrappers, external clients
                          └── internal/definition/v1  ← Domain types, DTOs, constants
```

Background work runs in **operators** (`internal/operators/v1/*/`) that implement `service.Operator` (Init/Run/Stop). They use Kubernetes-style workqueues and are started after the HTTP server via `micro.AfterStart`. API handlers enqueue work; operators dequeue and execute.

### Key conventions

- **Handler pattern**: each module in `internal/apis/v1/handlers/` has `handlers.go` (route + HTTP func) and `helper.go` (business logic). Keep handlers thin.
- **Role registration**: `internal/runtime/router.go:prepareApiHandlersByRole()` calls `apis.RegisterHandlersToRoles(module, handlers, roles...)`. Add new modules here.
- **System operations**: anything touching the OS, OpenStack, or cluster goes through `internal/cubecos/` which wraps `hex_sdk`, `hex_config`, and `hex_install` CLI tools.
- **SSE**: real-time endpoints support `?watch=true` returning `text/event-stream`.
- **Cross-node calls**: control nodes proxy to compute/storage nodes using URLs built by `internal/definition/v1/nodes/url.go`.

### Auth

Three auth paths checked in order per request (see `internal/runtime/router.go:verifyAuthToken`):
1. Internal node-to-node token (`Node:` header + `Authorization: Bearer <token>`)
2. OpenStack token (`X-Auth-Token`)
3. OIDC token via Keycloak (`Authorization: Bearer <jwt>`)
4. Falls back to SAML 2.0 SSO redirect

### External dependencies

Initialized in `internal/runtime/dependency.go`: MongoDB (primary store), InfluxDB (metrics), OpenStack, Keycloak (OIDC), Kubernetes (k3s), AWS S3. Config is read from `/etc/cube/api/cube-cos-api.yaml` in production; use `configs/cube-cos-api.yaml.template` as the local dev starting point.

### PR checklist

Per `.github/pull_request_template.md`: update API docs (`task generateApiDocs`) and verify the API works before submitting.
