SET search_path TO orchestrator_schema;

-- name: GetDocumentState :one
SELECT document_id, state, approver_id, revision_note, updated_at
FROM document_states
WHERE document_id = $1;

-- name: UpsertDocumentState :exec
INSERT INTO document_states (document_id, state, approver_id, revision_note, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (document_id)
DO UPDATE SET state = EXCLUDED.state,
              approver_id = EXCLUDED.approver_id,
              revision_note = EXCLUDED.revision_note,
              updated_at = EXCLUDED.updated_at;

-- name: InsertStateTransition :exec
INSERT INTO state_transitions (document_id, actor_login, from_state, to_state, revision_note)
VALUES ($1, $2, $3, $4, $5);

-- name: GetTransitionHistory :many
SELECT id, document_id, actor_login, from_state, to_state, revision_note, created_at
FROM state_transitions
WHERE document_id = $1
ORDER BY created_at DESC;
