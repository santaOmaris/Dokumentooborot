
\c docflow
SET search_path TO collaboration_schema;

CREATE TABLE IF NOT EXISTS messages (
    id           SERIAL PRIMARY KEY,
    document_id  INT          NOT NULL,
    sender_login VARCHAR(255) NOT NULL,
    content      TEXT         NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_document_id ON messages(document_id);

CREATE TABLE IF NOT EXISTS audit_logs (
    id            SERIAL PRIMARY KEY,
    department_id INT,
    document_id   INT,
    actor_login   VARCHAR(255),
    action        VARCHAR(64)  NOT NULL,
    details       TEXT,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_department_id ON audit_logs(department_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_document_id   ON audit_logs(document_id);
