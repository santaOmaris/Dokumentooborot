
CREATE SCHEMA IF NOT EXISTS iam_schema;
CREATE SCHEMA IF NOT EXISTS catalog_schema;
CREATE SCHEMA IF NOT EXISTS orchestrator_schema;
CREATE SCHEMA IF NOT EXISTS collaboration_schema;

DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'iam_user') THEN
        CREATE USER iam_user WITH PASSWORD 'iam_pass';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'catalog_user') THEN
        CREATE USER catalog_user WITH PASSWORD 'catalog_pass';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'orchestrator_user') THEN
        CREATE USER orchestrator_user WITH PASSWORD 'orchestrator_pass';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'collaboration_user') THEN
        CREATE USER collaboration_user WITH PASSWORD 'collaboration_pass';
    END IF;
END
$$;

GRANT ALL ON SCHEMA iam_schema TO iam_user;
GRANT ALL ON SCHEMA catalog_schema TO catalog_user;
GRANT ALL ON SCHEMA orchestrator_schema TO orchestrator_user;
GRANT ALL ON SCHEMA collaboration_schema TO collaboration_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA iam_schema GRANT ALL ON TABLES TO iam_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA iam_schema GRANT ALL ON SEQUENCES TO iam_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA catalog_schema GRANT ALL ON TABLES TO catalog_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA catalog_schema GRANT ALL ON SEQUENCES TO catalog_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA orchestrator_schema GRANT ALL ON TABLES TO orchestrator_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA orchestrator_schema GRANT ALL ON SEQUENCES TO orchestrator_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA collaboration_schema GRANT ALL ON TABLES TO collaboration_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA collaboration_schema GRANT ALL ON SEQUENCES TO collaboration_user;

CREATE TABLE IF NOT EXISTS iam_schema.departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    parent_id INT REFERENCES iam_schema.departments(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_departments_parent_id
    ON iam_schema.departments(parent_id);

CREATE TABLE IF NOT EXISTS iam_schema.users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    department_id INT REFERENCES iam_schema.departments(id) ON DELETE SET NULL,
    is_head BOOLEAN NOT NULL DEFAULT FALSE,
    system_role VARCHAR(50) NOT NULL DEFAULT 'USER',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email
    ON iam_schema.users(email);
CREATE INDEX IF NOT EXISTS idx_users_department_id
    ON iam_schema.users(department_id);
CREATE INDEX IF NOT EXISTS idx_users_login
    ON iam_schema.users(login);

CREATE TABLE IF NOT EXISTS catalog_schema.document_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS catalog_schema.folders (
    id SERIAL PRIMARY KEY,
    department_id INT NOT NULL,
    parent_id INT REFERENCES catalog_schema.folders(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (department_id, parent_id, name)
);

CREATE INDEX IF NOT EXISTS idx_folders_department_id
    ON catalog_schema.folders(department_id);
CREATE INDEX IF NOT EXISTS idx_folders_parent_id
    ON catalog_schema.folders(parent_id);

CREATE TABLE IF NOT EXISTS catalog_schema.documents (
    id SERIAL PRIMARY KEY,
    title VARCHAR(512) NOT NULL,
    description TEXT,
    type_id INT REFERENCES catalog_schema.document_types(id) ON DELETE SET NULL,
    folder_id INT NOT NULL REFERENCES catalog_schema.folders(id) ON DELETE RESTRICT,
    file_path VARCHAR(1024) NOT NULL,
    original_name VARCHAR(512) NOT NULL,
    department_id INT NOT NULL,
    created_by INT NOT NULL,
    assignee_id INT,
    is_hidden BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_documents_folder_id
    ON catalog_schema.documents(folder_id);
CREATE INDEX IF NOT EXISTS idx_documents_department_id
    ON catalog_schema.documents(department_id);
CREATE INDEX IF NOT EXISTS idx_documents_assignee_id
    ON catalog_schema.documents(assignee_id);
CREATE INDEX IF NOT EXISTS idx_documents_title
    ON catalog_schema.documents USING gin(to_tsvector('russian', title));

CREATE TABLE IF NOT EXISTS catalog_schema.document_history (
    id SERIAL PRIMARY KEY,
    document_id INT NOT NULL REFERENCES catalog_schema.documents(id) ON DELETE CASCADE,
    actor_login VARCHAR(255) NOT NULL,
    action VARCHAR(64) NOT NULL,
    details TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_document_history_document_id
    ON catalog_schema.document_history(document_id);

CREATE TABLE IF NOT EXISTS collaboration_schema.messages (
    id SERIAL PRIMARY KEY,
    document_id INT NOT NULL,
    sender_login VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_document_id
    ON collaboration_schema.messages(document_id);

CREATE TABLE IF NOT EXISTS collaboration_schema.audit_logs (
    id SERIAL PRIMARY KEY,
    department_id INT,
    document_id INT,
    actor_login VARCHAR(255),
    action VARCHAR(64) NOT NULL,
    details TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_department_id
    ON collaboration_schema.audit_logs(department_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_document_id
    ON collaboration_schema.audit_logs(document_id);

CREATE TABLE IF NOT EXISTS orchestrator_schema.document_states (
    document_id INT PRIMARY KEY,
    state VARCHAR(32) NOT NULL DEFAULT 'DRAFT',
    approver_id INT,
    revision_note TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS orchestrator_schema.state_transitions (
    id SERIAL PRIMARY KEY,
    document_id INT NOT NULL,
    actor_login VARCHAR(255) NOT NULL,
    from_state VARCHAR(32) NOT NULL,
    to_state VARCHAR(32) NOT NULL,
    revision_note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_state_transitions_document
    ON orchestrator_schema.state_transitions(document_id);

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA iam_schema TO iam_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA iam_schema TO iam_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA catalog_schema TO catalog_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA catalog_schema TO catalog_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA orchestrator_schema TO orchestrator_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA orchestrator_schema TO orchestrator_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA collaboration_schema TO collaboration_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA collaboration_schema TO collaboration_user;
