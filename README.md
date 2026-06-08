# GoInvoicing

A task and invoice management application with a Go backend, React frontend, and PostgreSQL database. Fully containerized with Docker, orchestrated with Docker Compose, and deployable to Kubernetes.

## Repository layout

```
backend/          Go API server (domain-driven design)
  cmd/server/     Entry point — wires dependencies and starts HTTP server
  internal/
    domain/       Invoice entity, business rules (no framework imports)
    services/     Business logic; depends on repository interfaces
    repositories/ Repository interfaces + implementations (memory/, postgres/)
    apperrors/    Shared sentinel errors (ErrNotFound, ErrConflict, …)
  api/            HTTP handlers and response helpers
  Dockerfile      Multi-stage build: golang:1.25-alpine → alpine:3.21

task-ui/          React + Vite frontend for task management
  Dockerfile      Multi-stage build: node:20-alpine → nginx:1.27-alpine
  nginx.conf.template  SPA routing + /api proxy (BACKEND_URL substituted at runtime)

k8s/              Kubernetes manifests (namespace: goinvoicing)
  namespace.yml
  postgres/       StatefulSet, Service, ConfigMap, Secret
  backend/        Deployment, Service, Ingress, ConfigMap, Secret
  task-ui/        Deployment, Service, Ingress

docker-compose.yml  Local 3-service orchestration (postgres, backend, task-ui)
.github/workflows/cicd.yml  GitHub Actions CI/CD pipeline
```

## Prerequisites

- Go 1.25+
- Node.js 20+ and npm
- Docker and Docker Compose
- [Supabase CLI](https://supabase.com/docs/guides/cli) (for migrations)
- kubectl + minikube (for Kubernetes deployment)

## Getting started

### 1. Clone and configure environment

```bash
git clone https://github.com/Sara-Stojilkova/GoInvoicing
cd GoInvoicing
cp backend/.env.example backend/.env   # fill in Supabase credentials
```

See [SUPABASE.md](SUPABASE.md) for the project ref, region, and where to find each credential.

### 2. Run the backend

```bash
cd backend
go build ./...
go run ./cmd/server
```

The server listens on `http://localhost:8080` by default.

### 3. Run the frontend

```bash
cd task-ui
npm install
npm run dev
```

The dev server starts at `http://localhost:5173`.

## Backend commands

All commands run from the `backend/` directory.

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run a specific test
go test ./internal/domain/... -run TestEvaluateInvoiceStatus

# Build
go build ./...
```

## Architecture

The backend follows domain-driven design with a clear layering rule: inner layers never import outer layers.

```
domain → (no imports)
repositories (interfaces) → domain
services → repositories interfaces, domain
api/handlers → services
cmd/server → everything (wiring only)
```

**Repository pattern** — interfaces live in `internal/repositories/`, concrete implementations in subdirectories (`memory/`, `postgres/`). Constructors return the interface, never the concrete type. The only place concrete types are wired together is `cmd/server/main.go`.

**Sentinel errors** — defined in `internal/apperrors/` so all layers can use them without circular imports. Repositories and services return errors; HTTP handlers translate them to status codes.

**Domain purity** — `internal/domain/` has no framework imports. Time is always passed as a parameter (`now time.Time`) so business logic is deterministic and testable.

## Testing

The project uses test-driven development. Tests are table-driven with fixed timestamps:

```go
tests := []struct {
    name  string
    input SomeType
    want  SomeType
}{
    {"case name", ...},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { ... })
}
```

## Database migrations

After writing or applying any migration, regenerate the TypeScript types before committing:

```bash
cd backend
supabase db push          # push migration to remote

cd ../task-ui
npm run gen:types         # regenerates src/database.types.ts from the live schema
```

`src/database.types.ts` must always reflect the current remote schema. Committing a migration without updating the types will cause type errors in the frontend.

## Supabase

See [SUPABASE.md](SUPABASE.md) for full details on the hosted project, credentials, and local development workflow.
