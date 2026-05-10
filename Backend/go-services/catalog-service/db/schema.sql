CREATE SCHEMA IF NOT EXISTS catalog_schema;
SET search_path TO catalog_schema;

-- Типы документов (UC-19: настраиваются ADMIN-ом)
CREATE TABLE document_types (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Папки отдела (UC-6, UC-11)
-- is_system = true: папка неудаляемая (main, templates, archived, head_only, collaborations)
CREATE TABLE folders (
    id            SERIAL PRIMARY KEY,
    department_id INT NOT NULL,
    parent_id     INT REFERENCES folders(id) ON DELETE CASCADE,
    name          VARCHAR(255) NOT NULL,
    is_system     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (department_id, parent_id, name)
);

CREATE INDEX idx_folders_department_id ON folders(department_id);
CREATE INDEX idx_folders_parent_id     ON folders(parent_id);

-- Документы (UC-4, UC-6, UC-7, UC-8, UC-10)
CREATE TABLE documents (
    id             SERIAL PRIMARY KEY,
    title          VARCHAR(512) NOT NULL,
    description    TEXT,

    -- Тип документа (UC-19)
    type_id        INT REFERENCES document_types(id) ON DELETE SET NULL,

    -- Физическое расположение (папка, MinIO)
    folder_id      INT NOT NULL REFERENCES folders(id) ON DELETE RESTRICT,
    file_path      VARCHAR(1024) NOT NULL, -- object_path в MinIO
    original_name  VARCHAR(512) NOT NULL,  -- оригинальное имя файла

    -- Мета
    department_id  INT NOT NULL,           -- отдел-владелец
    created_by     INT NOT NULL,           -- login/id из IAM
    assignee_id    INT,                    -- текущий ответственный

    -- UC-7: скрытие документа
    is_hidden      BOOLEAN NOT NULL DEFAULT FALSE, -- скрыт ADMIN-ом (global)

    -- UC-9/13: статус FSM (управляется оркестратором)
    status         VARCHAR(32) NOT NULL DEFAULT 'DRAFT',

    created_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_documents_folder_id     ON documents(folder_id);
CREATE INDEX idx_documents_department_id ON documents(department_id);
CREATE INDEX idx_documents_assignee_id   ON documents(assignee_id);
CREATE INDEX idx_documents_title         ON documents USING gin(to_tsvector('russian', title));

-- История документа: перемещения, переименования папок (UC-11, аудит)
CREATE TABLE document_history (
    id          SERIAL PRIMARY KEY,
    document_id INT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    actor_login VARCHAR(255) NOT NULL,
    action      VARCHAR(64)  NOT NULL, -- MOVED, FOLDER_CREATED, etc.
    details     TEXT,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_document_history_document_id ON document_history(document_id);
