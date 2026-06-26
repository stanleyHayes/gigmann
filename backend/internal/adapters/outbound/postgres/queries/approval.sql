-- name: ListApprovals :many
SELECT id, type, facility_id, amount, title, context, requested_by,
       status, decided_at, decision_note, created_at
FROM approvals
ORDER BY created_at, id;

-- name: GetApproval :one
SELECT id, type, facility_id, amount, title, context, requested_by,
       status, decided_at, decision_note, created_at
FROM approvals
WHERE id = $1;

-- name: UpsertApproval :exec
INSERT INTO approvals (
    id, type, facility_id, amount, title, context, requested_by,
    status, decided_at, decision_note, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (id) DO UPDATE SET
    type = EXCLUDED.type,
    facility_id = EXCLUDED.facility_id,
    amount = EXCLUDED.amount,
    title = EXCLUDED.title,
    context = EXCLUDED.context,
    requested_by = EXCLUDED.requested_by,
    status = EXCLUDED.status,
    decided_at = EXCLUDED.decided_at,
    decision_note = EXCLUDED.decision_note,
    created_at = EXCLUDED.created_at;
