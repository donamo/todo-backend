-- name: ListProjectNotes :many
SELECT id, user_id, project_id, body, created_at, updated_at
FROM project_notes
WHERE project_id = $1 AND user_id = $2
ORDER BY created_at DESC;

-- name: CreateProjectNote :one
INSERT INTO project_notes (user_id, project_id, body)
VALUES ($1, $2, $3)
RETURNING id, user_id, project_id, body, created_at, updated_at;

-- name: UpdateProjectNote :one
UPDATE project_notes
SET body = COALESCE(sqlc.narg('body'), body),
    updated_at = now()
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id')
RETURNING id, user_id, project_id, body, created_at, updated_at;

-- name: DeleteProjectNote :exec
DELETE FROM project_notes
WHERE id = $1 AND user_id = $2;
