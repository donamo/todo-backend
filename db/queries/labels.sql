-- name: ListLabels :many
SELECT id, user_id, name, color, created_at, updated_at
FROM labels
WHERE user_id = $1
ORDER BY name;

-- name: ListTodoLabels :many
SELECT l.id, l.user_id, l.name, l.color, l.created_at, l.updated_at
FROM labels l
JOIN todo_labels tl ON tl.label_id = l.id
WHERE tl.todo_id = $1 AND l.user_id = $2
ORDER BY l.name;

-- name: CreateLabel :one
INSERT INTO labels (user_id, name, color)
VALUES ($1, $2, $3)
RETURNING id, user_id, name, color, created_at, updated_at;

-- name: UpdateLabel :one
UPDATE labels
SET
    name = COALESCE(sqlc.narg('name'), name),
    color = COALESCE(sqlc.narg('color'), color),
    updated_at = now()
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id')
RETURNING id, user_id, name, color, created_at, updated_at;

-- name: DeleteLabel :exec
DELETE FROM labels
WHERE id = $1 AND user_id = $2;

-- name: AddTodoLabel :exec
INSERT INTO todo_labels (todo_id, label_id)
SELECT $1, $2
WHERE EXISTS (SELECT 1 FROM todos WHERE todos.id = $1 AND todos.user_id = $3)
  AND EXISTS (SELECT 1 FROM labels WHERE labels.id = $2 AND labels.user_id = $3)
ON CONFLICT DO NOTHING;

-- name: RemoveTodoLabel :exec
DELETE FROM todo_labels
WHERE todo_id = $1 AND label_id = $2
  AND EXISTS (SELECT 1 FROM todos WHERE id = $1 AND user_id = $3);
