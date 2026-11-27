-- ============================================================================
-- Migrasi Baseline v000001 (UP) — Enterprise Foundation IntegriOne
-- Diterapkan ke database development `enterprise_foundation` dan database testing
-- `enterprise_foundation_test`.
-- Selaras dengan spesifikasi DBML: db/schema/enterprise_foundation.dbml
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 1. Enum Types
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'legal_entity_status') THEN
        CREATE TYPE legal_entity_status AS ENUM (
            'pending',
            'active',
            'suspended',
            'dissolved'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'org_unit_kind') THEN
        CREATE TYPE org_unit_kind AS ENUM (
            'division',
            'department',
            'team',
            'branch',
            'other'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'location_kind') THEN
        CREATE TYPE location_kind AS ENUM (
            'headquarters',
            'warehouse',
            'retail_store',
            'branch_office'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'party_kind') THEN
        CREATE TYPE party_kind AS ENUM (
            'organization',
            'individual',
            'financial_entity',
            'government_body'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'party_status') THEN
        CREATE TYPE party_status AS ENUM (
            'pending_verification',
            'active',
            'suspended',
            'inactive'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'party_role_kind') THEN
        CREATE TYPE party_role_kind AS ENUM (
            'supplier',
            'vendor',
            'customer_b2b',
            'carrier',
            'financial_institution',
            'other'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'portal_access_status') THEN
        CREATE TYPE portal_access_status AS ENUM (
            'none',
            'pending_link',
            'active',
            'revoked'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'address_usage_kind') THEN
        CREATE TYPE address_usage_kind AS ENUM (
            'registered_office',
            'headquarters',
            'billing',
            'shipping',
            'other'
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

-- 2. Utility Trigger Function (updated_at)
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 3. Master Alamat Bersama
CREATE TABLE IF NOT EXISTS addresses (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    label             varchar(120),
    address_line_1    varchar(240) NOT NULL,
    address_line_2    varchar(240),
    city              varchar(120) NOT NULL,
    province          varchar(120),
    postal_code       varchar(20),
    country_code      char(2) NOT NULL DEFAULT 'ID',
    formatted_address text,
    latitude          numeric(10,7),
    longitude         numeric(10,7),
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ck_addresses_line_1_nonempty CHECK (btrim(address_line_1) <> ''),
    CONSTRAINT ck_addresses_city_nonempty CHECK (btrim(city) <> ''),
    CONSTRAINT ck_addresses_country_code_len CHECK (char_length(btrim(country_code)) = 2),
    CONSTRAINT ck_addresses_latitude CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90),
    CONSTRAINT ck_addresses_longitude CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180)
);

CREATE INDEX IF NOT EXISTS ix_addresses_country_postal ON addresses (country_code, postal_code);
CREATE INDEX IF NOT EXISTS ix_addresses_city_province ON addresses (city, province);

-- 4. Pilar Internal Perusahaan (Legal Entity, Locations, Organization Units)
CREATE TABLE IF NOT EXISTS legal_entities (
    id                          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_legal_entity_id      uuid REFERENCES legal_entities(id) ON DELETE SET NULL,
    entity_code                 varchar(60) NOT NULL,
    legal_name                  varchar(240) NOT NULL,
    trading_name                varchar(240),
    legal_form                  varchar(40),
    is_primary_operating_entity boolean NOT NULL DEFAULT false,
    status                      legal_entity_status NOT NULL DEFAULT 'active',
    jurisdiction_country_code   char(2) NOT NULL DEFAULT 'ID',
    nib                         varchar(60),
    tax_id_number               varchar(32),
    is_pkp                      boolean NOT NULL DEFAULT false,
    registered_address_id       uuid REFERENCES addresses(id) ON DELETE SET NULL,
    fiscal_year_start_month     smallint NOT NULL DEFAULT 1,
    incorporation_date          date,
    dissolution_date            date,
    created_at                  timestamptz NOT NULL DEFAULT now(),
    updated_at                  timestamptz NOT NULL DEFAULT now(),
    deleted_at                  timestamptz,
    CONSTRAINT ck_legal_entities_code_nonempty CHECK (btrim(entity_code) <> ''),
    CONSTRAINT ck_legal_entities_legal_name_nonempty CHECK (btrim(legal_name) <> ''),
    CONSTRAINT ck_legal_entities_country_code_len CHECK (char_length(btrim(jurisdiction_country_code)) = 2),
    CONSTRAINT ck_legal_entities_tax_id_nonempty CHECK (tax_id_number IS NULL OR btrim(tax_id_number) <> ''),
    CONSTRAINT ck_legal_entities_not_self_parent CHECK (parent_legal_entity_id IS NULL OR parent_legal_entity_id <> id),
    CONSTRAINT ck_legal_entities_fiscal_month CHECK (fiscal_year_start_month BETWEEN 1 AND 12),
    CONSTRAINT ck_legal_entities_dates CHECK (
        dissolution_date IS NULL
        OR incorporation_date IS NULL
        OR dissolution_date >= incorporation_date
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_legal_entities_code
    ON legal_entities (entity_code)
    WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX IF NOT EXISTS uq_legal_entities_primary_operating
    ON legal_entities (is_primary_operating_entity)
    WHERE (is_primary_operating_entity = true AND deleted_at IS NULL);
CREATE UNIQUE INDEX IF NOT EXISTS uq_legal_entities_tax_id
    ON legal_entities (tax_id_number)
    WHERE (tax_id_number IS NOT NULL AND deleted_at IS NULL);
CREATE INDEX IF NOT EXISTS ix_legal_entities_parent ON legal_entities (parent_legal_entity_id);
CREATE INDEX IF NOT EXISTS ix_legal_entities_status ON legal_entities (status);
CREATE INDEX IF NOT EXISTS ix_legal_entities_deleted_at ON legal_entities (deleted_at);

CREATE TABLE IF NOT EXISTS locations (
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    legal_entity_id        uuid NOT NULL REFERENCES legal_entities(id) ON DELETE RESTRICT,
    parent_location_id     uuid REFERENCES locations(id) ON DELETE SET NULL,
    address_id             uuid REFERENCES addresses(id) ON DELETE RESTRICT,
    location_code          varchar(60) NOT NULL,
    name                   varchar(200) NOT NULL,
    location_kind          location_kind NOT NULL DEFAULT 'warehouse',
    phone                  varchar(60),
    timezone               varchar(80) NOT NULL DEFAULT 'Asia/Jakarta',
    operational_attributes jsonb NOT NULL DEFAULT '{}'::jsonb,
    is_active              boolean NOT NULL DEFAULT true,
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),
    deleted_at             timestamptz,
    CONSTRAINT ck_locations_code_nonempty CHECK (btrim(location_code) <> ''),
    CONSTRAINT ck_locations_name_nonempty CHECK (btrim(name) <> ''),
    CONSTRAINT ck_locations_timezone_nonempty CHECK (btrim(timezone) <> ''),
    CONSTRAINT ck_locations_not_self_parent CHECK (parent_location_id IS NULL OR parent_location_id <> id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_locations_code
    ON locations (location_code)
    WHERE (deleted_at IS NULL);
CREATE INDEX IF NOT EXISTS ix_locations_entity_kind_active ON locations (legal_entity_id, location_kind, is_active);
CREATE INDEX IF NOT EXISTS ix_locations_parent ON locations (parent_location_id);
CREATE INDEX IF NOT EXISTS ix_locations_address ON locations (address_id);
CREATE INDEX IF NOT EXISTS ix_locations_deleted_at ON locations (deleted_at);

CREATE TABLE IF NOT EXISTS organization_units (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    legal_entity_id     uuid NOT NULL REFERENCES legal_entities(id) ON DELETE RESTRICT,
    parent_unit_id      uuid REFERENCES organization_units(id) ON DELETE SET NULL,
    primary_location_id uuid REFERENCES locations(id) ON DELETE SET NULL,
    unit_code           varchar(60) NOT NULL,
    name                varchar(180) NOT NULL,
    unit_kind           org_unit_kind NOT NULL DEFAULT 'department',
    cost_center_code    varchar(60),
    head_person_id      uuid,
    is_active           boolean NOT NULL DEFAULT true,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    deleted_at          timestamptz,
    CONSTRAINT ck_organization_units_code_nonempty CHECK (btrim(unit_code) <> ''),
    CONSTRAINT ck_organization_units_name_nonempty CHECK (btrim(name) <> ''),
    CONSTRAINT ck_organization_units_not_self_parent CHECK (parent_unit_id IS NULL OR parent_unit_id <> id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_organization_units_entity_code
    ON organization_units (legal_entity_id, unit_code)
    WHERE (deleted_at IS NULL);
CREATE INDEX IF NOT EXISTS ix_organization_units_entity_active ON organization_units (legal_entity_id, is_active);
CREATE INDEX IF NOT EXISTS ix_organization_units_parent ON organization_units (parent_unit_id);
CREATE INDEX IF NOT EXISTS ix_organization_units_location ON organization_units (primary_location_id);
CREATE INDEX IF NOT EXISTS ix_organization_units_deleted_at ON organization_units (deleted_at);

-- 5. Pilar Mitra Eksternal (Supplier, Vendor, Pelanggan B2B, Ekspedisi)
CREATE TABLE IF NOT EXISTS parties (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_code          varchar(60) NOT NULL,
    party_kind            party_kind NOT NULL DEFAULT 'organization',
    display_name          varchar(240) NOT NULL,
    legal_name            varchar(240),
    email                 varchar(255),
    phone                 varchar(60),
    website_url           varchar(500),
    tax_id_number         varchar(32),
    tax_registered_name   varchar(240),
    is_pkp                boolean NOT NULL DEFAULT false,
    is_tax_exempt         boolean NOT NULL DEFAULT false,
    tax_exemption_reason  text,
    tax_verified_at       timestamptz,
    preferred_language    varchar(20) NOT NULL DEFAULT 'id',
    status                party_status NOT NULL DEFAULT 'pending_verification',
    notes                 text,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    deleted_at            timestamptz,
    CONSTRAINT ck_parties_partner_code_nonempty CHECK (btrim(partner_code) <> ''),
    CONSTRAINT ck_parties_display_name_nonempty CHECK (btrim(display_name) <> ''),
    CONSTRAINT ck_parties_tax_id_nonempty CHECK (tax_id_number IS NULL OR btrim(tax_id_number) <> ''),
    CONSTRAINT ck_parties_tax_exemption_reason CHECK (
        is_tax_exempt = false
        OR (tax_exemption_reason IS NOT NULL AND btrim(tax_exemption_reason) <> '')
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_parties_partner_code
    ON parties (partner_code)
    WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX IF NOT EXISTS uq_parties_tax_id_number
    ON parties (tax_id_number)
    WHERE (tax_id_number IS NOT NULL AND deleted_at IS NULL);
CREATE INDEX IF NOT EXISTS ix_parties_kind_status ON parties (party_kind, status);
CREATE INDEX IF NOT EXISTS ix_parties_display_name ON parties (display_name);
CREATE INDEX IF NOT EXISTS ix_parties_display_name_trgm
    ON parties USING gin (display_name gin_trgm_ops)
    WHERE (deleted_at IS NULL);
CREATE INDEX IF NOT EXISTS ix_parties_deleted_at ON parties (deleted_at);

CREATE TABLE IF NOT EXISTS party_roles (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    party_id       uuid NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    role           party_role_kind NOT NULL,
    is_active      boolean NOT NULL DEFAULT true,
    reference_note varchar(200),
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_party_roles_party_role UNIQUE (party_id, role)
);

CREATE INDEX IF NOT EXISTS ix_party_roles_role_active ON party_roles (role, is_active);

CREATE TABLE IF NOT EXISTS party_contacts (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    party_id             uuid NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    iam_user_id          uuid,
    full_name            varchar(200) NOT NULL,
    job_title            varchar(120),
    contact_role         varchar(60) NOT NULL DEFAULT 'general',
    email                varchar(255),
    phone                varchar(60),
    portal_access_status portal_access_status NOT NULL DEFAULT 'none',
    is_primary           boolean NOT NULL DEFAULT false,
    is_active            boolean NOT NULL DEFAULT true,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz,
    CONSTRAINT ck_party_contacts_full_name_nonempty CHECK (btrim(full_name) <> ''),
    CONSTRAINT ck_party_contacts_contact_role_nonempty CHECK (btrim(contact_role) <> ''),
    CONSTRAINT ck_party_contacts_has_channel_or_user CHECK (
        btrim(coalesce(email, '')) <> ''
        OR btrim(coalesce(phone, '')) <> ''
        OR iam_user_id IS NOT NULL
    ),
    CONSTRAINT ck_party_contacts_portal_link_invariant CHECK (
        (iam_user_id IS NULL AND portal_access_status = 'none')
        OR (iam_user_id IS NOT NULL AND portal_access_status <> 'none')
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_party_contacts_party_iam_user
    ON party_contacts (party_id, iam_user_id)
    WHERE (iam_user_id IS NOT NULL AND deleted_at IS NULL);
CREATE INDEX IF NOT EXISTS ix_party_contacts_iam_user
    ON party_contacts (iam_user_id)
    WHERE (iam_user_id IS NOT NULL AND deleted_at IS NULL);
CREATE INDEX IF NOT EXISTS ix_party_contacts_party_lookup
    ON party_contacts (party_id, contact_role, is_primary, is_active);
CREATE INDEX IF NOT EXISTS ix_party_contacts_deleted_at ON party_contacts (deleted_at);

CREATE TABLE IF NOT EXISTS party_addresses (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    party_id   uuid NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    address_id uuid NOT NULL REFERENCES addresses(id) ON DELETE RESTRICT,
    usage      address_usage_kind NOT NULL DEFAULT 'headquarters',
    is_primary boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_party_addresses_party_addr_usage UNIQUE (party_id, address_id, usage)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_party_addresses_primary_per_usage
    ON party_addresses (party_id, usage)
    WHERE (is_primary = true);
CREATE INDEX IF NOT EXISTS ix_party_addresses_address ON party_addresses (address_id);

-- 6. Outbox Sinkronisasi Lintas Layanan
CREATE TABLE IF NOT EXISTS foundation_outbox_events (
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
    CONSTRAINT ck_foundation_outbox_event_type_nonempty CHECK (btrim(event_type) <> ''),
    CONSTRAINT ck_foundation_outbox_aggregate_type_nonempty CHECK (btrim(aggregate_type) <> ''),
    CONSTRAINT ck_foundation_outbox_attempt_count CHECK (attempt_count >= 0)
);

CREATE INDEX IF NOT EXISTS ix_foundation_outbox_pending
    ON foundation_outbox_events (status, next_retry_at, created_at);
CREATE INDEX IF NOT EXISTS ix_foundation_outbox_aggregate
    ON foundation_outbox_events (aggregate_type, aggregate_id, created_at);

-- 7. Triggers for updated_at
DROP TRIGGER IF EXISTS trg_addresses_updated_at ON addresses;
CREATE TRIGGER trg_addresses_updated_at
    BEFORE UPDATE ON addresses
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_legal_entities_updated_at ON legal_entities;
CREATE TRIGGER trg_legal_entities_updated_at
    BEFORE UPDATE ON legal_entities
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_locations_updated_at ON locations;
CREATE TRIGGER trg_locations_updated_at
    BEFORE UPDATE ON locations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_organization_units_updated_at ON organization_units;
CREATE TRIGGER trg_organization_units_updated_at
    BEFORE UPDATE ON organization_units
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_parties_updated_at ON parties;
CREATE TRIGGER trg_parties_updated_at
    BEFORE UPDATE ON parties
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_party_roles_updated_at ON party_roles;
CREATE TRIGGER trg_party_roles_updated_at
    BEFORE UPDATE ON party_roles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_party_contacts_updated_at ON party_contacts;
CREATE TRIGGER trg_party_contacts_updated_at
    BEFORE UPDATE ON party_contacts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_party_addresses_updated_at ON party_addresses;
CREATE TRIGGER trg_party_addresses_updated_at
    BEFORE UPDATE ON party_addresses
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_foundation_outbox_events_updated_at ON foundation_outbox_events;
CREATE TRIGGER trg_foundation_outbox_events_updated_at
    BEFORE UPDATE ON foundation_outbox_events
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
