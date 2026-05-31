-- name: DashboardCounts :one
SELECT
    COUNT(*) FILTER (WHERE status IN ('OPEN', 'IN_PROGRESS', 'BLOCKED'))::int AS open_todos,
    COUNT(*) FILTER (WHERE due_date < CURRENT_DATE AND status IN ('OPEN', 'IN_PROGRESS', 'BLOCKED'))::int AS overdue_todos,
    COUNT(*) FILTER (WHERE next_action = TRUE)::int AS next_actions,
    (SELECT COUNT(*)::int FROM projects WHERE status = 'ACTIVE' AND projects.user_id = $1) AS active_projects,
    COUNT(*) FILTER (WHERE due_date >= CURRENT_DATE AND due_date <= CURRENT_DATE + INTERVAL '7 days' AND status IN ('OPEN', 'IN_PROGRESS', 'BLOCKED'))::int AS upcoming_deadlines
FROM todos
WHERE user_id = $1;
