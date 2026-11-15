# Konteks Aktif Enterprise Foundation (`backend/services/enterprise_foundation`)

## Apa yang sudah dikerjakan

- **Perancangan Ulang DBML Ramping untuk Retail UMKM (10–50 Staf) & Integrasi IAM:**
  - Menyederhanakan [enterprise_foundation.dbml](../../backend/services/enterprise_foundation/db/schema/enterprise_foundation.dbml) dari 17 tabel menjadi **9 tabel inti** dan 9 enum sesuai target operasional 1 kantor utama, 1 gudang, dan mitra eksternal.
  - **Pilar Internal:** `legal_entities` (menyerap NIB, NPWP, `is_pkp`, dan `registered_address_id` -> target `scope_type = 'legal_entity'`), `organization_units` (UUID stabil tanpa SCD2 berisiko, pohon `parent_unit_id`, dan `primary_location_id` -> target `scope_type = 'organization_unit'`), serta `locations` (`headquarters`, `warehouse`, `retail_store`, `branch_office`, menunjuk `address_id` langsung -> target `scope_type = 'facility'`).
  - **Pilar Mitra Eksternal & Koneksi IAM:** `parties` (status lifecycle `party_status`, email/phone bisnis, dan NPWP/PKP ringkas; khusus mitra/B2B, bukan pembeli ritel B2C), `party_roles` (memisahkan eksplisit `supplier` -> IAM `supplier_party` dan `vendor` -> IAM `vendor_party`), `party_contacts` (menambah referensi logis `iam_user_id` ke `iam.users.id` dan `portal_access_status` `none`/`pending_link`/`active`/`revoked`), serta `party_addresses`.
  - **Infrastruktur & Konsistensi PostgreSQL:** Menambah `foundation_outbox_events` (pola seragam `iam_outbox_events`), seluruh PK memakai `default: gen_random_uuid()`, indeks duplikat dihapus, dan partial unique index `WHERE deleted_at IS NULL` didokumentasikan untuk migrasi `000001`.

## Apa yang belum dikerjakan

- Penulisan baseline SQL `000001_*.up.sql` / `000001_*.down.sql` beserta `schema.sql` dan `db/init.go` (`enterprise_foundation` & `enterprise_foundation_test`).
- Implementasi service Go, kontrak protobuf gRPC internal (`legal_entity`, `organization_unit`, `location`, `party`, `party_contact` lookup), adapter IAM/Gateway, dan operasi GraphQL pada `/api/graphql`.
- Pengisian seed awal operasional (badan hukum utama, kantor utama, gudang utama, unit organisasi awal) yang aman dan idempoten.

## Apa yang harus dikerjakan selanjutanya

1. Membentuk struktur `backend/services/enterprise_foundation/db/` (`migrations/000001_create_enterprise_foundation_schema.{up,down}.sql`, `schema/schema.sql`, `init.go`) serta memastikan trigger `updated_at` dan seluruh partial unique index PostgreSQL terpasang bersih.
2. Mendefinisikan kontrak gRPC privat `enterprise_foundation` agar IAM/Gateway dapat memvalidasi `scope_id` (`legal_entity`, `organization_unit`, `facility`, `supplier_party`, `vendor_party`) serta menautkan PIC `createVendorUser` / `createSupplierUser` ke `party_contacts.iam_user_id`.
