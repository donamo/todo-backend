# Suggested commands

Use workspace-local Go caches inside Codex sandbox:

```sh
env GOCACHE=/Users/donamo/project/todo-backend/.cache/go-build GOMODCACHE=/Users/donamo/project/todo-backend/.cache/go-mod go test ./...
```

Normal local development commands:

```sh
go test ./...
go run ./cmd/server
go run github.com/sqlc-dev/sqlc/cmd/sqlc generate
go run github.com/99designs/gqlgen generate
make test
make run
make sqlc
make gqlgen
```

Docker PostgreSQL:

```sh
docker compose up -d postgres
```