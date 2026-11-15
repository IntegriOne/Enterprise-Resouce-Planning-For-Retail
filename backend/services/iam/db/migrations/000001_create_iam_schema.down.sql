-- ============================================================================
-- Migrasi Baseline v000001 (DOWN) — IAM IntegriOne
-- Menghapus seluruh objek skema IAM secara terbalik dari dependensinya.
-- ============================================================================

DROP TRIGGER IF EXISTS trg_iam_outbox_events_updated_at ON iam_outbox_events;
DROP TRIGGER IF EXISTS trg_iam_security_control_updated_at ON iam_security_control;
DROP TRIGGER IF EXISTS trg_groups_updated_at ON groups;
DROP TRIGGER IF EXISTS trg_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS trg_user_totp_factors_updated_at ON user_totp_factors;
DROP TRIGGER IF EXISTS trg_user_credentials_updated_at ON user_credentials;
DROP TRIGGER IF EXISTS trg_user_identities_updated_at ON user_identities;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;

DROP TABLE IF EXISTS iam_outbox_events CASCADE;
DROP TABLE IF EXISTS iam_security_control CASCADE;
DROP TABLE IF EXISTS group_roles CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS group_members CASCADE;
DROP TABLE IF EXISTS groups CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS user_tokens CASCADE;
DROP TABLE IF EXISTS user_recovery_codes CASCADE;
DROP TABLE IF EXISTS user_totp_factors CASCADE;
DROP TABLE IF EXISTS user_credentials CASCADE;
DROP TABLE IF EXISTS user_identities CASCADE;
DROP TABLE IF EXISTS users CASCADE;

DROP FUNCTION IF EXISTS set_updated_at() CASCADE;

DROP TYPE IF EXISTS outbox_status CASCADE;
DROP TYPE IF EXISTS scope_type CASCADE;
DROP TYPE IF EXISTS session_end_reason CASCADE;
DROP TYPE IF EXISTS token_status CASCADE;
DROP TYPE IF EXISTS token_purpose CASCADE;
DROP TYPE IF EXISTS user_identities_kind CASCADE;
DROP TYPE IF EXISTS user_status CASCADE;
DROP TYPE IF EXISTS user_kind CASCADE;
