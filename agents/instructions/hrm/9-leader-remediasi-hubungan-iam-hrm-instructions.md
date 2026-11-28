# Perbaikan pengajuan akun dan bukti hubungan kerja di HRM

- **ID** 9 · **Pemilik** Agent HRM · **Status** Siap dikerjakan.
- **Prioritas** Tinggi untuk IH-02 dan IH-03. Sedang untuk IH-04 dan IH-05.
- **Baseline** Checkout bersama `0fd7bb5`. Lihat [audit IAM–HRM](../../../agents/audit/iam/18-2026-11-28-205822-audit-kode-inti-hubungan-iam-hrm.md).

## Task

Pastikan pengajuan yang dikirim ke IAM memiliki status dan bukti yang dapat dipertanggungjawabkan HRM. Izinkan petugas HR menjalankan alur ini untuk worker di badan hukum yang menjadi tanggung jawabnya.

## Request

### IH-03 dan IH-04 — Masa berlaku pengajuan dan bukti pengiriman

- file: [service.go L71](../../../backend/services/human_resource/internal/service/accessrequest/service.go#L71) · [mapper.go L123](../../../backend/services/human_resource/internal/service/accessrequest/mapper.go#L123) · [access_request.go L47](../../../backend/services/human_resource/infrastructure/security/access_request.go#L47) · [resource.go L413](../../../backend/services/human_resource/internal/service/operation/resource.go#L413) · [employee_provisioning.go L44](../../../backend/services/human_resource/infrastructure/integration/iam/employee_provisioning.go#L44) · [employee_provisioning.proto L24](../../../backend/services/iam/contract/proto/iam/v1/employee_provisioning.proto#L24) · [hrm_accessrequest_entities.graphql L85](../../../backend/gateway/graphql/human_resource/hrm_accessrequest_entities.graphql#L85).
- function: `CreateAccessRequest`, `UpdateAccessRequest`, `DeleteAccessRequest`, `ValidateMutation`, pembentukan event dalam `specAccessRequest` dan `EmployeeProvisioningPublisher.Publish`.
- logic: Bekukan fakta pengajuan provisioning sejak pengajuan dan outbox berhasil di-commit. Cegah CRUD generik menghapus atau mengubah pengajuan yang sudah diserahkan. Jika bisnis mendukung penarikan, simpan keputusan terminal yang dapat dibaca IAM, tentukan batas klaim, dan kirim perubahan secara durable. Sediakan fakta kanonik yang sama untuk pengiriman sinkron dan asinkron, termasuk pengajuan, worker, badan hukum, katalog, pemohon, waktu dan event. Untuk data lama, identifikasi pengajuan hilang atau berbeda dari receipt dan rekonsiliasi bersama IAM.
- error_handling: Konflik state tidak boleh memicu pengiriman pengajuan baru. Pertahankan retry dengan ID yang sama ketika IAM sementara tidak tersedia. Perlakukan perubahan fakta sebagai konflik terminal yang perlu direkonsiliasi.
- security: Hanya workload IAM beridentitas sah yang boleh membaca bukti privat lintas service. Tetap ikat actor dan scope pada record HRM, dan simpan jejak keputusan tanpa data identitas berlebihan.
- expected_result: Pengajuan yang sudah diserahkan dapat ditelusuri. Pengiriman ulang peristiwa lama tidak mengubah keputusan terminal dan tidak menghasilkan akun dari pengajuan yang sudah tak berlaku.
- testing: Setelah implementasi, verifikasi di service HRM percobaan update/delete saat pending delivery, sesudah IAM menerima, serta setelah klaim. Periksa juga dua pengiriman serentak dan pemulihan ketika respons owner hilang.

### IH-02 dan IH-05 — Bukti hubungan kerja dan cakupan petugas HR

- file: [termination_access.go L20](../../../backend/services/human_resource/internal/service/personnel/termination_access.go#L20) · [handler.go L40](../../../backend/services/human_resource/transport/grpc/handler/terminationaccess/handler.go#L40) · [termination_access.proto L9](../../../backend/services/human_resource/contract/proto/human_resource/workload/termination/v1/termination_access.proto#L9) · [access_request.go L106](../../../backend/services/human_resource/infrastructure/security/access_request.go#L106) · [policy.go L124](../../../backend/services/human_resource/infrastructure/security/policy.go#L124) · [mapper.go L103](../../../backend/services/human_resource/internal/service/accessrequest/mapper.go#L103) · [principal_link.go L25](../../../backend/services/human_resource/internal/service/personnel/principal_link.go#L25).
- function: `VerifyTerminationAccess`, `VerifyWorkerAccountLink`, `Policy.Authorize`, `Policy.ReadScope` dan service pembacaan/pengajuan access request.
- logic: Sediakan fakta yang dibutuhkan IAM untuk membedakan hubungan berakhir, hubungan lain yang masih aktif dan rehire. Untuk akses petugas HR, tentukan badan hukum dari worker milik HRM, bandingkan dengan grant, periksa kembali pada write. Terapkan filter pada lookup pengajuan, finder untuk retry dan penghitungan sebelum pagination. Pastikan `requested_scope_reference_id` dengan tipe scope lain tidak dianggap badan hukum. Pertahankan jalur cetak ulang mantan karyawan melalui petugas HR.
- error_handling: Hindari mengungkap keberadaan worker di luar scope. Bedakan data bertentangan dari ketidaktersediaan owner agar proses pendataan dan aktivasi bisa dipulihkan.
- security: Terapkan least privilege di service HRM dan pertahankan aturan MFA untuk pembacaan sensitif. Gunakan kontrak privat dengan IAM untuk bukti relasi. Penautan worker tetap dilakukan otomatis melalui alur IAM.
- expected_result: Petugas di badan hukum A dapat mengajukan dan memantau pengajuan A. Worker, status, dan jumlah milik B tidak terungkap. IAM mendapat keputusan yang cocok untuk worker aktif, mantan karyawan dan kasus hubungan paralel.
- testing: Verifikasi seusai perubahan pada service dengan dua badan hukum, beberapa grant yang sah, status tanpa grant, grant yang dicabut, worker terminated/rehire, serta hubungan kedua yang aktif. Jalankan pembuktian Gateway bersama pemiliknya.

## Action

1. Sepakati titik pengajuan menjadi tetap, protokol penarikan bila diperlukan, dan kompatibilitas bukti dengan IAM.
2. Perbaiki transaksi, validator, relay serta policy owner. Perbarui inventaris [pengajuan akun](../../../agents/grpc/hrm/33-alur-pengajuan-akun-employee.md) dan [terminasi](../../../agents/grpc/hrm/40-deaktivasi-akun-setelah-terminasi.md) untuk perubahan kontrak.
3. Lakukan verifikasi di service, kemudian HTTPS melalui Gateway. Siapkan role dan permission melalui GraphQL IAM dan catat hasil aktual serta SHA.

## Context

Baca [AGENTS.md](../../../AGENTS.md), [arsitektur](../../../agents/arsitektur.md), [boundary](../../../agents/boundary.md), [identitas dan akses](../../../agents/users.md), [cara coding gRPC](../../../agents/how_to_code_grpc_managements.md), [cara wiring Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md), [aturan slip mantan karyawan](../../../agents/grpc/hrm/39-cetak-ulang-slip-gaji-mantan-karyawan.md), [instruksi IAM](../../../agents/instructions/iam/2-leader-remediasi-hubungan-iam-hrm-instructions.md) dan [instruksi Gateway](../../../agents/instructions/gateway/4-leader-remediasi-hubungan-iam-hrm-instructions.md).
