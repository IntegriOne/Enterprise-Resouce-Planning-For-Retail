# Simplifikasi Skema Enterprise Foundation untuk UMKM Retail & Integrasi IAM

- **Tanggal:** 2026-11-15 18:22 WIB
- **Cakupan:** [enterprise_foundation.dbml](../../backend/services/enterprise_foundation/db/schema/enterprise_foundation.dbml)

## Ringkasan Perubahan
- Merampingkan DBML dari 17 tabel menjadi **9 tabel inti** (`addresses`, `legal_entities`, `organization_units`, `locations`, `parties`, `party_roles`, `party_contacts`, `party_addresses`, `foundation_outbox_events`) untuk perusahaan ~10–50 staf (1 kantor utama, 1 gudang).
- Menghapus tabel/kolom berlebihan (`organization_unit_relationships`, `legal_entity_registrations`, `location_address_usages`, `party_tax_schemes`, `party_identifiers`, `party_relationships`, `external_identifiers`, `contact_points`, `lei`, `gln`, `plant`, `manufacturing`, `digital_endpoint`).
- Memasang NIB/NPWP/PKP internal pada `legal_entities`, menautkan `organization_units.primary_location_id`, dan memisahkan peran `supplier` vs `vendor` serta `party_status`.
- Menambahkan jembatan logis `party_contacts.iam_user_id` + `portal_access_status` dan tabel `foundation_outbox_events`, ditambah `gen_random_uuid()` serta partial unique index `WHERE deleted_at IS NULL`.
