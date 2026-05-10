-- ─── Document Types ───────────────────────────────────────────────────────────

-- name: ListDocumentTypes :many
SELECT id, name FROM document_types ORDER BY name;

-- name: CreateDocumentType :one
INSERT INTO document_types (name) VALUES ($1) RETURNING id, name;

-- name: DeleteDocumentType :exec
DELETE FROM document_types WHERE id = $1;

-- ─── Folders ──────────────────────────────────────────────────────────────────

-- name: GetFolder :one
SELECT id, department_id, parent_id, name, is_system FROM folders WHERE id = $1;

-- name: ListFoldersByDepartment :many
SELECT id, parent_id, name, is_system FROM folders
WHERE department_id = $1
ORDER BY name;

-- name: CreateFolder :one
INSERT INTO folders (department_id, parent_id, name, is_system)
VALUES ($1, $2, $3, $4)
RETURNING id, name, is_system;

-- name: DeleteFolder :exec
DELETE FROM folders WHERE id = $1 AND is_system = false;

-- name: RenameFolder :exec
UPDATE folders SET name = $2 WHERE id = $1 AND is_system = false;

-- name: MoveFolder :exec
UPDATE folders SET parent_id = $2 WHERE id = $1 AND is_system = false;

-- ─── Documents ────────────────────────────────────────────────────────────────

-- name: GetDocument :one
SELECT id, title, description, type_id, folder_id, file_path, original_name,
       department_id, created_by, assignee_id, is_hidden, status, created_at, updated_at
FROM documents WHERE id = $1;

-- name: ListDocumentsByFolder :many
SELECT id, title, original_name, type_id, assignee_id, is_hidden, created_at
FROM documents
WHERE folder_id = $1
ORDER BY created_at DESC;

-- UC-11: поиск по названию внутри отдела
-- name: SearchDocumentsByTitle :many
SELECT id, title, original_name, folder_id, created_at
FROM documents
WHERE department_id = $1
  AND is_hidden = false
  AND to_tsvector('russian', title) @@ plainto_tsquery('russian', $2)
ORDER BY created_at DESC;

-- name: CreateDocument :one
INSERT INTO documents (title, description, type_id, folder_id, file_path, original_name,
                       department_id, created_by, assignee_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id;

-- name: MoveDocument :exec
UPDATE documents
SET folder_id = $2,
  department_id = (SELECT folders.department_id FROM folders WHERE folders.id = $2),
    updated_at = NOW()
WHERE documents.id = $1;

-- UC-7 вариант 2: ADMIN полностью скрывает документ
-- name: HideDocument :exec
UPDATE documents SET is_hidden = true, updated_at = NOW() WHERE id = $1;

-- name: UnhideDocument :exec
UPDATE documents SET is_hidden = false, updated_at = NOW() WHERE id = $1;

-- name: ChangeDocumentAssignee :exec
UPDATE documents SET assignee_id = $2, updated_at = NOW() WHERE id = $1;

-- name: UpdateDocumentStatus :exec
UPDATE documents SET status = $2, updated_at = NOW() WHERE id = $1;

-- name: DeleteDocument :exec
DELETE FROM documents WHERE id = $1;

-- ─── Document History ─────────────────────────────────────────────────────────

-- name: AddDocumentHistory :exec
INSERT INTO document_history (document_id, actor_login, action, details)
VALUES ($1, $2, $3, $4);

-- name: GetDocumentHistory :many
SELECT id, actor_login, action, details, created_at
FROM document_history
WHERE document_id = $1
ORDER BY created_at DESC;
