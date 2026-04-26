# Unified DB Schema (docflow)

This file combines all schemas from:
- docker/postgres/02_iam_tables.sql
- docker/postgres/03_catalog_tables.sql
- docker/postgres/04_collaboration_tables.sql
- docker/postgres/05_orchestrator_tables.sql

## ER Diagram (all schemas in one view)

```mermaid
erDiagram
    IAM_DEPARTMENTS {
        int id PK
        string name UK
        int parent_id FK
        timestamptz created_at
    }

    IAM_USERS {
        int id PK
        string login UK
        string email UK
        string password_hash
        string full_name
        int department_id FK
        bool is_head
        string system_role
        timestamptz created_at
    }

    CATALOG_DOCUMENT_TYPES {
        int id PK
        string name UK
        timestamptz created_at
    }

    CATALOG_FOLDERS {
        int id PK
        int department_id
        int parent_id FK
        string name
        bool is_system
        timestamptz created_at
    }

    CATALOG_DOCUMENTS {
        int id PK
        string title
        string description
        int type_id FK
        int folder_id FK
        string file_path
        string original_name
        int department_id
        int created_by
        int assignee_id
        bool is_hidden
        string status
        timestamptz created_at
        timestamptz updated_at
    }

    CATALOG_DOCUMENT_HISTORY {
        int id PK
        int document_id FK
        string actor_login
        string action
        string details
        timestamptz created_at
    }

    COLLAB_MESSAGES {
        int id PK
        int document_id
        string sender_login
        string content
        timestamptz created_at
    }

    COLLAB_AUDIT_LOGS {
        int id PK
        int department_id
        int document_id
        string actor_login
        string action
        string details
        timestamptz created_at
    }

    ORCH_DOCUMENT_STATES {
        int document_id PK
        string state
        int approver_id
        string revision_note
        timestamptz updated_at
    }

    ORCH_STATE_TRANSITIONS {
        int id PK
        int document_id
        string actor_login
        string from_state
        string to_state
        string revision_note
        timestamptz created_at
    }

    IAM_DEPARTMENTS ||--o{ IAM_DEPARTMENTS : parent_id
    IAM_DEPARTMENTS ||--o{ IAM_USERS : department_id

    CATALOG_DOCUMENT_TYPES ||--o{ CATALOG_DOCUMENTS : type_id
    CATALOG_FOLDERS ||--o{ CATALOG_FOLDERS : parent_id
    CATALOG_FOLDERS ||--o{ CATALOG_DOCUMENTS : folder_id
    CATALOG_DOCUMENTS ||--o{ CATALOG_DOCUMENT_HISTORY : document_id

    %% Logical cross-schema links (without SQL FK constraints):
    IAM_DEPARTMENTS ||..o{ CATALOG_FOLDERS : department_id
    IAM_DEPARTMENTS ||..o{ CATALOG_DOCUMENTS : department_id
    IAM_USERS ||..o{ CATALOG_DOCUMENTS : created_by
    IAM_USERS ||..o{ CATALOG_DOCUMENTS : assignee_id

    CATALOG_DOCUMENTS ||..|| ORCH_DOCUMENT_STATES : document_id
    CATALOG_DOCUMENTS ||..o{ ORCH_STATE_TRANSITIONS : document_id
    CATALOG_DOCUMENTS ||..o{ COLLAB_MESSAGES : document_id
    CATALOG_DOCUMENTS ||..o{ COLLAB_AUDIT_LOGS : document_id

    IAM_USERS ||..o{ ORCH_STATE_TRANSITIONS : actor_login
    IAM_USERS ||..o{ COLLAB_MESSAGES : sender_login
    IAM_USERS ||..o{ COLLAB_AUDIT_LOGS : actor_login
```

## Notes

- Solid lines are actual SQL foreign keys inside a schema.
- Dotted lines are logical references between schemas used by services.
- Cross-schema fields are intentionally not enforced by DB-level FK in current architecture.
