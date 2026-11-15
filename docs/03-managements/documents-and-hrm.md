# Document & Records dan Human Resource Management

> **Dokumen management retail · 12 November 2026.** Kepemilikan bisnis normatif mengikuti [agents/boundary.md](../../agents/boundary.md). Ini ringkasan, bukan instruksi melakukan perubahan lintas management atau seed IAM otomatis.

## DRM: bukti, retensi, dan keamanan anti-IDOR
DRM menguasai metadata dokumen, versi immutable, digest, klasifikasi sensitivitas, legal hold, retention dan tautan ke objek owner lintas departemen — antara lain **slip gaji karyawan & kontrak kerja** (HRM), **invoice customer setelah pembayaran produk berhasil** dan bukti AP/AR/faktur pajak (Finance/Commerce), serta **RFQ, penawaran, kontrak maklun CMT, dan PO** (Procurement/Supplier/Vendor).

Upload dokumen supplier, bukti pembayaran atau return **tidak** membuat PO fulfilled, invoice paid atau refund settled; keputusan domain tetap di Procurement, WMS, Finance/OMS.

**Kewajiban Anti-IDOR & Anti-*Privilege Escalation* (RBAC + ABAC):** Setiap akses metadata, pratinjau, maupun unduhan `document_id`/`version_id` wajib diverifikasi di server/DRM berdasarkan IAM `principal_id`, permission RBAC, `owner_management`, `legal_entity_id`, `party_id`, serta kepemilikan objek secara *fail-closed*. Customer tidak boleh pernah dapat mengakses slip gaji karyawan internal atau invoice customer lain lewat tebakan ID; supplier dan vendor CMT hanya dapat mengakses dokumen sesuai `party_id` miliknya. Di frontend, DRM diserap sebagai sub-layanan di dashboard terkait dan `/profile` terotorisasi (tanpa route `/documents` mandiri).

DRM memiliki migration/kerangka pada audit awal; jangan mengklaim scan atau sertifikasi regulator aktif hanya karena tabel disediakan.

## HRM: pegawai dan status hubungan kerja
HRM menguasai worker, kontrak, penempatan, cuti/kompensasi sesuai domain aktual. Referensi `iam_principal_reference_id` adalah hubungan logis, worker ID bukan user IAM. HRM boleh **meminta** role/grant dan menandai perubahan penempatan, tetapi IAM tetap sumber keputusan akses, revoke dan sesi.

## F1/F2
Employee onboarding→HRM persetujuan hubungan kerja→IAM access request/independent approval→active role scoped enterprise unit. Supplier/vendor **bukan** HRM worker hanya karena menerima akses portal. Kasus PII employee atau kontrak membutuhkan audit/retensi serta prinsip least privilege.

## Test dan status
Kedua domain memerlukan service tests dan E2E sesuai cakupan. Dokumen source/audit HRM yang lebih baru memegang detail aktual; jangan menganggap status planned DRM sudah hilang tanpa bukti.

[DRM](../../backend/services/document_records_management/README.md) · [HRM](../../backend/services/human_resource_management/README.md) · [source principal](../../agents/principals.md).
