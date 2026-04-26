
CREATE TABLE iam_departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    parent_id INT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_iam_departments_parent
        FOREIGN KEY (parent_id) REFERENCES iam_departments(id) ON DELETE SET NULL
);

CREATE TABLE iam_users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    department_id INT,
    is_head BOOLEAN NOT NULL DEFAULT FALSE,
    system_role VARCHAR(50) NOT NULL DEFAULT 'USER',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_iam_users_department
        FOREIGN KEY (department_id) REFERENCES iam_departments(id) ON DELETE SET NULL
);

CREATE TABLE catalog_document_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE catalog_folders (
    id SERIAL PRIMARY KEY,
    department_id INT NOT NULL,
    parent_id INT,
    name VARCHAR(255) NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_catalog_folders UNIQUE (department_id, parent_id, name),
    CONSTRAINT fk_catalog_folders_parent
        FOREIGN KEY (parent_id) REFERENCES catalog_folders(id) ON DELETE CASCADE,
    CONSTRAINT fk_catalog_folders_department
        FOREIGN KEY (department_id) REFERENCES iam_departments(id)
);

CREATE TABLE catalog_documents (
    id SERIAL PRIMARY KEY,
    title VARCHAR(512) NOT NULL,
    description TEXT,
    type_id INT,
    folder_id INT NOT NULL,
    file_path VARCHAR(1024) NOT NULL,
    original_name VARCHAR(512) NOT NULL,
    department_id INT NOT NULL,
    created_by INT NOT NULL,
    assignee_id INT,
    is_hidden BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_catalog_documents_type
        FOREIGN KEY (type_id) REFERENCES catalog_document_types(id) ON DELETE SET NULL,
    CONSTRAINT fk_catalog_documents_folder
        FOREIGN KEY (folder_id) REFERENCES catalog_folders(id) ON DELETE RESTRICT,
    CONSTRAINT fk_catalog_documents_department
        FOREIGN KEY (department_id) REFERENCES iam_departments(id),
    CONSTRAINT fk_catalog_documents_creator
        FOREIGN KEY (created_by) REFERENCES iam_users(id),
    CONSTRAINT fk_catalog_documents_assignee
        FOREIGN KEY (assignee_id) REFERENCES iam_users(id)
);

CREATE TABLE catalog_document_history (
    id SERIAL PRIMARY KEY,
    document_id INT NOT NULL,
    actor_login VARCHAR(255) NOT NULL,
    action VARCHAR(64) NOT NULL,
    details TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_catalog_document_history_document
        FOREIGN KEY (document_id) REFERENCES catalog_documents(id) ON DELETE CASCADE,
    CONSTRAINT fk_catalog_document_history_actor
        FOREIGN KEY (actor_login) REFERENCES iam_users(login)
);

CREATE TABLE collaboration_messages (
    id SERIAL PRIMARY KEY,
    document_id INT NOT NULL,
    sender_login VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_collaboration_messages_document
        FOREIGN KEY (document_id) REFERENCES catalog_documents(id),
    CONSTRAINT fk_collaboration_messages_sender
        FOREIGN KEY (sender_login) REFERENCES iam_users(login)
);

CREATE TABLE collaboration_audit_logs (
    id SERIAL PRIMARY KEY,
    department_id INT,
    document_id INT,
    actor_login VARCHAR(255),
    action VARCHAR(64) NOT NULL,
    details TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_collaboration_audit_department
        FOREIGN KEY (department_id) REFERENCES iam_departments(id),
    CONSTRAINT fk_collaboration_audit_document
        FOREIGN KEY (document_id) REFERENCES catalog_documents(id),
    CONSTRAINT fk_collaboration_audit_actor
        FOREIGN KEY (actor_login) REFERENCES iam_users(login)
);

CREATE TABLE orchestrator_document_states (
    document_id INT PRIMARY KEY,
    state VARCHAR(32) NOT NULL DEFAULT 'DRAFT',
    approver_id INT,
    revision_note TEXT,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_orchestrator_state_document
        FOREIGN KEY (document_id) REFERENCES catalog_documents(id),
    CONSTRAINT fk_orchestrator_state_approver
        FOREIGN KEY (approver_id) REFERENCES iam_users(id)
);

CREATE TABLE orchestrator_state_transitions (
    id SERIAL PRIMARY KEY,
    document_id INT NOT NULL,
    actor_login VARCHAR(255) NOT NULL,
    from_state VARCHAR(32) NOT NULL,
    to_state VARCHAR(32) NOT NULL,
    revision_note TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_orchestrator_transition_document
        FOREIGN KEY (document_id) REFERENCES catalog_documents(id),
    CONSTRAINT fk_orchestrator_transition_actor
        FOREIGN KEY (actor_login) REFERENCES iam_users(login)
);