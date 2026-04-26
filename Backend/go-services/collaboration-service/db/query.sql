SET search_path TO collaboration_schema;

-- name: CreateMessage :one
INSERT INTO messages (document_id, sender_login, content)
VALUES ($1, $2, $3)
RETURNING id, document_id, sender_login, content, created_at;

-- name: ListMessagesByDocument :many
SELECT id, document_id, sender_login, content, created_at
FROM messages
WHERE document_id = $1
ORDER BY created_at ASC;

-- name: InsertAuditLog :exec
INSERT INTO audit_logs (department_id, document_id, actor_login, action, details)
VALUES ($1, $2, $3, $4, $5);

-- name: ListAuditByDepartment :many
SELECT id, department_id, document_id, actor_login, action, details, created_at
FROM audit_logs
WHERE department_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAuditByDocument :many
SELECT id, department_id, document_id, actor_login, action, details, created_at
FROM audit_logs
WHERE document_id = $1
ORDER BY created_at DESC;
