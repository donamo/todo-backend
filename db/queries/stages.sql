-- name: ListStages :many
SELECT id, user_id, project_id, name, description, status, start_date, target_date, position, created_at, updated_at
FROM stages
WHERE project_id = $1 AND user_id = $2
ORDER BY position, created_at;

-- name: GetStage :one
SELECT id, user_id, project_id, name, description, status, start_date, target_date, position, created_at, updated_at
FROM stages
WHERE id = $1 AND user_id = $2;

-- name: CreateStage :one
INSERT INTO stages (user_id, project_id, name, description, status, start_date, target_date, position)
VALUES (sqlc.arg('user_id'), sqlc.arg('project_id'), sqlc.arg('name'), sqlc.arg('description'), COALESCE(sqlc.arg('status'), 'PLANNED'), sqlc.arg('start_date'), sqlc.arg('target_date'), COALESCE(sqlc.arg('position'), 0))
RETURNING id, user_id, project_id, name, description, status, start_date, target_date, position, created_at, updated_at;

-- name: UpdateStage :one
UPDATE stages
SET
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    status = COALESCE(sqlc.narg('status'), status),
    start_date = COALESCE(sqlc.narg('start_date'), start_date),
    target_date = COALESCE(sqlc.narg('target_date'), target_date),
    position = COALESCE(sqlc.narg('position'), position),
    updated_at = now()
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id')
RETURNING id, user_id, project_id, name, description, status, start_date, target_date, position, created_at, updated_at;

-- name: DeleteStage :exec
DELETE FROM stages
WHERE id = $1 AND user_id = $2;

-- name: StageProgress :one
SELECT
    COUNT(*)::int AS total,
    COUNT(*) FILTER (WHERE status = 'DONE')::int AS done
FROM todos
WHERE stage_id = $1 AND user_id = $2;
