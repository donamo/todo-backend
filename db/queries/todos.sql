-- name: ListTodos :many
SELECT id, user_id, project_id, stage_id, title, description, priority, status, start_date, due_date, estimated_effort, position, next_action, milestone, recurrence, completed_at, created_at, updated_at
FROM todos
WHERE user_id = sqlc.arg('user_id')
  AND (sqlc.narg('project_id')::uuid IS NULL OR project_id = sqlc.narg('project_id')::uuid)
  AND (sqlc.narg('stage_id')::uuid IS NULL OR stage_id = sqlc.narg('stage_id')::uuid)
  AND (sqlc.narg('label_id')::uuid IS NULL OR EXISTS (
      SELECT 1 FROM todo_labels tl WHERE tl.todo_id = todos.id AND tl.label_id = sqlc.narg('label_id')::uuid
  ))
  AND (sqlc.narg('priority')::text IS NULL OR priority = sqlc.narg('priority')::text)
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')::text)
ORDER BY position, created_at;

-- name: ListInboxTodos :many
SELECT id, user_id, project_id, stage_id, title, description, priority, status, start_date, due_date, estimated_effort, position, next_action, milestone, recurrence, completed_at, created_at, updated_at
FROM todos
WHERE user_id = $1 AND project_id IS NULL
ORDER BY created_at DESC;

-- name: ListOpenTodos :many
SELECT id, user_id, project_id, stage_id, title, description, priority, status, start_date, due_date, estimated_effort, position, next_action, milestone, recurrence, completed_at, created_at, updated_at
FROM todos
WHERE user_id = $1 AND status IN ('OPEN', 'IN_PROGRESS', 'BLOCKED')
ORDER BY
    CASE priority
        WHEN 'CRITICAL' THEN 0
        WHEN 'HIGH' THEN 1
        WHEN 'NORMAL' THEN 2
        ELSE 3
    END,
    due_date NULLS LAST,
    position,
    created_at;

-- name: ListOverdueTodos :many
SELECT id, user_id, project_id, stage_id, title, description, priority, status, start_date, due_date, estimated_effort, position, next_action, milestone, recurrence, completed_at, created_at, updated_at
FROM todos
WHERE user_id = $1
  AND due_date < CURRENT_DATE
  AND status IN ('OPEN', 'IN_PROGRESS', 'BLOCKED')
ORDER BY due_date, position, created_at;

-- name: ListNextActionTodos :many
SELECT id, user_id, project_id, stage_id, title, description, priority, status, start_date, due_date, estimated_effort, position, next_action, milestone, recurrence, completed_at, created_at, updated_at
FROM todos
WHERE user_id = $1 AND next_action = TRUE
ORDER BY due_date NULLS LAST, position, created_at;

-- name: ListDoneTodos :many
SELECT id, user_id, project_id, stage_id, title, description, priority, status, start_date, due_date, estimated_effort, position, next_action, milestone, recurrence, completed_at, created_at, updated_at
FROM todos
WHERE user_id = $1 AND status = 'DONE'
ORDER BY completed_at DESC NULLS LAST, updated_at DESC;

-- name: GetTodo :one
SELECT id, user_id, project_id, stage_id, title, description, priority, status, start_date, due_date, estimated_effort, position, next_action, milestone, recurrence, completed_at, created_at, updated_at
FROM todos
WHERE id = $1 AND user_id = $2;

-- name: CreateTodo :one
INSERT INTO todos (user_id, project_id, stage_id, title, description, priority, status, start_date, due_date, estimated_effort, position, next_action, milestone, recurrence)
VALUES (sqlc.arg('user_id'), sqlc.arg('project_id'), sqlc.arg('stage_id'), sqlc.arg('title'), sqlc.arg('description'), COALESCE(sqlc.arg('priority'), 'NORMAL'), COALESCE(sqlc.arg('status'), 'OPEN'), sqlc.arg('start_date'), sqlc.arg('due_date'), sqlc.arg('estimated_effort'), COALESCE(sqlc.arg('position'), 0), COALESCE(sqlc.arg('next_action'), FALSE), COALESCE(sqlc.arg('milestone'), FALSE), COALESCE(sqlc.arg('recurrence'), 'NONE'))
RETURNING id, user_id, project_id, stage_id, title, description, priority, status, start_date, due_date, estimated_effort, position, next_action, milestone, recurrence, completed_at, created_at, updated_at;

-- name: UpdateTodo :one
UPDATE todos
SET
    project_id = COALESCE(sqlc.narg('project_id'), project_id),
    stage_id = COALESCE(sqlc.narg('stage_id'), stage_id),
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    priority = COALESCE(sqlc.narg('priority'), priority),
    status = COALESCE(sqlc.narg('status'), status),
    start_date = COALESCE(sqlc.narg('start_date'), start_date),
    due_date = COALESCE(sqlc.narg('due_date'), due_date),
    estimated_effort = COALESCE(sqlc.narg('estimated_effort'), estimated_effort),
    position = COALESCE(sqlc.narg('position'), position),
    next_action = COALESCE(sqlc.narg('next_action'), next_action),
    milestone = COALESCE(sqlc.narg('milestone'), milestone),
    recurrence = COALESCE(sqlc.narg('recurrence'), recurrence),
    completed_at = CASE
        WHEN sqlc.narg('status')::text = 'DONE' AND completed_at IS NULL THEN now()
        WHEN sqlc.narg('status')::text IS NOT NULL AND sqlc.narg('status')::text <> 'DONE' THEN NULL
        ELSE completed_at
    END,
    updated_at = now()
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id')
RETURNING id, user_id, project_id, stage_id, title, description, priority, status, start_date, due_date, estimated_effort, position, next_action, milestone, recurrence, completed_at, created_at, updated_at;

-- name: ClearProjectNextActions :exec
UPDATE todos
SET next_action = FALSE, updated_at = now()
WHERE user_id = $1 AND project_id = $2 AND id <> $3;

-- name: DeleteTodo :exec
DELETE FROM todos
WHERE id = $1 AND user_id = $2;
