# Perbaiki keputusan Procurement yang bergantung pada IAM dan Enterprise Foundation

- **ID** 5 · **Pemilik** Procurement & Sourcing · **Status** Siap dikerjakan.
- **Prioritas** Tinggi. **Baseline** Checkout bersama `e1d8766`. [Laporan audit dan bukti produksi](../../../agents/audit/procurement_sourcing/3-2026-11-28-210516-audit-kode-inti-iam-ef-procurement.md)

## Task

Kerjakan IEP-02, IEP-03, IEP-04, IEP-06 dan IEP-07 dari titik masuk yang dipakai Gateway sampai transaksi dan outbox. Hasil akhir harus dapat dijelaskan dari keputusan owner pada saat transaksi.

## Request

### IEP-02 — Otorisasi detail, daftar dan transisi di service

- file: [interceptor/interceptor.go L47](../../../backend/services/procurement_sourcing/transport/grpc/interceptor/interceptor.go#L47) · [handler/purchase_order_handler.go L228](../../../backend/services/procurement_sourcing/transport/grpc/handler/purchase_order_handler.go#L228) · [purchaseorder/service.go L667](../../../backend/services/procurement_sourcing/internal/service/purchaseorder/service.go#L667) · [partner/service.go L106](../../../backend/services/procurement_sourcing/internal/service/partner/service.go#L106) · [security/approval.go L25](../../../backend/services/procurement_sourcing/internal/security/approval.go#L25) · [cmd/wiring.go L88](../../../backend/services/procurement_sourcing/cmd/wiring.go#L88) · [procurement_sourcing.json](../../../backend/shared/rbac/catalogs/procurement_sourcing.json).
- function: `GetPurchaseOrder`, `ListPurchaseOrders`, `GetProcurementPartner`, `ListProcurementPartners`. Telaah ulang caller yang menggunakan jalur pembacaan sama, termasuk filter, hitung serta referensi turunan.
- logic: Hubungkan session reader IAM. Verifikasi principal, freshness, permission dan scope. Untuk detail, cocokkan badan hukum dari objek tepercaya. Untuk list, gabungkan cakupan grant dengan filter request sebelum query, count dan pagination. Pertahankan dukungan operator global yang memang berhak.
- error_handling: Tolak sesi dicabut atau grant tidak berlaku. Petakan kegagalan proyeksi ke unavailable. Gunakan respons akses di luar cakupan yang konsisten.
- security: Cek di service melalui jalur transport biasa. Pisahkan identitas manusia dan workload. Fase portal supplier/vendor mengikuti keputusan owner dan tetap dibatasi selama pengembangan internal.
- expected_result: Operator satu badan hukum hanya menerima objek dan jumlah dalam cakupannya. Panggilan privat yang konteksnya sah tetap memerlukan izin operasi.
- testing: Verifikasi pada service dan E2E HTTPS untuk izin global, izin satu badan hukum, dua entitas, grant dicabut, lookup ID luar scope, daftar tanpa filter, pagination dan session reader padam.

### IEP-03 dan IEP-07 — Referensi organisasi dan settings

- file: [service/ports.go L82](../../../backend/services/procurement_sourcing/domain/service/ports.go#L82) · [foundation/client.go L73](../../../backend/services/procurement_sourcing/internal/integration/foundation/client.go#L73) · [foundation/client.go L133](../../../backend/services/procurement_sourcing/internal/integration/foundation/client.go#L133) · [cmd/wiring.go L48](../../../backend/services/procurement_sourcing/cmd/wiring.go#L48) · [configuration/service.go L59](../../../backend/services/procurement_sourcing/internal/service/configuration/service.go#L59) · [requisition/service.go L53](../../../backend/services/procurement_sourcing/internal/service/requisition/service.go#L53) · [requisition/service.go L146](../../../backend/services/procurement_sourcing/internal/service/requisition/service.go#L146) · [purchaseorder/service.go L130](../../../backend/services/procurement_sourcing/internal/service/purchaseorder/service.go#L130) · [purchaseorder/service.go L910](../../../backend/services/procurement_sourcing/internal/service/purchaseorder/service.go#L910).
- function: `ConfigureProcurement`, `CreatePurchaseRequest`, perubahan draft, pembuatan/pembaruan revisi PO, dan validasi unit policy approval.
- logic: Tambahkan kedua pemeriksaan yang sudah ada ke port dan wiring. Minta konfirmasi entity, unit dan delivery location yang digunakan. Periksa ID, status, dan kepemilikan entity. Cakup perubahan atas draft lama serta verifikasi sebelum submit/issue. Saat port EF kosong, hentikan mutasi yang bergantung pada fakta tersebut.
- error_handling: Bedakan invalid reference, scope berbeda, dan unavailable. Batalkan transaksi beserta audit/outbox sukses apabila keputusan owner belum diperoleh.
- security: Badan hukum berasal dari dokumen dan scope terotorisasi. Gunakan workload privat yang disetujui EF.
- expected_result: Unit dan lokasi sah diterima di listener privat. Ketidaktersediaan verifier mencegah settings baru atau perubahan yang tak terverifikasi.
- testing: Uji di service untuk unit/lokasi lintas entity, inaktif, field opsional kosong, settings lama, port nil, timeout serta perbaikan draft lama. Verifikasi dua konfigurasi listener EF bersama pemiliknya.

### IEP-04 dan IEP-06 — Komitmen komersial dan role peninjau

- file: [partner/service.go L91](../../../backend/services/procurement_sourcing/internal/service/partner/service.go#L91) · [sourcing/service.go L563](../../../backend/services/procurement_sourcing/internal/service/sourcing/service.go#L563) · [service/registry.go L238](../../../backend/services/procurement_sourcing/internal/service/registry.go#L238) · [purchaseorder/service.go L419](../../../backend/services/procurement_sourcing/internal/service/purchaseorder/service.go#L419) · [entity/requisition_approval.go L33](../../../backend/services/procurement_sourcing/domain/entity/requisition_approval.go#L33) · [configuration/service.go L121](../../../backend/services/procurement_sourcing/internal/service/configuration/service.go#L121) · [approval/service.go L119](../../../backend/services/procurement_sourcing/internal/service/approval/service.go#L119) · [security/approval.go L25](../../../backend/services/procurement_sourcing/internal/security/approval.go#L25).
- function: `IssuePurchaseOrderRevision`, persetujuan akhir sourcing award, `CreateProcurementApprovalPolicy`, aktivasi policy dan `StartApprovalRound`.
- logic: Cocokkan ulang kualifikasi lokal dan party/role dari EF ketika menerbitkan komitmen baru. Bedakan penambahan kuantitas/nilai dari revisi korektif atau penurunan kewajiban. Gunakan bukti role aktif IAM sebelum mengaktifkan policy atau memulai putaran baru. Rekonsiliasi policy lama yang tidak valid. Pertahankan langkah dan snapshot historis.
- error_handling: Dependency padam menunda keputusan. Mitra atau role tidak layak menghasilkan konflik/alasan domain yang bisa ditindaklanjuti. Jangan publikasikan event sukses.
- security: Izin operator, maker-checker, MFA, digest, role efektif dan scope peninjau tetap diperiksa. Hindari akses database IAM/EF.
- expected_result: Mitra dinonaktifkan setelah kualifikasi tidak menerima komitmen baru. Kewajiban sebelumnya dapat diselesaikan. Referensi role bermasalah terdeteksi sebelum dokumen masuk antrean peninjau.
- testing: Verifikasi di service untuk status EF berubah antara draft dan issue, award split, kenaikan dan penurunan revisi, retur historis, role hilang/inaktif serta pencabutan grant setelah submit.

## Action

1. Selesaikan kontrol service, konfirmasikan kontrak pada kedua owner, lalu perbarui katalog gRPC Procurement untuk query, mutation, RBAC dan kegagalan owner yang benar-benar berubah.
2. Jalankan skenario penerimaan setelah implementasi. Untuk gRPC gunakan tes pada lapisan service sesuai aturan repositori. Untuk HTTPS, login superadmin dan siapkan role, permission, scope serta assignment lewat GraphQL IAM. Gunakan sesi aktor yang sesuai untuk memeriksa hak sebenarnya.
3. Catat hasil aktual untuk setiap ID, file yang berubah, kontrak antarpemilik, status data lama serta SHA. Pertahankan gate fase eksternal yang tercatat dalam katalog.

## Context

Baca [AGENTS.md](../../../AGENTS.md), [arsitektur](../../../agents/arsitektur.md), [boundary](../../../agents/boundary.md), [identitas dan akses](../../../agents/users.md), [cara coding gRPC](../../../agents/how_to_code_grpc_managements.md), [cara wiring Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md), [logika Procurement](../../../agents/logic/procurement/logic.md), [katalog gRPC Procurement](../../../agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md). Koordinasi antaragent mengikuti [instruksi Enterprise Foundation](../../../agents/instructions/enterprise_foundation/3-leader-remediasi-iam-ef-procurement-instructions.md), [instruksi IAM](../../../agents/instructions/iam/3-leader-remediasi-iam-ef-procurement-instructions.md), [instruksi Gateway](../../../agents/instructions/gateway/5-leader-remediasi-iam-ef-procurement-instructions.md).
