# todo-backend

Go backend for the personal todo/project planning system in `todo-spec.md`.

## Stack

- Go 1.26
- chi HTTP router
- gqlgen schema-first GraphQL
- sqlc + PostgreSQL
- goose-style SQL migrations

## Development

```sh
make test
go run ./cmd/server
```

If `DATABASE_URL` is not set, the server still starts and `/health` works, but GraphQL domain operations return a database configuration error.

## Endpoints

- `GET /health`
- `GET /auth/google/login`
- `GET /auth/callback/google`
- `POST /auth/logout`
- `GET /auth/me`
- `POST /graphql`
- `GET /playground`

REST is reserved for infrastructure and auth. Domain operations should be implemented through GraphQL.

## Domain API

Implemented through GraphQL:

- Epic, Project, Stage, Todo, Label CRUD basics
- Inbox, open todos, overdue todos, next actions, done todos
- Project and stage progress
- Dashboard counts
- Project notes
- Todo labels
- Todo dependencies
- Next action uniqueness per project
- AI magic text proposal generation and explicit proposal acceptance

`aiStatusReport` is available as an explicit stub and returns `ready: false` until AI provider configuration is added.

GraphQL domain operations require an authenticated and approved Google session. REST is limited to `/health` and auth endpoints through `/auth/me`.

## Environment

| Name | Default | Purpose |
|---|---:|---|
| `PORT` | `3000` | HTTP server port |
| `FRONTEND_URL` | `http://localhost:5173` | Allowed CORS origin |
| `LOG_LEVEL` | `DEBUG` | `DEBUG`, `INFO`, `WARN`, or `ERROR` |
| `HTTP_SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown timeout |
| `HTTP_READ_HEADER_TIMEOUT` | `5s` | HTTP read header timeout |
| `DATABASE_URL` | none | PostgreSQL connection string. If empty, health-only startup is allowed and domain GraphQL is disabled |
| `TEST_DATABASE_URL` | `postgres://postgres:postgres@192.168.1.2:5432/todo-e2e?sslmode=disable` | PostgreSQL connection string for DB-backed e2e tests. Tests fail when empty or unreachable |
| `SESSION_SECRET` | none | Cookie session secret, needed when auth is enabled |
| `ADMIN_EMAIL` | none | Email address that is automatically treated as admin and approved on Google login |
| `GOOGLE_CLIENT_ID` | none | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | none | Google OAuth client secret |
| `GOOGLE_REDIRECT_URL` | `http://localhost:3000/auth/callback/google` | Google OAuth callback URL |
| `OPENAI_API_KEY` | none | OpenAI API key for AI proposal generation |
| `OPENAI_PROJECT_ID` | none | Optional OpenAI project ID sent as the `OpenAI-Project` header |
| `OPENAI_BASE_URL` | `https://api.openai.com/v1` | OpenAI-compatible API base URL |
| `OPENAI_MODEL` | `gpt-5.4-mini` | Model used for structured magic-text proposals |
| `OPENAI_TIMEOUT` | `30s` | Timeout for OpenAI proposal generation |

## Codegen

```sh
make api
go run github.com/sqlc-dev/sqlc/cmd/sqlc generate
go run github.com/99designs/gqlgen generate
```

`make api` writes:

- `api/schema.graphql`
- `api/openapi.yaml`

## Docker Release

Pushing a tag like `v1.0.0` builds and pushes:

```text
ghcr.io/<owner>/<repo>:v1.0.0
```

The GitHub Actions workflow intentionally skips tests and multi-arch builds to keep CI runtime low.
