# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

YABA (Yet Another Budgeting App) is a self-hosted budgeting application. Backend is Go + PostgreSQL with a GraphQL API generated via `gqlgen`. Frontend is React, served as static files from the Go binary.

## Commands

```bash
# Backend
make build       # Generate GraphQL code + compile binary
make test        # Run all Go tests
make cover       # Tests with coverage (race detector, atomic mode)
make lint        # Run golangci-lint with auto-fix
make graphql     # Regenerate GraphQL server/model code from schema

# Frontend
make start-ui    # Start React dev server
make build-ui    # Build and package React UI (creates dist.tar.gz)

# Docker
make docker      # Build Docker image (requires make build-ui first)

# Single test
go test ./internal/handlers/... -run TestSpecificName
go test ./... -run TestSpecificName -v
```

## Architecture

### Request Flow

```
HTTP Request
  → SessionInterceptor (decodes `sid` cookie, loads user into context)
  → /graphql          → GraphQL resolvers in handlers/schema_graph_resolver.go
  → /api/login|register|logout → Auth handlers in internal/auth/
  → /api/export/expenditure    → Export handler
  → /* (React routes)          → Serves embedded React index.html
```

### Layer Responsibilities

| Layer | Location | Role |
|-------|----------|------|
| GraphQL schema | `graph/schema.graphqls` | Source of truth; run `make graphql` after changes |
| Generated code | `graph/model/`, `graph/server/` | Do not edit directly |
| Resolvers | `internal/handlers/schema_graph_resolver.go` | Implements generated interfaces |
| HTTP setup | `internal/handlers/` | Route wiring, middleware |
| DAOs | `internal/database/` | SQL via `squirrel` + `pgxscan` |
| Domain models | `internal/model/` | Business types with GraphQL conversion methods |
| Auth | `internal/auth/` | Session token creation, cookie handling, middleware |
| Context utils | `internal/ctxutil/` | Typed getters/setters for user/session in `context.Context` |
| Error types | `errors/` | `InvalidInput`, `InvalidState`, `NoSuchElement`, `Unauthorized` |

### Authentication

Sessions use a `sid` cookie containing a hex-encoded token. The token is stored in the `token` PostgreSQL table with an expiration. `SessionInterceptor` middleware validates the cookie and injects the user into context on every request. GraphQL endpoints require auth via `auth.NewAuthRequired()`.

### Database

- Migrations in `migrations/` managed by `golang-migrate/migrate`; run automatically on startup
- Queries built with `Masterminds/squirrel`, scanned with `scany/v2`
- Connection pooling via `jackc/pgx/v5` pool

### Testing

Tests use `testcontainers-go` to spin up a real PostgreSQL instance — no mocks for the database layer. Test helpers live in `internal/test/`.

## Development Notes

- After editing `graph/schema.graphqls`, run `make graphql` to regenerate; then implement any new resolver stubs in `internal/handlers/schema_graph_resolver.go`.
- See `docs/DEVELOPMENT.md` for DB migration workflow and GraphQL code generation details.
- Local environment: `docker-compose.yml` provides Postgres + the app server.
