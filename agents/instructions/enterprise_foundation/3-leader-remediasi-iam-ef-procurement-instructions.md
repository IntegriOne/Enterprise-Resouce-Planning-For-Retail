# Lindungi mandat kontak mitra dan selaraskan kontrak EF dengan Procurement

- **ID** 3 · **Pemilik** Enterprise Foundation · **Status** Siap dikerjakan.
- **Prioritas** Tinggi. **Baseline** Checkout bersama `e1d8766`. [Laporan audit dan bukti produksi](../../../agents/audit/procurement_sourcing/3-2026-11-28-210516-audit-kode-inti-iam-ef-procurement.md)

## Task

Kerjakan IEP-01, bagian EF dari IEP-03, dan pembuktian hasil penautan untuk IEP-05. EF tetap menjadi sumber identitas party, peran, kontak, unit, lokasi dan badan hukum.

## Request

### IEP-01 — Serialisasi perubahan kontak dan akses portal

- file: [party/contact.go L214](../../../backend/services/enterprise_foundation/services/party/contact.go#L214) · [party/contact.go L305](../../../backend/services/enterprise_foundation/services/party/contact.go#L305) · [party/contact.go L373](../../../backend/services/enterprise_foundation/services/party/contact.go#L373) · [party/contact.go L439](../../../backend/services/enterprise_foundation/services/party/contact.go#L439) · [party/party_contact_repository.go L32](../../../backend/services/enterprise_foundation/repository/party/party_contact_repository.go#L32) · [party/party_contact_repository.go L273](../../../backend/services/enterprise_foundation/repository/party/party_contact_repository.go#L273) · [pgstore/tx.go L78](../../../backend/services/enterprise_foundation/repository/pgstore/tx.go#L78) · [enterprise_foundation.dbml L305](../../../backend/services/enterprise_foundation/db/schema/enterprise_foundation.dbml#L305).
- function: `UpdateContact`, `RequestPortalLink`, `ActivatePortalAccess`, `RevokePortalAccess`, `UnlinkPortalUser`, perubahan aktif/primary, serta repository yang menulis record kontak.
- logic: Pisahkan kolom profile dari kolom IAM/portal. Gunakan lock row atau compare-and-set versi lama pada seluruh writer yang berbagi record. Evaluasi syarat transisi terhadap nilai terkini. Pertahankan keunikan per party/principal, perubahan primary, dan event portal dalam transaksi yang sama.
- error_handling: Tangani versi usang sebagai konflik yang bisa dibaca ulang. Rollback mutasi dan outbox jika penulisan bersyarat gagal. Retry identik menjaga satu hasil bisnis.
- security: Perubahan data kontak biasa hanya menyentuh atribut profil. Simpan identitas pelaku transisi dan minimalkan data kontak di log.
- expected_result: Edit profil bersamaan dengan pencabutan tidak mengaktifkan ulang akses atau mengembalikan principal lama. Hasil akhir dan event menjelaskan transisi yang sama.
- testing: Verifikasi lewat service dengan interleaving update profil/revoke, deactivate/activate, link dua principal untuk PIC yang sama dan pengulangan. Lakukan re-read melalui operasi EF yang digunakan Gateway.

### IEP-03 dan IEP-05 — Lookup privat yang tepat sasaran

- file: [grpc/workload_profiles.go L39](../../../backend/services/enterprise_foundation/transport/grpc/workload_profiles.go#L39) · [grpc/procurement_workload.go L62](../../../backend/services/enterprise_foundation/transport/grpc/procurement_workload.go#L62) · [grpc/server.go L74](../../../backend/services/enterprise_foundation/transport/grpc/server.go#L74) · [party/contact.go L96](../../../backend/services/enterprise_foundation/services/party/contact.go#L96) · [party/party_contact_repository.go L108](../../../backend/services/enterprise_foundation/repository/party/party_contact_repository.go#L108) · [party.proto](../../../backend/services/enterprise_foundation/contract/proto/enterprise_foundation/v1/party.proto).
- function: `workloadProfileGuard`, `procurementWorkloadGuard`, lookup unit/lokasi dan pengamatan hasil provisioning PIC.
- logic: Izinkan hanya dua metode baca unit/lokasi yang telah disepakati pada kedua profil listener Procurement. Kembalikan ID, entity, status sesuai record. Sediakan pembacaan tepat party/principal/contact termasuk status pending dan hasil final operasi untuk rekonsiliasi. Pasang identitas operasi bila diperlukan untuk membedakan penolakan terminal dan commit yang masih berjalan.
- error_handling: Bedakan belum teramati, berhasil, ditolak terminal dan unavailable. Tolak identitas workload/metode yang tidak berhak.
- security: Jaga SAN/SPIFFE, transport mTLS, tanda tangan konteks dan izin per metode. Fakta mandat portal untuk fase eksternal dibatasi principal/party/role.
- expected_result: Referensi sah dapat diverifikasi lewat kedua listener. Gateway dapat memulihkan PIC yang persis diminta, termasuk principal yang memiliki mandat pada beberapa party.
- testing: Verifikasi workload disetujui/ditolak, hasil unit milik entity lain, kontak pending/aktif/dicabut, dan satu principal dengan beberapa mandat.

## Action

1. Dahulukan penulisan kontak yang aman, tetapkan kontrak hasil penautan bersama IAM/Gateway, lalu koordinasikan allowlist sebelum Procurement mengaktifkan pemeriksaan referensi.
2. Jalankan skenario penerimaan setelah implementasi. Untuk gRPC gunakan tes pada lapisan service sesuai aturan repositori. Untuk HTTPS, login superadmin dan siapkan role, permission, scope serta assignment lewat GraphQL IAM. Gunakan sesi aktor yang sesuai untuk memeriksa hak sebenarnya.
3. Catat hasil aktual untuk setiap ID, file yang berubah, kontrak antarpemilik, status data lama serta SHA. Pertahankan gate fase eksternal yang tercatat dalam katalog.

## Context

Baca [AGENTS.md](../../../AGENTS.md), [arsitektur](../../../agents/arsitektur.md), [boundary](../../../agents/boundary.md), [identitas dan akses](../../../agents/users.md), [cara coding gRPC](../../../agents/how_to_code_grpc_managements.md), [cara wiring Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md), [logika Procurement](../../../agents/logic/procurement/logic.md), [katalog gRPC Procurement](../../../agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md). Koordinasi antaragent mengikuti [instruksi Procurement & Sourcing](../../../agents/instructions/procurement_sourcing/5-leader-remediasi-iam-ef-procurement-instructions.md), [instruksi IAM](../../../agents/instructions/iam/3-leader-remediasi-iam-ef-procurement-instructions.md), [instruksi Gateway](../../../agents/instructions/gateway/5-leader-remediasi-iam-ef-procurement-instructions.md).
