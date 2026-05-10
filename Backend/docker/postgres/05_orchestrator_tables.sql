
\c docflow
SET search_path TO orchestrator_schema;

CREATE TABLE IF NOT EXISTS document_states (
    document_id    INT PRIMARY KEY,
    state          VARCHAR(32)  NOT NULL DEFAULT 'DRAFT',
    approver_id    INT,          -- ID визировщика из IAM (назначается при SendForVisa)
    revision_note  TEXT,         -- причина отказа при REJECTED
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS state_transitions (
    id             SERIAL PRIMARY KEY,
    document_id    INT          NOT NULL,
    actor_login    VARCHAR(255) NOT NULL,
    from_state     VARCHAR(32)  NOT NULL,
    to_state       VARCHAR(32)  NOT NULL,
    revision_note  TEXT,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_state_transitions_document ON state_transitions(document_id);
