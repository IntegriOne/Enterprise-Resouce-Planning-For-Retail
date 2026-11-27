-- Inisialisasi role dan database untuk service Enterprise Foundation.
-- Membaca password service dari variabel lingkungan SERVICES_DB_PASSWORD tanpa menyimpan kredensial di repositori.

\getenv services_db_password SERVICES_DB_PASSWORD
\o /dev/null
\if :{?services_db_password}
SELECT set_config('app.ef_role_password', :'services_db_password', false);
\else
SELECT set_config('app.ef_role_password', '', false);
\endif
\o

DO $$
DECLARE
    role_pw text := NULLIF(current_setting('app.ef_role_password', true), '');
BEGIN
    IF role_pw IS NULL THEN
        RAISE EXCEPTION 'SERVICES_DB_PASSWORD environment variable is required to initialize enterprise_foundation_svc';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'enterprise_foundation_svc') THEN
        EXECUTE format(
            'CREATE ROLE enterprise_foundation_svc WITH LOGIN NOSUPERUSER CREATEDB NOCREATEROLE NOINHERIT NOREPLICATION PASSWORD %L',
            role_pw
        );
    ELSE
        EXECUTE format(
            'ALTER ROLE enterprise_foundation_svc WITH LOGIN NOSUPERUSER CREATEDB NOCREATEROLE NOINHERIT NOREPLICATION PASSWORD %L',
            role_pw
        );
    END IF;
END
$$;

SELECT format('CREATE DATABASE %I OWNER enterprise_foundation_svc ENCODING %L', db_name, 'UTF8')
FROM (VALUES ('enterprise_foundation'), ('enterprise_foundation_test')) AS d(db_name)
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = d.db_name)
\gexec

\connect enterprise_foundation
ALTER DATABASE enterprise_foundation OWNER TO enterprise_foundation_svc;
ALTER SCHEMA public OWNER TO enterprise_foundation_svc;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

\connect enterprise_foundation_test
ALTER DATABASE enterprise_foundation_test OWNER TO enterprise_foundation_svc;
ALTER SCHEMA public OWNER TO enterprise_foundation_svc;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
