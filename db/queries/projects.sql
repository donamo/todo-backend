-- name: ListProjects :many
SELECT id, user_id, epic_id, name, description, status, start_date, target_date, position, created_at, updated_at
FROM projects
WHERE user_id = sqlc.arg('user_id')
  AND (sqlc.narg('epic_id')::uuid IS NULL OR epic_id = sqlc.narg('epic_id')::uuid)
ORDER BY position, created_at;

-- name: GetProject :one
SELECT id, user_id, epic_id, name, description, status, start_date, target_date, position, created_at, updated_at
FROM projects
WHERE id = $1 AND user_id = $2;

-- name: CreateProject :one
INSERT INTO projects (user_id, epic_id, name, description, status, start_date, target_date, position)
VALUES (sqlc.arg('user_id'), sqlc.arg('epic_id'), sqlc.arg('name'), sqlc.arg('description'), COALESCE(sqlc.arg('status'), 'ACTIVE'), sqlc.arg('start_date'), sqlc.arg('target_date'), COALESCE(sqlc.arg('position'), 0))
RETURNING id, user_id, epic_id, name, description, status, start_date, target_date, position, created_at, updated_at;

-- name: UpdateProject :one
UPDATE projects
SET
    epic_id = COALESCE(sqlc.narg('epic_id'), epic_id),
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    status = COALESCE(sqlc.narg('status'), status),
    start_date = COALESCE(sqlc.narg('start_date'), start_date),
    target_date = COALESCE(sqlc.narg('target_date'), target_date),
    position = COALESCE(sqlc.narg('position'), position),
    updated_at = now()
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id')
RETURNING id, user_id, epic_id, name, description, status, start_date, target_date, position, created_at, updated_at;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1 AND user_id = $2;

-- name: DetachStagesFromProject :exec
UPDATE stages
SET project_id = NULL, updated_at = now()
WHERE project_id = $1 AND user_id = $2;

-- name: DetachTodosFromProject :exec
UPDATE todos
SET project_id = NULL, updated_at = now()
WHERE project_id = $1 AND user_id = $2;

-- name: ProjectProgress :one
SELECT
    COUNT(*)::int AS total,
    COUNT(*) FILTER (WHERE status = 'DONE')::int AS done
FROM todos
WHERE project_id = $1 AND user_id = $2;
