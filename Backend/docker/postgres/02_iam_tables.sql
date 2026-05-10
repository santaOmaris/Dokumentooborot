
SET search_path TO iam_schema;

CREATE TABLE IF NOT EXISTS departments (
    id        SERIAL PRIMARY KEY,
    name      VARCHAR(255) UNIQUE NOT NULL,
    parent_id INT REFERENCES departments(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_departments_parent_id ON departments(parent_id);

CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    login         VARCHAR(255) UNIQUE NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name     VARCHAR(255) NOT NULL,
    department_id INT REFERENCES departments(id) ON DELETE SET NULL,
    is_head       BOOLEAN NOT NULL DEFAULT FALSE,
    system_role   VARCHAR(50) NOT NULL DEFAULT 'USER',
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email         ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_department_id ON users(department_id);
CREATE INDEX IF NOT EXISTS idx_users_login         ON users(login);
