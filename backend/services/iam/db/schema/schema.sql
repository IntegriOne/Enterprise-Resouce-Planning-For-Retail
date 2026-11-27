-- ============================================================================
-- INTEGRIONE — IDENTITY & ACCESS MANAGEMENT (IAM)
-- PostgreSQL 18+ Baseline Schema (v000001)
-- ============================================================================
-- Berlaku identik untuk database development (`iam`) dan testing (`iam_test`).
-- Diselaraskan 100% dengan: backend/services/iam/db/schema/identity_access_management.dbml
--
-- Catatan pembuatan database fisik (dijalankan otomatis oleh db.EnsureDatabase):
--   CREATE DATABASE iam ENCODING 'UTF8';
--   CREATE DATABASE iam_test ENCODING 'UTF8';
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ============================================================================
-- 1. ENUM TYPES
-- ============================================================================

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_kind') THEN
        CREATE TYPE user_kind AS ENUM (
            'customer',
            'employee',
            'vendor',
            'supplier',
            'system'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_status') THEN
        CREATE TYPE user_status AS ENUM (
            'pending',
            'active',
            'suspended',
            'disabled'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_identities_kind') THEN
        CREATE TYPE user_identities_kind AS ENUM (
            'email',
            'username'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'token_purpose') THEN
        CREATE TYPE token_purpose AS ENUM (
            'email_verification',
            'password_reset',
            'account_invitation',
            'mfa_challenge'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'token_status') THEN
        CREATE TYPE token_status AS ENUM (
            'active',
            'consumed',
            'expired',
            'revoked'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'session_end_reason') THEN
        CREATE TYPE session_end_reason AS ENUM (
            'logout',
            'idle_timeout',
            'absolute_timeout',
            'password_changed',
            'admin_revoked',
            'user_disabled'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'scope_type') THEN
        CREATE TYPE scope_type AS ENUM (
            'global',
            'legal_entity',
            'organization_unit',
            'facility',
            'supplier_party',
            'vendor_party',
            'self'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'outbox_status') THEN
        CREATE TYPE outbox_status AS ENUM (
            'pending',
            'effective',
            'failed'
        );
    END IF;
END$$;

-- ============================================================================
-- 2. UTILITY TRIGGER FUNCTION (updated_at)
-- ============================================================================

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 3. USERS, IDENTITIES & CREDENTIALS
-- ============================================================================

CREATE TABLE IF NOT EXISTS users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_kind     user_kind NOT NULL DEFAULT 'customer',
    status        user_status NOT NULL DEFAULT 'pending',
    display_name  varchar(200) NOT NULL,
    avatar_url    varchar(1000),
    locale        varchar(20) NOT NULL DEFAULT 'id-ID',
    timezone      varchar(80) NOT NULL DEFAULT 'Asia/Jakarta',
    auth_epoch    bigint NOT NULL DEFAULT 1,
    perm_revision bigint NOT NULL DEFAULT 1,
    last_login_at timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz,
    CONSTRAINT ck_users_display_name_nonempty CHECK (btrim(display_name) <> ''),
    CONSTRAINT ck_users_auth_epoch CHECK (auth_epoch >= 1),
    CONSTRAINT ck_users_perm_revision CHECK (perm_revision >= 1)
);

CREATE INDEX IF NOT EXISTS ix_users_kind_status ON users (user_kind, status);
CREATE INDEX IF NOT EXISTS ix_users_deleted_at ON users (deleted_at);

CREATE TABLE IF NOT EXISTS user_identities (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind        user_identities_kind NOT NULL,
    value       varchar(320) NOT NULL,
    is_primary  boolean NOT NULL DEFAULT true,
    is_verified boolean NOT NULL DEFAULT false,
    verified_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_user_identities_kind_value UNIQUE (kind, value),
    CONSTRAINT ck_user_identities_value_nonempty CHECK (btrim(value) <> ''),
    CONSTRAINT ck_user_identities_verified_at CHECK (is_verified = false OR verified_at IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS ix_user_identities_user_kind ON user_identities (user_id, kind);
CREATE UNIQUE INDEX IF NOT EXISTS uq_user_identities_one_pending_email_per_user
    ON user_identities (user_id)
    WHERE (kind = 'email' AND is_verified = false);
CREATE UNIQUE INDEX IF NOT EXISTS uq_user_identities_one_verified_email_per_user
    ON user_identities (user_id)
    WHERE (kind = 'email' AND is_verified = true);
CREATE UNIQUE INDEX IF NOT EXISTS uq_user_identities_one_username_per_user
    ON user_identities (user_id)
    WHERE (kind = 'username');

CREATE TABLE IF NOT EXISTS user_credentials (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    password_hash         text NOT NULL,
    password_changed_at   timestamptz NOT NULL DEFAULT now(),
    must_change_password  boolean NOT NULL DEFAULT false,
    failed_login_attempts integer NOT NULL DEFAULT 0,
    locked_until          timestamptz,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ck_user_credentials_failed_attempts CHECK (failed_login_attempts >= 0)
);

CREATE TABLE IF NOT EXISTS user_totp_factors (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    encrypted_secret bytea NOT NULL,
    key_version      smallint NOT NULL DEFAULT 1,
    is_enabled       boolean NOT NULL DEFAULT false,
    confirmed_at     timestamptz,
    last_used_step   bigint,
    last_used_at     timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ck_user_totp_key_version CHECK (key_version >= 1),
    CONSTRAINT ck_user_totp_confirmed_at CHECK (is_enabled = false OR confirmed_at IS NOT NULL)
);

CREATE TABLE IF NOT EXISTS user_recovery_codes (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash  text NOT NULL UNIQUE,
    used_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_user_recovery_codes_user_unused ON user_recovery_codes (user_id, used_at);

CREATE TABLE IF NOT EXISTS user_tokens (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    identity_id  uuid REFERENCES user_identities(id) ON DELETE SET NULL,
    purpose      token_purpose NOT NULL,
    status       token_status NOT NULL DEFAULT 'active',
    token_digest bytea NOT NULL UNIQUE,
    expires_at   timestamptz NOT NULL,
    consumed_at  timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ck_user_tokens_expiry CHECK (expires_at > created_at),
    CONSTRAINT ck_user_tokens_consumed_status CHECK (
        (status = 'consumed' AND consumed_at IS NOT NULL)
        OR (status <> 'consumed' AND consumed_at IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS ix_user_tokens_user_purpose ON user_tokens (user_id, purpose, expires_at);
CREATE INDEX IF NOT EXISTS ix_user_tokens_identity_id ON user_tokens (identity_id) WHERE identity_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_user_tokens_one_active_per_purpose
    ON user_tokens (user_id, purpose)
    WHERE (status = 'active');

-- ============================================================================
-- 4. SESSIONS
-- ============================================================================

CREATE TABLE IF NOT EXISTS sessions (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_digest        bytea NOT NULL UNIQUE,
    csrf_token_digest   bytea,
    generation          bigint NOT NULL DEFAULT 1,
    csrf_key_version    integer NOT NULL DEFAULT 1,
    ip_address          inet,
    user_agent          text,
    authenticated_at    timestamptz NOT NULL DEFAULT now(),
    last_seen_at        timestamptz NOT NULL DEFAULT now(),
    idle_expires_at     timestamptz NOT NULL,
    absolute_expires_at timestamptz NOT NULL,
    mfa_verified        boolean NOT NULL DEFAULT false,
    mfa_verified_at     timestamptz,
    ended_at            timestamptz,
    end_reason          session_end_reason,
    created_at          timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ck_sessions_generation CHECK (generation >= 1),
    CONSTRAINT ck_sessions_csrf_key_version CHECK (csrf_key_version >= 1),
    CONSTRAINT ck_sessions_expiry_order CHECK (
        idle_expires_at <= absolute_expires_at
        AND absolute_expires_at > authenticated_at
    ),
    CONSTRAINT ck_sessions_mfa_state CHECK (
        (mfa_verified = false AND mfa_verified_at IS NULL)
        OR (mfa_verified = true AND mfa_verified_at IS NOT NULL)
    ),
    CONSTRAINT ck_sessions_end_state CHECK (
        (ended_at IS NULL AND end_reason IS NULL)
        OR (ended_at IS NOT NULL AND end_reason IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS ix_sessions_user_active ON sessions (user_id, ended_at);
CREATE INDEX IF NOT EXISTS ix_sessions_expiry ON sessions (absolute_expires_at);

-- ============================================================================
-- 5. ROLES & PERMISSIONS
-- ============================================================================

CREATE TABLE IF NOT EXISTS roles (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                  varchar(100) NOT NULL UNIQUE,
    name                  varchar(160) NOT NULL,
    description           text,
    management_domain     varchar(100) NOT NULL DEFAULT 'iam',
    default_redirect_path varchar(120) NOT NULL DEFAULT '/',
    redirect_priority     smallint NOT NULL DEFAULT 100,
    is_system             boolean NOT NULL DEFAULT false,
    is_active             boolean NOT NULL DEFAULT true,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ck_roles_redirect_priority CHECK (redirect_priority >= 0)
);

CREATE INDEX IF NOT EXISTS ix_roles_domain_active ON roles (management_domain, is_active);

CREATE TABLE IF NOT EXISTS permissions (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code         varchar(160) NOT NULL UNIQUE,
    resource     varchar(100) NOT NULL,
    action       varchar(60) NOT NULL,
    description  text,
    is_sensitive boolean NOT NULL DEFAULT false,
    created_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_permissions_resource_action UNIQUE (resource, action)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       uuid NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id uuid NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT pk_role_permissions PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS ix_role_permissions_permission ON role_permissions (permission_id);

-- ============================================================================
-- 6. GROUPS & ROLE ASSIGNMENTS
-- ============================================================================

CREATE TABLE IF NOT EXISTS groups (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code        varchar(100) NOT NULL UNIQUE,
    name        varchar(160) NOT NULL,
    description text,
    is_active   boolean NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS group_members (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id   uuid NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    added_by   uuid REFERENCES users(id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_group_members_group_user UNIQUE (group_id, user_id)
);

CREATE INDEX IF NOT EXISTS ix_group_members_user ON group_members (user_id);

CREATE TABLE IF NOT EXISTS user_roles (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id     uuid NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    scope_type  scope_type NOT NULL DEFAULT 'global',
    scope_id    uuid,
    assigned_by uuid REFERENCES users(id) ON DELETE SET NULL,
    expires_at  timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_user_roles_assignment UNIQUE NULLS NOT DISTINCT (user_id, role_id, scope_type, scope_id),
    CONSTRAINT ck_user_roles_scope CHECK (
        (scope_type = 'global' AND scope_id IS NULL)
        OR (scope_type <> 'global' AND scope_id IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS ix_user_roles_user ON user_roles (user_id);

CREATE TABLE IF NOT EXISTS group_roles (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    uuid NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    role_id     uuid NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    scope_type  scope_type NOT NULL DEFAULT 'global',
    scope_id    uuid,
    assigned_by uuid REFERENCES users(id) ON DELETE SET NULL,
    expires_at  timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_group_roles_assignment UNIQUE NULLS NOT DISTINCT (group_id, role_id, scope_type, scope_id),
    CONSTRAINT ck_group_roles_scope CHECK (
        (scope_type = 'global' AND scope_id IS NULL)
        OR (scope_type <> 'global' AND scope_id IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS ix_group_roles_group ON group_roles (group_id);

-- ============================================================================
-- 7. SECURITY CONTROL & OUTBOX RUNTIME REDIS
-- ============================================================================

CREATE TABLE IF NOT EXISTS iam_security_control (
    id             smallint PRIMARY KEY DEFAULT 1,
    recovery_epoch bigint NOT NULL DEFAULT 1,
    updated_at     timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ck_iam_security_control_singleton CHECK (id = 1),
    CONSTRAINT ck_iam_security_control_recovery_epoch CHECK (recovery_epoch >= 1)
);

INSERT INTO iam_security_control (id, recovery_epoch)
VALUES (1, 1)
ON CONFLICT (id) DO NOTHING;

-- BEGIN IAM employee provisioning
-- Antrean provisioning menyimpan penerimaan permintaan HRM dan jejak pemrosesan IAM.
CREATE TABLE IF NOT EXISTS employee_provisioning_requests (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    hrm_request_id uuid NOT NULL UNIQUE,
    worker_id uuid NOT NULL,
    worker_number varchar(100) NOT NULL CHECK (btrim(worker_number) <> ''),
    legal_entity_id uuid NOT NULL,
    catalog_role_id uuid NOT NULL,
    requested_by_principal_id uuid NOT NULL,
    requested_at timestamptz NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now(),
    status varchar(20) NOT NULL DEFAULT 'pending',
    processed_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    claimed_at timestamptz,
    account_user_id uuid UNIQUE REFERENCES users(id) ON DELETE RESTRICT,
    resolved_at timestamptz,
    CONSTRAINT ck_employee_provisioning_state CHECK (
        (status = 'pending' AND processed_by_user_id IS NULL AND claimed_at IS NULL AND account_user_id IS NULL AND resolved_at IS NULL)
        OR (status = 'processing' AND processed_by_user_id IS NOT NULL AND claimed_at IS NOT NULL AND resolved_at IS NULL)
        OR (status = 'linked' AND processed_by_user_id IS NOT NULL AND claimed_at IS NOT NULL AND account_user_id IS NOT NULL AND resolved_at IS NOT NULL)
        OR (status = 'rejected' AND processed_by_user_id IS NOT NULL AND claimed_at IS NOT NULL AND account_user_id IS NULL AND resolved_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_employee_provisioning_open_worker
    ON employee_provisioning_requests (worker_id)
    WHERE status IN ('pending', 'processing');

CREATE INDEX IF NOT EXISTS ix_employee_provisioning_status_received
    ON employee_provisioning_requests (status, received_at, id);
-- END IAM employee provisioning

CREATE TABLE IF NOT EXISTS iam_outbox_events (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type     varchar(80) NOT NULL,
    aggregate_type varchar(40) NOT NULL,
    aggregate_id   uuid NOT NULL,
    payload        jsonb NOT NULL DEFAULT '{}'::jsonb,
    status         outbox_status NOT NULL DEFAULT 'pending',
    attempt_count  integer NOT NULL DEFAULT 0,
    next_retry_at  timestamptz NOT NULL DEFAULT now(),
    last_error     text,
    processed_at   timestamptz,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ck_iam_outbox_attempt_count CHECK (attempt_count >= 0)
);

CREATE INDEX IF NOT EXISTS ix_iam_outbox_pending ON iam_outbox_events (status, next_retry_at, created_at);
CREATE INDEX IF NOT EXISTS ix_iam_outbox_aggregate ON iam_outbox_events (aggregate_type, aggregate_id, created_at);

-- ============================================================================
-- 8. TRIGGERS FOR updated_at
-- ============================================================================

DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_user_identities_updated_at ON user_identities;
CREATE TRIGGER trg_user_identities_updated_at
    BEFORE UPDATE ON user_identities
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_user_credentials_updated_at ON user_credentials;
CREATE TRIGGER trg_user_credentials_updated_at
    BEFORE UPDATE ON user_credentials
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_user_totp_factors_updated_at ON user_totp_factors;
CREATE TRIGGER trg_user_totp_factors_updated_at
    BEFORE UPDATE ON user_totp_factors
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_roles_updated_at ON roles;
CREATE TRIGGER trg_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_groups_updated_at ON groups;
CREATE TRIGGER trg_groups_updated_at
    BEFORE UPDATE ON groups
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_iam_security_control_updated_at ON iam_security_control;
CREATE TRIGGER trg_iam_security_control_updated_at
    BEFORE UPDATE ON iam_security_control
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_iam_outbox_events_updated_at ON iam_outbox_events;
CREATE TRIGGER trg_iam_outbox_events_updated_at
    BEFORE UPDATE ON iam_outbox_events
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
