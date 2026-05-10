
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

GRANT ALL ON SCHEMA iam_schema          TO iam_user;
GRANT ALL ON SCHEMA catalog_schema      TO catalog_user;
GRANT ALL ON SCHEMA orchestrator_schema TO orchestrator_user;
GRANT ALL ON SCHEMA collaboration_schema TO collaboration_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA iam_schema           GRANT ALL ON TABLES    TO iam_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA iam_schema           GRANT ALL ON SEQUENCES TO iam_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA catalog_schema       GRANT ALL ON TABLES    TO catalog_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA catalog_schema       GRANT ALL ON SEQUENCES TO catalog_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA orchestrator_schema  GRANT ALL ON TABLES    TO orchestrator_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA orchestrator_schema  GRANT ALL ON SEQUENCES TO orchestrator_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA collaboration_schema GRANT ALL ON TABLES    TO collaboration_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA collaboration_schema GRANT ALL ON SEQUENCES TO collaboration_user;
