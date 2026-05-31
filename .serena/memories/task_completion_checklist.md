# Task completion checklist

Before finishing backend changes:

```sh
make test
```

`make test` runs unit packages and then `make e2e`. `make e2e` starts the local `postgres-test` Docker Compose service and runs DB-backed e2e tests with:

```sh
TEST_DATABASE_URL=postgres://todo:todo@localhost:5433/todo_backend_test?sslmode=disable
```

DB-backed e2e tests intentionally fail when `TEST_DATABASE_URL` is empty or unreachable; they must not skip missing infrastructure.

Inside Codex sandbox, prefer workspace-local caches when running raw Go commands:

```sh
env GOCACHE=/Users/donamo/project/todo-backend/.cache/go-build GOMODCACHE=/Users/donamo/project/todo-backend/.cache/go-mod go test ./cmd/... ./db/... ./internal/...
```

If SQL migrations or `db/queries` changed:

```sh
go run github.com/sqlc-dev/sqlc/cmd/sqlc generate
```

If GraphQL schema changed:

```sh
go run github.com/99designs/gqlgen generate
```