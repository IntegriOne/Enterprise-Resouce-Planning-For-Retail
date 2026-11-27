-- ============================================================================
-- Migrasi Baseline v000001 (DOWN) — Enterprise Foundation IntegriOne
-- Menghapus seluruh objek skema Enterprise Foundation secara terbalik dari dependensinya.
-- ============================================================================

DROP TRIGGER IF EXISTS trg_foundation_outbox_events_updated_at ON foundation_outbox_events;
DROP TRIGGER IF EXISTS trg_party_addresses_updated_at ON party_addresses;
DROP TRIGGER IF EXISTS trg_party_contacts_updated_at ON party_contacts;
DROP TRIGGER IF EXISTS trg_party_roles_updated_at ON party_roles;
DROP TRIGGER IF EXISTS trg_parties_updated_at ON parties;
DROP TRIGGER IF EXISTS trg_organization_units_updated_at ON organization_units;
DROP TRIGGER IF EXISTS trg_locations_updated_at ON locations;
DROP TRIGGER IF EXISTS trg_legal_entities_updated_at ON legal_entities;
DROP TRIGGER IF EXISTS trg_addresses_updated_at ON addresses;

DROP TABLE IF EXISTS foundation_outbox_events CASCADE;
DROP TABLE IF EXISTS party_addresses CASCADE;
DROP TABLE IF EXISTS party_contacts CASCADE;
DROP TABLE IF EXISTS party_roles CASCADE;
DROP TABLE IF EXISTS parties CASCADE;
DROP TABLE IF EXISTS organization_units CASCADE;
DROP TABLE IF EXISTS locations CASCADE;
DROP TABLE IF EXISTS legal_entities CASCADE;
DROP TABLE IF EXISTS addresses CASCADE;

DROP FUNCTION IF EXISTS set_updated_at() CASCADE;

DROP TYPE IF EXISTS outbox_status CASCADE;
DROP TYPE IF EXISTS address_usage_kind CASCADE;
DROP TYPE IF EXISTS portal_access_status CASCADE;
DROP TYPE IF EXISTS party_role_kind CASCADE;
DROP TYPE IF EXISTS party_status CASCADE;
DROP TYPE IF EXISTS party_kind CASCADE;
DROP TYPE IF EXISTS location_kind CASCADE;
DROP TYPE IF EXISTS org_unit_kind CASCADE;
DROP TYPE IF EXISTS legal_entity_status CASCADE;
