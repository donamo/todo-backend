# Style and conventions

- Follow `AGENTS.md` first.
- Keep REST limited to `/health` and auth infrastructure endpoints.
- Put todo/project domain behavior behind GraphQL resolvers.
- Prefer simple Go code; do not introduce interfaces or service layers until there is a concrete need.
- Use `slog`; default log level is DEBUG. Errors handled at boundaries must be logged at ERROR.
- SQL belongs in `db/queries`; generated sqlc files under `internal/db` are not manually edited.
- GraphQL schema lives in `internal/graph/schema.graphqls`; run gqlgen after schema changes.
- Run sqlc after query or schema shape changes.
- User-facing config/API/startup changes require README and `.env.example` updates.
- Long-running server/worker code must be stoppable through context/signal handling.