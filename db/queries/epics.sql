-- name: ListEpics :many
SELECT id, user_id, name, description, color, position, created_at, updated_at
FROM epics
WHERE user_id = $1
ORDER BY position, created_at;

-- name: GetEpic :one
SELECT id, user_id, name, description, color, position, created_at, updated_at
FROM epics
WHERE id = $1 AND user_id = $2;

-- name: CreateEpic :one
INSERT INTO epics (user_id, name, description, color, position)
VALUES (sqlc.arg('user_id'), sqlc.arg('name'), sqlc.arg('description'), sqlc.arg('color'), COALESCE(sqlc.arg('position'), 0))
RETURNING id, user_id, name, description, color, position, created_at, updated_at;

-- name: UpdateEpic :one
UPDATE epics
SET
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    color = COALESCE(sqlc.narg('color'), color),
    position = COALESCE(sqlc.narg('position'), position),
    updated_at = now()
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id')
RETURNING id, user_id, name, description, color, position, created_at, updated_at;

-- name: DeleteEpic :exec
DELETE FROM epics
WHERE id = $1 AND user_id = $2;

-- name: DetachProjectsFromEpic :exec
UPDATE projects
SET epic_id = NULL, updated_at = now()
WHERE epic_id = $1 AND user_id = $2;

-- name: DeleteProjectsByEpic :exec
DELETE FROM projects
WHERE epic_id = $1 AND user_id = $2;
