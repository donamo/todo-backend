TEST_DATABASE_URL ?= postgres://postgres:postgres@192.168.1.2:5432/todo-e2e?sslmode=disable
GOCACHE ?= $(CURDIR)/.cache/go-build
GOMODCACHE ?= $(CURDIR)/.cache/go-mod

.PHONY: test unit e2e e2e-docker run api sqlc gqlgen

test: unit e2e

unit:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go test ./cmd/... ./db/... ./internal/...

e2e:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -count=1 ./e2e/...

e2e-docker:
	docker compose up -d postgres-test
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" TEST_DATABASE_URL="postgres://todo:todo@localhost:5433/todo_backend_test?sslmode=disable" go test -count=1 ./e2e/...

run:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go run ./cmd/server

api:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go run ./cmd/apigen

sqlc:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go run github.com/sqlc-dev/sqlc/cmd/sqlc generate

gqlgen:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go run github.com/99designs/gqlgen generate
