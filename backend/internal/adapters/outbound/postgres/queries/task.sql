-- name: ListTasks :many
SELECT id, title, detail, facility_id, priority, status, due_date,
       assigned_to, created_by, source, created_at
FROM tasks
ORDER BY created_at, id;

-- name: GetTask :one
SELECT id, title, detail, facility_id, priority, status, due_date,
       assigned_to, created_by, source, created_at
FROM tasks
WHERE id = $1;

-- name: UpsertTask :exec
INSERT INTO tasks (
    id, title, detail, facility_id, priority, status, due_date,
    assigned_to, created_by, source, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    detail = EXCLUDED.detail,
    facility_id = EXCLUDED.facility_id,
    priority = EXCLUDED.priority,
    status = EXCLUDED.status,
    due_date = EXCLUDED.due_date,
    assigned_to = EXCLUDED.assigned_to,
    created_by = EXCLUDED.created_by,
    source = EXCLUDED.source,
    created_at = EXCLUDED.created_at;
