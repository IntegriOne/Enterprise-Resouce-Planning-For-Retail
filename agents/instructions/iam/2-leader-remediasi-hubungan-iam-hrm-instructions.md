# Perbaikan siklus akun employee di IAM

- **ID** 2 · **Pemilik** Agent IAM · **Status** Siap dikerjakan.
- **Prioritas** Tinggi untuk IH-01, IH-02 dan IH-03. Sedang untuk IH-04.
- **Baseline** Checkout bersama `0fd7bb5`. Bukti dan batas pemeriksaan tersedia di [audit IAM–HRM](../../../agents/audit/iam/18-2026-11-28-205822-audit-kode-inti-hubungan-iam-hrm.md).

## Task

Jadikan service IAM penentu syarat pembuatan dan aktivasi akun employee. Selaraskan penerimaan dan klaim dengan pengajuan HR yang sah. Kerjakan sampai jalur sinkron, relay, dan pemulihan menggunakan fakta yang sama.

## Request

### IH-01 dan IH-02 — Pembuatan dan aktivasi

- file: [user_handler.go L38](../../../backend/services/iam/transport/grpc/handler/user_handler.go#L38) · [service.go L162](../../../backend/services/iam/services/user/service.go#L162) · [employee_provisioning.go L12](../../../backend/services/iam/services/user/employee_provisioning.go#L12) · [termination.go L34](../../../backend/services/iam/services/user/termination.go#L34) · [identity.go L536](../../../backend/services/iam/services/user/identity.go#L536) · [registration_runtime.go L353](../../../backend/services/iam/services/user/registration_runtime.go#L353) · [employee_repository.go L140](../../../backend/services/iam/repository/provisioning/employee_repository.go#L140) · [bootstrap.go L620](../../../backend/services/iam/services/bootstrap.go#L620).
- function: `CreateUser`, `claimEmployeeProvisioning`, `ApplyEmployeeTermination`, `transitionUserStatus`, `requireEmployeeLinkForActivation` serta semua transisi identitas yang dapat mengaktifkan employee.
- logic: Validasi worker dan receipt sebelum membuat employee reguler. Pertahankan klaim, asosiasi akun dan idempotensi dalam satu transaksi. Bedakan bootstrap yang dimiliki IAM secara eksplisit. Catat keputusan deaktif akibat terminasi secara persisten dalam transaksi deaktivasi atau terapkan verifikasi HRM setara yang tahan terhadap urutan event. Saat aktivasi, gunakan bukti hubungan aktif untuk principal bersangkutan. Tangani pegawai lama, hubungan paralel dan rehire yang sah.
- error_handling: Tolak referensi kosong/tidak valid sebelum write. Kembalikan konflik untuk receipt/worker yang berbeda atau aktivasi yang tidak memenuhi syarat. Jika fakta wajib dari HRM tidak tersedia, hentikan transisi dan pertahankan status sebelumnya.
- security: Pertahankan pengecekan sesi, izin, MFA pada jalur administratif terkait, dan status pending sampai tautan diverifikasi. Lindungi pengecualian bootstrap pada entry point tepercaya. Keputusan terminasi dan revokasi sesi harus tetap berlaku sesudah restart.
- expected_result: Satu pengajuan sah menghasilkan satu akun pending, retry mengembalikan akun yang sama, aktivasi akun pascaterminasi memerlukan hubungan aktif yang sah, dan jenis akun lain tetap mengikuti alurnya.
- testing: Verifikasi setelah perbaikan pada lapisan service IAM untuk pembuatan reguler, bootstrap baru/lama, aktivasi manual dan otomatis, akun linked yang sudah terminasi, rehire, dua hubungan kerja, restart dan outage HRM. Sertakan percobaan serentak terminasi dan aktivasi dengan hasil akhir sesuai keputusan HRM terkini.

### IH-03 dan IH-04 — Keabsahan pengajuan dan kesetaraan replay

- file: [employee_provisioning.proto L10](../../../backend/services/iam/contract/proto/iam/v1/employee_provisioning.proto#L10) · [employee_provisioning_handler.go L34](../../../backend/services/iam/transport/grpc/handler/employee_provisioning_handler.go#L34) · [employee.go L92](../../../backend/services/iam/services/provisioning/employee.go#L92) · [employee_provisioning.go L36](../../../backend/services/iam/services/user/employee_provisioning.go#L36) · [requisition_mandate.go L22](../../../backend/services/iam/cmd/requisition_mandate.go#L22) · [employee_repository.go L32](../../../backend/services/iam/repository/provisioning/employee_repository.go#L32) · [000001_create_iam_schema.up.sql L355](../../../backend/services/iam/db/migrations/000001_create_iam_schema.up.sql#L355).
- function: `Receive`, `ReceiveFromHRMWorkload`, `InsertReceipt`, klaim dalam `CreateUser`, dan `Complete`.
- logic: Sepakati bersama HRM kapan pengajuan bersifat tetap dan status apa yang masih dapat diklaim. Bila penarikan didukung, tentukan pemenang persaingan klaim/penarikan secara eksplisit. Terima fakta identik pada jalur Gateway dan workload. Ikat ID pengajuan, event lifecycle, versi yang didukung dan digest ke receipt. Bandingkan pada replay. Untuk record historis tanpa bukti, cocokkan ke pengajuan kanonik HRM sebelum melengkapinya. Buat perubahan database yang aman untuk instalasi yang sudah memiliki receipt.
- error_handling: Pisahkan fakta berbeda/pengajuan tidak berlaku dari dependency sementara tidak tersedia. Pastikan pengiriman terlambat tidak menghidupkan kembali pengajuan terminal. Jangan mengisi bukti historis dengan nilai rekaan.
- security: Verifikasi identitas workload dan scope pemohon pada jalurnya. Digest dipakai untuk integritas isi dan pembandingan. Otorisasi tetap bersumber dari IAM dan fakta HRM.
- expected_result: Pengiriman sinkron dan relay dalam urutan mana pun hanya menghasilkan satu receipt. Perbedaan event atau versi terdeteksi. Setelah pengajuan sah ditarik, klaim baru gagal sesuai protokol yang disepakati.
- testing: Pada service, periksa urutan sinkron/relay terbalik, replay identik/berbeda, versi tak dikenal, pemulihan receipt lama, penarikan berbarengan dengan klaim serta outage di setiap batas commit. Jalankan E2E bersama kedua pemilik.

## Action

1. Sepakati perubahan proto dan kompatibilitas rollout dengan HRM dan Gateway. Dokumentasikan perlakuan data lama dan aktifkan aturan baru setelah pembaca/pengirim kompatibel.
2. Implementasikan kontrol service dan penyimpanan IAM. Pisahkan mekanisme takeover operator jika kelak mendapat keputusan owner IAM. Jaga pemrosesan oleh pengklaim yang berwenang selama ini.
3. Setelah perubahan, jalankan verifikasi unit di paket service, verifikasi integrasi nyata melalui HTTPS dan catat hasil per temuan serta SHA. Kelola grant untuk E2E melalui GraphQL IAM.

## Context

Baca [AGENTS.md](../../../AGENTS.md), [arsitektur](../../../agents/arsitektur.md), [boundary](../../../agents/boundary.md), [identitas dan akses](../../../agents/users.md), [cara coding gRPC](../../../agents/how_to_code_grpc_managements.md), [cara wiring Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md), [katalog pengajuan](../../../agents/grpc/hrm/33-alur-pengajuan-akun-employee.md), [katalog terminasi](../../../agents/grpc/hrm/40-deaktivasi-akun-setelah-terminasi.md), [instruksi HRM](../../../agents/instructions/hrm/9-leader-remediasi-hubungan-iam-hrm-instructions.md) dan [instruksi Gateway](../../../agents/instructions/gateway/4-leader-remediasi-hubungan-iam-hrm-instructions.md).
