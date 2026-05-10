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

-- name: CountDocumentStates :one
SELECT COUNT(*)
FROM document_states;

-- name: CountTransitionsTotal :one
SELECT COUNT(*)
FROM state_transitions;

-- name: CountTransitionsLast24h :one
SELECT COUNT(*)
FROM state_transitions
WHERE created_at >= NOW() - INTERVAL '24 hours';

-- name: CountDocumentsByState :one
SELECT COUNT(*)
FROM document_states
WHERE state = $1;

-- name: CountDocumentsUpdatedLast24h :one
SELECT COUNT(*)
FROM document_states
WHERE updated_at >= NOW() - INTERVAL '24 hours';

-- name: CountDistinctActorsLast24h :one
SELECT COUNT(DISTINCT actor_login)
FROM state_transitions
WHERE created_at >= NOW() - INTERVAL '24 hours';

-- name: GetStateDistribution :many
SELECT state, COUNT(*) AS documents_count
FROM document_states
GROUP BY state
ORDER BY state;

-- name: GetTransitionMatrixLast24h :many
SELECT from_state, to_state, COUNT(*) AS transitions_count
FROM state_transitions
WHERE created_at >= NOW() - INTERVAL '24 hours'
GROUP BY from_state, to_state
ORDER BY transitions_count DESC, from_state, to_state;

-- name: GetActorActivityLast24h :many
SELECT actor_login, COUNT(*) AS transitions_count
FROM state_transitions
WHERE created_at >= NOW() - INTERVAL '24 hours'
GROUP BY actor_login
ORDER BY transitions_count DESC, actor_login;

-- name: GetHourlyTransitionsLast24h :many
SELECT to_char(date_trunc('hour', created_at), 'YYYY-MM-DD HH24:MI:SS') AS hour_bucket, COUNT(*) AS transitions_count
FROM state_transitions
WHERE created_at >= NOW() - INTERVAL '24 hours'
GROUP BY hour_bucket
ORDER BY hour_bucket;

-- name: GetTransitionsSinceID :many
SELECT id, document_id, actor_login, from_state, to_state, revision_note, created_at
FROM state_transitions
WHERE id > $1
ORDER BY id ASC;
