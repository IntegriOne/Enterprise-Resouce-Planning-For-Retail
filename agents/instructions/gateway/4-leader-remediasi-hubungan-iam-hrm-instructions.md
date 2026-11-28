# Perbaikan pengajuan akun employee di GraphQL Gateway

- **ID** 4 · **Pemilik** Agent Gateway · **Status** Siap dikerjakan.
- **Prioritas** Tinggi untuk pengiriman pengajuan IH-03. Sedang untuk IH-04 dan IH-05.
- **Baseline** Checkout bersama `0fd7bb5`. Bukti dan kepemilikan ada pada [audit IAM–HRM](../../../agents/audit/iam/18-2026-11-28-205822-audit-kode-inti-hubungan-iam-hrm.md).

## Task

Sambungkan perubahan owner IAM dan HRM ke operasi GraphQL yang digunakan HR dan admin IAM. Pertahankan penautan otomatis pada `createEmployeeUser` dan berikan status yang dapat dipahami sesuai hasil yang sudah dikonfirmasi owner.

## Request

### IH-03 dan IH-04 — Samakan pengiriman dari browser dan relay

- file: [access_request.go L20](../../../backend/gateway/internal/adapter/human_resource/access_request.go#L20) · [user_admin_adapter.go L568](../../../backend/gateway/internal/adapter/iam/user_admin_adapter.go#L568) · [worker_provisioning.go L60](../../../backend/gateway/internal/adapter/human_resource/worker_provisioning.go#L60) · [hrm_access_request.graphql L22](../../../backend/gateway/graphql/human_resource/hrm_access_request.graphql#L22) · [iam_employee_provisioning.graphql L25](../../../backend/gateway/graphql/iam/iam_employee_provisioning.graphql#L25) · [iam_user.graphql L342](../../../backend/gateway/graphql/iam/iam_user.graphql#L342) · [employee_provisioning.proto L10](../../../backend/services/iam/contract/proto/iam/v1/employee_provisioning.proto#L10).
- function: `RequestEmployeeAccount`, `EmployeeAccountRequest`, `WorkerEmployeeAccountRequest`, `CreateEmployeeUser`, serta pemetaan hasil receipt IAM ke model GraphQL.
- logic: Ambil identitas pengajuan dan bukti lifecycle dari respons kanonik HRM. Kirim bukti sesuai kontrak yang disepakati kedua owner. Pertahankan satu ID untuk retry. Bedakan pengajuan tercatat di HRM, menunggu pengiriman, sudah diterima IAM, sedang diproses, dan sudah terhubung. Bila owner mengimplementasikan penarikan, tampilkan keputusan terminal hanya setelah terkonfirmasi.
- error_handling: Saat respons hilang sesudah commit, baca ulang memakai identitas yang sama. Kembalikan status sementara yang jujur bila belum ada konfirmasi. Bedakan konflik fakta dari gangguan koneksi. Hindari penerbitan akun kedua sebagai pemulihan.
- security: Pertahankan MFA, validasi worker terhadap receipt, serta pemrosesan admin berizin. Tautan worker/principal merupakan langkah internal. Tampilkan nomor pekerja, nama dan status sesuai scope pada data untuk UI. ID Relay tetap tersedia bagi klien sebagai referensi teknis.
- expected_result: Pengiriman sinkron dan asinkron bermuara pada satu receipt. Replay dapat diselesaikan dengan akun yang sama dan `wasSuccessful` hanya dikembalikan sesudah tahapan wajib selesai.
- testing: Seusai implementasi, verifikasi HTTPS pengiriman sukses, response timeout, replay setelah pencatatan HRM/IAM, konflik bukti, receipt terminal dan retry setelah link. Gunakan satu-satunya endpoint aplikasi `/api/graphql`.

### IH-05 — Scope HR sesuai badan hukum worker

- file: [access_request.go L20](../../../backend/gateway/internal/adapter/human_resource/access_request.go#L20) · [access_request.go L129](../../../backend/gateway/internal/adapter/human_resource/access_request.go#L129) · [worker.go L116](../../../backend/gateway/internal/adapter/human_resource/worker.go#L116) · [payroll_run.go L18](../../../backend/gateway/internal/adapter/human_resource/payroll_run.go#L18) · [hrm_access_request.graphql L26](../../../backend/gateway/graphql/human_resource/hrm_access_request.graphql#L26).
- function: `RequestEmployeeAccount`, `EmployeeAccountRequest`, `WorkerEmployeeAccountRequest`, termasuk finder idempotensi dan pembacaan status IAM.
- logic: Gunakan hasil resolusi worker yang diotorisasi HRM untuk menentukan badan hukum target. Dukung grant HR global atau grant sesuai badan hukum yang tepat. Untuk lookup berdasarkan ID pengajuan, gunakan jalur HRM yang membatasi record sesuai scope sejak awal. Selaraskan create, retry, lookup dan status. Antrean provisioning admin IAM mengikuti izin administratifnya sendiri.
- error_handling: Kembalikan respons konsisten untuk akses di luar cakupan agar detail worker atau pengajuan tidak terungkap. Petakan IAM/HRM unavailable sebagai dependency sementara.
- security: Validasi sesi employee, MFA yang dibutuhkan, scope serta revisi izin. Pertahankan perlindungan CSRF. Jangan memperluas grant petugas agar operasi bisa lolos.
- expected_result: HR dengan grant hanya badan hukum A dapat menyelesaikan pengajuan dan membaca status untuk A. Pengajuan badan hukum B tetap terlindungi pada setiap query, mutation dan retry.
- testing: Untuk E2E, login superadmin dan kelola role, permission, scope dan assignment melalui GraphQL IAM. Gunakan sesi petugas berscope A untuk bukti least privilege, petugas B untuk batas antarunit, serta sesi admin IAM untuk memproses. Catat operasi, scope dan hasil tanpa token/sandi.

## Action

1. Baca SDL utama, SDL IAM dan HRM, kontrak versi aktif serta rencana kompatibilitas owner. Pertahankan konvensi Node/Connection dan bentuk payload untuk caller lama.
2. Implementasikan adapter dan resolver yang diperlukan setelah kontrak owner tersedia. Regenerasikan artefak melalui generator resmi apabila kontrak berubah.
3. Verifikasi hasil dari pengajuan sampai link dan pembacaan status melalui HTTPS. Laporkan hasil per temuan, file dan SHA.

## Context

Baca [AGENTS.md](../../../AGENTS.md), [arsitektur](../../../agents/arsitektur.md), [boundary](../../../agents/boundary.md), [identitas dan akses](../../../agents/users.md), [cara coding gRPC](../../../agents/how_to_code_grpc_managements.md), [cara wiring Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md), [katalog pengajuan](../../../agents/grpc/hrm/33-alur-pengajuan-akun-employee.md), [instruksi IAM](../../../agents/instructions/iam/2-leader-remediasi-hubungan-iam-hrm-instructions.md) dan [instruksi HRM](../../../agents/instructions/hrm/9-leader-remediasi-hubungan-iam-hrm-instructions.md).
