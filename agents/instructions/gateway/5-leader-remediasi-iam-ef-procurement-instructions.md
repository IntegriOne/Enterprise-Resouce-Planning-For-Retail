# Perbaiki rekonsiliasi akun mitra dan pemeriksaan scope di Gateway

- **ID** 5 · **Pemilik** Gateway · **Status** Siap dikerjakan.
- **Prioritas** Tinggi. **Baseline** Checkout bersama `e1d8766`. [Laporan audit dan bukti produksi](../../../agents/audit/procurement_sourcing/3-2026-11-28-210516-audit-kode-inti-iam-ef-procurement.md)

## Task

Kerjakan IEP-05 dan selaraskan IEP-02 dengan service Procurement. Pertahankan satu endpoint `/api/graphql` dan kontrak operasi yang telah dipakai UI.

## Request

### IEP-05 — Rekonsiliasi provisioning vendor dan supplier

- file: [iam/user_admin_adapter.go L738](../../../backend/gateway/internal/adapter/iam/user_admin_adapter.go#L738) · [iam/user_admin_adapter.go L783](../../../backend/gateway/internal/adapter/iam/user_admin_adapter.go#L783) · [iam/user_admin_adapter.go L865](../../../backend/gateway/internal/adapter/iam/user_admin_adapter.go#L865) · [enterprise_foundation/portal_provisioning.go L180](../../../backend/gateway/internal/adapter/enterprise_foundation/portal_provisioning.go#L180) · [cmd/main.go L321](../../../backend/gateway/cmd/main.go#L321) · [iam_user.graphql L173](../../../backend/gateway/graphql/iam/iam_user.graphql#L173) · [party/contact.go L305](../../../backend/services/enterprise_foundation/services/party/contact.go#L305).
- function: `createPartyPortalUser`, `VerifyPortalUserLink`, `rollbackUnlinkedUser`, `createVendorUser`, `createSupplierUser`.
- logic: Teruskan identitas operasi stabil ke IAM dan EF. Gunakan hasil kanonik berdasarkan operasi, party, contact bila tersedia, dan principal. Ketiadaan hasil sesaat adalah status menunggu. Resume sesudah timeout, restart, atau respons hilang dari ID operasi yang sama. Dukung principal dengan beberapa party. Tampilkan status dan nama kontak/party di UI melalui data yang diotorisasi.
- error_handling: Gunakan tenggat terbatas untuk pembacaan ulang. Izinkan kompensasi hanya sesudah konfirmasi terminal atas operasi dan pembuktian akun IAM memang dapat dikompensasi. Petakan status ambigu ke hasil yang dapat dipulihkan.
- security: Jaga izin IAM, MFA dan scope EF pada seluruh retry. Hanya konfirmasikan link dengan party, principal, contact dan status yang sesuai. Jangan log sandi atau token.
- expected_result: Dua pengiriman membawa ke akun/tautan yang sama. Commit EF yang baru terlihat sesudah timeout tetap dapat dipulihkan, termasuk ketika ada kontak lain milik principal yang sama.
- testing: Sesudah perbaikan, verifikasi HTTPS untuk kontak lama/baru, duplicate submit, hilangnya respons, commit EF terlambat, restart dan akses scope yang dicabut. Periksa satu identitas/tautan melalui query owner.

### IEP-02 — Grant terbatas pada pembacaan Procurement

- file: [procurement_sourcing/permissions.go L35](../../../backend/gateway/internal/adapter/procurement_sourcing/permissions.go#L35) · [procurement_sourcing/adapter.go L52](../../../backend/gateway/internal/adapter/procurement_sourcing/adapter.go#L52) · [procurement_sourcing/orders_adapter.go L181](../../../backend/gateway/internal/adapter/procurement_sourcing/orders_adapter.go#L181) · [resolvers/procurement_sourcing_orders.resolvers.go L141](../../../backend/gateway/graph/resolvers/procurement_sourcing_orders.resolvers.go#L141) · [procurement_sourcing_orders.graphql L346](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L346) · [procurement_sourcing_core.graphql L365](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L365) · [security/approval.go L25](../../../backend/services/procurement_sourcing/internal/security/approval.go#L25).
- function: `PurchaseOrder`, `PurchaseOrders`, `ProcurementPartners`, serta otorisasi adapter yang mengawali RPC detail/list.
- logic: Setelah pemilik service menegakkan filter server-side, terima grant legal entity/unit hanya pada operasi yang mendukungnya. Teruskan sesi dan filter yang diminta. Validasi kepemilikan dan irisan scope pada service sesuai record. Biarkan pemeriksaan keputusan approval tetap menuju policy dan snapshot IAM terkini.
- error_handling: Gunakan respons konsisten untuk lookup di luar cakupan. Hentikan request saat sesi kedaluwarsa atau proyeksi IAM tidak dapat diverifikasi.
- security: Pengaktifan akses terbatas harus mengikuti readiness service penerima. Selama gate F2 tertutup, jenis akun mitra tidak memperoleh jalur Procurement melalui relaksasi grant global.
- expected_result: Petugas sah dapat membaca PO/mitra dalam wilayah yang ditugaskan, dan filter kosong, detail serta pagination tidak memperluas wilayah tersebut.
- testing: Jalankan HTTPS untuk dua entity, dua sesi petugas, operator global, grant yang dicabut dan akun mitra saat gate F2 tertutup. Periksa detail/daftar yang kosong dan hasil pagination.

## Action

1. Integrasikan kontrak owner terlebih dahulu. Jika SDL perlu berubah, baca skema dasar, pertahankan caller lama dan jalankan generator resmi. Laporkan status per temuan dengan bukti operasi dan SHA.
2. Jalankan skenario penerimaan setelah implementasi. Untuk gRPC gunakan tes pada lapisan service sesuai aturan repositori. Untuk HTTPS, login superadmin dan siapkan role, permission, scope serta assignment lewat GraphQL IAM. Gunakan sesi aktor yang sesuai untuk memeriksa hak sebenarnya.
3. Catat hasil aktual untuk setiap ID, file yang berubah, kontrak antarpemilik, status data lama serta SHA. Pertahankan gate fase eksternal yang tercatat dalam katalog.

## Context

Baca [AGENTS.md](../../../AGENTS.md), [arsitektur](../../../agents/arsitektur.md), [boundary](../../../agents/boundary.md), [identitas dan akses](../../../agents/users.md), [cara coding gRPC](../../../agents/how_to_code_grpc_managements.md), [cara wiring Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md), [logika Procurement](../../../agents/logic/procurement/logic.md), [katalog gRPC Procurement](../../../agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md). Koordinasi antaragent mengikuti [instruksi Procurement & Sourcing](../../../agents/instructions/procurement_sourcing/5-leader-remediasi-iam-ef-procurement-instructions.md), [instruksi Enterprise Foundation](../../../agents/instructions/enterprise_foundation/3-leader-remediasi-iam-ef-procurement-instructions.md), [instruksi IAM](../../../agents/instructions/iam/3-leader-remediasi-iam-ef-procurement-instructions.md).
