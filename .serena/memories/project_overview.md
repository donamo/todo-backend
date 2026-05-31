# Project overview

- Project: `todo-backend` at `/Users/donamo/project/todo-backend`.
- Purpose: personal AI-assisted project and todo backend based on `todo-spec.md`.
- Backend direction follows `AGENTS.md` and `ARCHITECTURE.md`: REST only for infrastructure/auth, domain operations through GraphQL.
- Go module: `github.com/donamo/todo-backend`, Go 1.26.3.
- Current implementation:
  - `cmd/server/main.go`: HTTP server with graceful shutdown, optional PostgreSQL connection, embedded goose migrations.
  - `internal/app`: chi router, CORS, health/auth infrastructure routes, gqlgen `/graphql`, `/playground`.
  - `internal/auth`: Google OIDC login/callback/logout/me, gorilla cookie session, auth middleware.
  - `internal/config`: env helpers and server config loader.
  - `internal/logging`: slog setup, default DEBUG level.
  - `internal/graph`: gqlgen schema and resolvers for Epic, Project, Stage, Todo, Label, notes, dependencies, dashboard, progress, todo views.
  - `db/migrations`: initial PostgreSQL schema for users plus user-owned epics, projects, stages, todos, labels, notes, dependencies.
  - `db/queries`: sqlc source queries, scoped by authenticated `user_id`.
  - `internal/db`: sqlc generated code; do not edit manually.
  - `e2e`: startup, auth-required GraphQL, and authenticated GraphQL todo workflow tests.
- GraphQL `/graphql` requires an authenticated approved user. REST remains limited to `/health` and `/auth/*` through `/auth/me`.
- If `DATABASE_URL` is empty, `/health` works but auth/domain GraphQL cannot operate.
- AI status report exists as a stub (`ready: false`) until provider integration is added.