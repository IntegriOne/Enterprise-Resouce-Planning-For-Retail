# Sediakan bukti role approval dan pemulihan akun mitra di IAM

- **ID** 3 · **Pemilik** IAM · **Status** Siap dikerjakan.
- **Prioritas** Tinggi untuk IEP-05, sedang untuk IEP-06. **Baseline** Checkout bersama `e1d8766`. [Laporan audit dan bukti produksi](../../../agents/audit/procurement_sourcing/3-2026-11-28-210516-audit-kode-inti-iam-ef-procurement.md)

## Task

Kerjakan sisi IAM agar perubahan status provisioning vendor/supplier dapat dipulihkan dan Procurement dapat memverifikasi role yang dipakai policy. Pertahankan verifikasi scope assignment melalui EF.

## Request

### IEP-05 — Identitas operasi provisioning partner

- file: [user/service.go L185](../../../backend/services/iam/services/user/service.go#L185) · [iam/user_admin_adapter.go L756](../../../backend/gateway/internal/adapter/iam/user_admin_adapter.go#L756) · [iam/user_admin_adapter.go L783](../../../backend/gateway/internal/adapter/iam/user_admin_adapter.go#L783) · [iam/user_admin_adapter.go L865](../../../backend/gateway/internal/adapter/iam/user_admin_adapter.go#L865) · [party/contact.go L305](../../../backend/services/enterprise_foundation/services/party/contact.go#L305) · [iam_user.graphql L173](../../../backend/gateway/graphql/iam/iam_user.graphql#L173).
- function: `CreateUser` untuk vendor/supplier dan kontrak privat pemulihan pembuatan akun partner. Koordinasikan dengan `createVendorUser` dan `createSupplierUser`.
- logic: Persistensikan ID operasi, party, jenis role, kontak opsional, account ID dan status. Kaitkan penciptaan akun pending dan pencatatan operasi dalam satu transaksi. Replay identik melanjutkan akun yang sama. Pemulihan harus dapat mencari operasi setelah timeout atau restart. Sepakati terminal/retry dengan EF.
- error_handling: Kembalikan konflik untuk pemakaian ID sama dengan fakta berbeda, dan status menunggu bila belum ada keputusan EF. Syarat kompensasi harus melindungi akun yang sudah tertaut atau digunakan.
- security: Batasi pembacaan/pemulihan sesuai izin administrator dan scope. Jaga data kredensial di IAM. Jangan menyimpan password atau digest turunan sandi sebagai identitas pengajuan.
- expected_result: Satu upaya logis menghasilkan satu principal, dapat dipulihkan setelah crash, dan hasil sementara tidak memicu penghapusan akun yang sah.
- testing: Verifikasi di service untuk pengiriman ganda, dua operator, putus sesudah commit IAM, respons EF tertunda, konflik payload dan pemulihan sesudah restart. Pastikan aktivitas sah akun yang sudah tertaut terlindungi.

### IEP-06 — Verifikasi role policy oleh workload Procurement

- file: [role.proto L184](../../../backend/services/iam/contract/proto/iam/v1/role.proto#L184) · [role/service.go L493](../../../backend/services/iam/services/role/service.go#L493) · [group/service.go L378](../../../backend/services/iam/services/group/service.go#L378) · [scopeverifier/ef_verifier.go L143](../../../backend/services/iam/infrastructure/scopeverifier/ef_verifier.go#L143) · [configuration/service.go L121](../../../backend/services/procurement_sourcing/internal/service/configuration/service.go#L121) · [security/approval.go L25](../../../backend/services/procurement_sourcing/internal/security/approval.go#L25) · [procurement_sourcing.dbml L235](../../../backend/services/procurement_sourcing/db/schema/procurement_sourcing.dbml#L235).
- function: Kontrak owner untuk keputusan status role approval Procurement. Evaluasi kemampuan kontrak baca role yang ada dan pembatasan workload sebelum memilih penambahan RPC.
- logic: Jawab role ID kanonik dan status layak/aktif untuk digunakan sebagai rujukan policy. Setujui bersama Procurement perbedaan antara validitas role dan kemampuan principal tertentu untuk memutuskan. Gunakan proyeksi/kontrak resmi dengan freshness yang terukur. Keputusan assignment scope EF yang sudah ada tetap independen.
- error_handling: Bedakan role tak ditemukan/inaktif dari outage. Permintaan dengan identitas tidak cocok ditolak. Berikan error yang dapat dipetakan sebelum putaran dimulai.
- security: Batasi audience dan metode, gunakan autentikasi workload privat. Hanya kembalikan data role minimal yang diperlukan. Izin efektif, MFA dan SoD tetap diperiksa saat keputusan.
- expected_result: Procurement mendeteksi referensi role yang tidak mungkin dipenuhi sebelum menggunakan policy. Verifikasi role tidak memberikan grant baru kepada siapa pun.
- testing: Verifikasi service IAM untuk role aktif, role inaktif/dihapus, scope sah dan tidak sah, outage serta identitas workload keliru.

## Action

1. Sepakati kontrak bersama Gateway/EF dan Procurement. Inventarisasi akun partner lama yang perlu rekonsiliasi berdasarkan bukti owner. Implementasikan alur kompatibel untuk instalasi aktif.
2. Jalankan skenario penerimaan setelah implementasi. Untuk gRPC gunakan tes pada lapisan service sesuai aturan repositori. Untuk HTTPS, login superadmin dan siapkan role, permission, scope serta assignment lewat GraphQL IAM. Gunakan sesi aktor yang sesuai untuk memeriksa hak sebenarnya.
3. Catat hasil aktual untuk setiap ID, file yang berubah, kontrak antarpemilik, status data lama serta SHA. Pertahankan gate fase eksternal yang tercatat dalam katalog.

## Context

Baca [AGENTS.md](../../../AGENTS.md), [arsitektur](../../../agents/arsitektur.md), [boundary](../../../agents/boundary.md), [identitas dan akses](../../../agents/users.md), [cara coding gRPC](../../../agents/how_to_code_grpc_managements.md), [cara wiring Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md), [logika Procurement](../../../agents/logic/procurement/logic.md), [katalog gRPC Procurement](../../../agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md). Koordinasi antaragent mengikuti [instruksi Procurement & Sourcing](../../../agents/instructions/procurement_sourcing/5-leader-remediasi-iam-ef-procurement-instructions.md), [instruksi Enterprise Foundation](../../../agents/instructions/enterprise_foundation/3-leader-remediasi-iam-ef-procurement-instructions.md), [instruksi Gateway](../../../agents/instructions/gateway/5-leader-remediasi-iam-ef-procurement-instructions.md).
