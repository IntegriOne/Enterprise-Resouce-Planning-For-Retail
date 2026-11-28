# Instruksi Human Resource untuk remediasi Procurement, HRM dan Finance

- **id** — 10
- **owner** — Agent Human Resource, reviewer leader lintas management.
- **priority** — Tinggi. Dahulukan kerahasiaan payroll, integritas pembayaran dan validasi owner.
- **status** — Siap dikerjakan.
- **baseline** — Working tree `main` berbasis `da64877287ad`, audit kode produksi 28 November 2026 WIB.

## Task

Pulihkan dua transaksi anggaran, sediakan fakta pekerja minimum, dan tutup alur bukti slip per hasil. Gunakan [laporan PHF-01 sampai PHF-08](../../../agents/audit/finance_accounting/9-2026-11-28-213747-audit-kode-inti-procurement-hrm-finance.md) untuk bukti dan [matriks operasi](../../../agents/grpc/finance_accounting/6-matriks-operasi-procurement-hrm-finance.md) untuk daftar field yang harus ditelusuri.

## Request

### PHF-06 · Pemulihan komitmen anggaran lintas transaksi

- file: [backend/services/human_resource/internal/service/recruitment/requisition_opening.go](../../../backend/services/human_resource/internal/service/recruitment/requisition_opening.go#L17) · [backend/services/human_resource/internal/service/workforceplanning/plan_approval.go](../../../backend/services/human_resource/internal/service/workforceplanning/plan_approval.go#L123) · [backend/services/human_resource/internal/service/operation/command.go](../../../backend/services/human_resource/internal/service/operation/command.go#L162) · [backend/services/human_resource/infrastructure/integration/finance/budget_verifier.go](../../../backend/services/human_resource/infrastructure/integration/finance/budget_verifier.go#L94) · [backend/services/finance_accounting/repository/budget/hr_authorization.go](../../../backend/services/finance_accounting/repository/budget/hr_authorization.go#L178)
- function: OpenRequisition, ApproveWorkforcePlan, CommitRequisitionBudget, CommitWorkforceBudget
- logic: Buat intent dan outbox di transaksi HRM sebelum memanggil Finance. Simpan identitas intent, versi subject, peninjau, scope, digest dan authorization. Worker membedakan belum dikirim, hasil ambigu dan selesai. Finalisasi lokal dengan guard versi serta cek ulang mandat. Rekonsiliasi atau kompensasi dengan Finance setelah menentukan status kedua sisi. Ganti fallback verifier-only dengan dependency commit wajib ketika workflow memerlukan komitmen.
- error_handling: Commit Finance sukses lalu HRM gagal, respons hilang, crash dan delivery duplikat harus menuju satu hasil terminal. Tolak digest/mandat tak cocok; anggaran dilepas hanya bila intent sah tidak dipakai.
- security: Reviewer wajib mempunyai grant IAM spesifik unit dan persetujuan anggaran Finance yang sah. Worker mTLS dibatasi per method; retry mempertahankan identitas keputusan.
- expected_result: Requisition open/plan approved memiliki komitmen Finance yang sesuai. Intent tersendat dapat dipulihkan tanpa memulai keputusan pengguna baru.
- testing: Simulasikan gagal commit lokal, Finance timeout sebelum/sesudah commit, dua approval bersamaan, revoke mandat sebelum finalisasi dan retry.

### PHF-02 · Fakta penerima klaim untuk Finance

- file: [backend/services/finance_accounting/domain/service/ports.go](../../../backend/services/finance_accounting/domain/service/ports.go#L127) · [backend/services/finance_accounting/internal/service/expenses/service.go](../../../backend/services/finance_accounting/internal/service/expenses/service.go#L93) · [backend/services/human_resource/internal/service/personnel/principal_link.go](../../../backend/services/human_resource/internal/service/personnel/principal_link.go#L1)
- function: Verifikasi privat scope pekerja dan relasinya ke principal
- logic: Pakai sumber worker dan tautan IAM yang sudah dipakai domain personel. Berikan fakta minimum untuk Finance mengecek legal entity, worker, serta apakah principal adalah penerima pada tanggal yang disepakati. Kunci definisi perilaku jika tautan berakhir, belum ada atau pekerja berpindah entity. Dokumentasikan pemilik dan versi proto.
- error_handling: Data ambigu, tidak berwenang dan integrasi tidak tersedia memberi status terpisah sesuai kebijakan anti-enumerasi. Finance menahan approval sampai fakta wajib tersedia.
- security: Audience mTLS Finance dan izin workload khusus. Dilarang menerima identitas reviewer melalui field GraphQL bebas atau membuka data profil penuh.
- expected_result: Finance dapat membedakan ID akun dan ID worker. Reimbursement oleh petugas atas nama orang lain mempertahankan SoD.
- testing: Pekerja dengan UUID berbeda dari principal, relasi berakhir, perubahan penempatan, salah entity, caller tanpa grant.

### PHF-07/08 dan PHF-01 · Slip setelah payroll terbit dan bukti pembayaran terpisah

- file: [backend/services/human_resource/internal/service/payroll/payslip_publication.go](../../../backend/services/human_resource/internal/service/payroll/payslip_publication.go#L241) · [backend/services/human_resource/infrastructure/inbox/payroll_publication.go](../../../backend/services/human_resource/infrastructure/inbox/payroll_publication.go#L25) · [backend/services/human_resource/internal/service/payroll/payslip_reissue.go](../../../backend/services/human_resource/internal/service/payroll/payslip_reissue.go#L27) · [backend/gateway/graphql/human_resource/hrm_payslip.graphql](../../../backend/gateway/graphql/human_resource/hrm_payslip.graphql#L31) · [agents/grpc/hrm/41-penerimaan-event-payroll-finance.md](../../../agents/grpc/hrm/41-penerimaan-event-payroll-finance.md#L1) · [agents/grpc/hrm/44-snapshot-payroll-untuk-finance.md](../../../agents/grpc/hrm/44-snapshot-payroll-untuk-finance.md#L1)
- function: HandlePayrollPublication, PublishPayslipsForRun, ReissueEmployeePayslip dan konfirmasi result_paid
- logic: Verifikasi publikasi run dari Finance. Simpan versi PDF privat yang terikat pada publication dan result secara idempoten di DRM. Jika Finance membutuhkan receipt per item, kembalikan digest terverifikasi dan tulis outbox di transaksi yang mengikat dokumen ke hasil. Sepakati konsumen privat yang mengakui receipt. Pertahankan konfirmasi paid dari alokasi Finance sebagai event terpisah. Untuk mantan karyawan, petugas memilih worker dan hasil/periode; apabila DRM memang kehilangan versi, pulihkan berdasarkan publikasi asli dan catat audit.
- error_handling: DRM unavailable dapat diulang, missing terverifikasi mengikuti prosedur restore, ketidakcocokan owner/digest gagal tertutup. Retry tidak menerbitkan PDF berbeda atau mengubah status paid.
- security: myPayslip/pdf/list membutuhkan pegawai aktif dan scope SELF. Reissue memerlukan grant petugas HR scoped dan MFA serta validasi worker–result. Lindungi PDF, identifier payroll dan tautan dokumen.
- expected_result: Pegawai aktif bisa cetak sendiri setelah publikasi; mantan karyawan dilayani HR. Slip terbit sebelum atau sesudah pembayaran sesuai fakta Finance tanpa menandai lunas sendiri.
- testing: Uji urutan publication/paid, event duplikat, PDF hilang vs timeout, wrong worker/result, former self ditolak dan petugas berwenang dapat mengembalikan PDF.

## Action

1. Cocokkan setiap simbol dan status wiring pada checkout terbaru. Tautkan temuan ke perubahan yang sudah masuk agar pekerjaan agent lain dapat dipakai. Kerjakan sampai seluruh kriteria penerimaan dalam scope terpenuhi.
2. Mulai dari invariant DBML/migrasi dan kontrak event/proto, kemudian repository, transaksi service, handler/policy dan wiring startup. Jalankan generator resmi sebelum wiring SDL → resolver → adapter → client. Untuk perubahan lintas owner, sepakati bentuk payload, scope, versi, replay, serta urutan rollout consumer/producer.
3. Pertahankan kompatibilitas data lama dan pemanggil AP, payroll dan provisioning. Perbarui inventaris di `agents/grpc/{management}/` (untuk HRM gunakan `agents/grpc/hrm/`), beserta penjelasan operasi, Query/Mutation/privat, RBAC, validasi, state, error dan sumber fakta.
4. Untuk perubahan Go, lakukan fmt, vet, build dan pemeriksaan service yang relevan. Unit test server gRPC ditempatkan pada lapisan service sesuai AGENTS.md. Gunakan HTTPS `/api/graphql` untuk verifikasi publik, login superadmin dan atur role, permission, scope serta assignment melalui GraphQL IAM. Gunakan principal pekerja/petugas yang sesuai untuk skenario SELF/SoD. Pengujian ini adalah kriteria penyerahan implementasi berikutnya.
5. Lampirkan per ID temuan lokasi file/fungsi, operasi yang dipublikasikan, bukti hasil positif/penolakan, migrasi data lama dan pemulihan dari retry. Catat hambatan owner spesifik bila ada, perbarui katalog dan serahkan ke leader untuk audit ulang.

## Context

[Aturan repo](../../../AGENTS.md) · [Arsitektur](../../../agents/arsitektur.md) · [Boundary](../../../agents/boundary.md) · [Identitas dan akses](../../../agents/users.md) · [Konteks Human Resource](../../../agents/context/human_resource/1-context.md) · [how_to_code gRPC](../../../agents/how_to_code_grpc_managements.md) · [how_to_code Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md) · [Template instruksi](../../../agents/instruction_template.md)

[agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md](../../../agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md#L1) · [agents/grpc/finance_accounting/4-matriks-query-mutation-dan-publikasi-gateway.md](../../../agents/grpc/finance_accounting/4-matriks-query-mutation-dan-publikasi-gateway.md#L1) · [agents/grpc/hrm/1-inventaris-operasi-grpc-hrm.md](../../../agents/grpc/hrm/1-inventaris-operasi-grpc-hrm.md#L30) · [agents/logic/procurement/logic.md](../../../agents/logic/procurement/logic.md#L945) · [agents/grpc/procurement_sourcing/2-kontrak-integrasi-finance.md](../../../agents/grpc/procurement_sourcing/2-kontrak-integrasi-finance.md#L1) · [agents/grpc/hrm/41-penerimaan-event-payroll-finance.md](../../../agents/grpc/hrm/41-penerimaan-event-payroll-finance.md#L1) · [agents/grpc/finance_accounting/3-integrasi-inbox-outbox.md](../../../agents/grpc/finance_accounting/3-integrasi-inbox-outbox.md#L1)

[Instruksi Finance & Accounting](../../../agents/instructions/finance_accounting/8-leader-remediasi-procurement-hrm-finance-instructions.md) · [Instruksi Procurement & Sourcing](../../../agents/instructions/procurement_sourcing/6-leader-remediasi-procurement-hrm-finance-instructions.md) · [Instruksi GraphQL Gateway](../../../agents/instructions/gateway/6-leader-remediasi-procurement-hrm-finance-instructions.md)
