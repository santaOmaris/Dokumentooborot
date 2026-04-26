
SET search_path TO catalog_schema;

CREATE TABLE IF NOT EXISTS document_types (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS folders (
    id            SERIAL PRIMARY KEY,
    department_id INT NOT NULL,
    parent_id     INT REFERENCES folders(id) ON DELETE CASCADE,
    name          VARCHAR(255) NOT NULL,
    is_system     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (department_id, parent_id, name)
);

CREATE INDEX IF NOT EXISTS idx_folders_department_id ON folders(department_id);
CREATE INDEX IF NOT EXISTS idx_folders_parent_id     ON folders(parent_id);

CREATE TABLE IF NOT EXISTS documents (
    id             SERIAL PRIMARY KEY,
    title          VARCHAR(512) NOT NULL,
    description    TEXT,
    type_id        INT REFERENCES document_types(id) ON DELETE SET NULL,
    folder_id      INT NOT NULL REFERENCES folders(id) ON DELETE RESTRICT,
    file_path      VARCHAR(1024) NOT NULL,
    original_name  VARCHAR(512) NOT NULL,
    department_id  INT NOT NULL,
    created_by     INT NOT NULL,
    assignee_id    INT,
    is_hidden      BOOLEAN NOT NULL DEFAULT FALSE,
    status         VARCHAR(32) NOT NULL DEFAULT 'DRAFT',
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS status VARCHAR(32) NOT NULL DEFAULT 'DRAFT';

CREATE INDEX IF NOT EXISTS idx_documents_folder_id     ON documents(folder_id);
CREATE INDEX IF NOT EXISTS idx_documents_department_id ON documents(department_id);
CREATE INDEX IF NOT EXISTS idx_documents_assignee_id   ON documents(assignee_id);
CREATE INDEX IF NOT EXISTS idx_documents_title
    ON documents USING gin(to_tsvector('russian', title));

CREATE TABLE IF NOT EXISTS document_history (
    id          SERIAL PRIMARY KEY,
    document_id INT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    actor_login VARCHAR(255) NOT NULL,
    action      VARCHAR(64)  NOT NULL,
    details     TEXT,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_document_history_document_id ON document_history(document_id);
