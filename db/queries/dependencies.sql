-- name: ListTodoDependencies :many
SELECT d.todo_id, d.depends_on_todo_id, d.created_at
FROM todo_dependencies d
JOIN todos t ON t.id = d.todo_id
WHERE d.todo_id = $1 AND t.user_id = $2
ORDER BY d.created_at;

-- name: AddTodoDependency :exec
INSERT INTO todo_dependencies (todo_id, depends_on_todo_id)
SELECT $1, $2
WHERE EXISTS (SELECT 1 FROM todos WHERE todos.id = $1 AND todos.user_id = $3)
  AND EXISTS (SELECT 1 FROM todos WHERE todos.id = $2 AND todos.user_id = $3)
ON CONFLICT DO NOTHING;

-- name: RemoveTodoDependency :exec
DELETE FROM todo_dependencies
WHERE todo_id = $1 AND depends_on_todo_id = $2
  AND EXISTS (SELECT 1 FROM todos WHERE id = $1 AND user_id = $3);

-- name: CountIncompleteDependencies :one
SELECT COUNT(*)::int
FROM todo_dependencies d
JOIN todos dep ON dep.id = d.depends_on_todo_id
JOIN todos t ON t.id = d.todo_id
WHERE d.todo_id = $1 AND t.user_id = $2 AND dep.status <> 'DONE';
