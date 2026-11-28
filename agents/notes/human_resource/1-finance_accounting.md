# Finance Accounting – human_resource – Bukti publikasi payroll dan slip

- **Role** — Agent backend human_resource, koordinasi dengan owner Finance
- **Priority** — P0
- **Status** — SIAP DIKERJAKAN
- **Repository dan baseline** — ERP Retail, `main` berbasis `bba5f46cfec19129dce58b8a70d813e19f4ce3a1`, checkout 29 November 2026
- **Target management** — human_resource

## Task

HRM menerima run dengan no-journal posting, lalu pada salinan yang ditinjau membaca pointer journal tanpa syarat saat verifikasi publikasi. Finance menerbitkan payroll.published dengan journal_entry_id saja. Verifier publikasi dan lookup tidak mengikat bukti no-journal seperti verifier posting/payment. Risiko mencakup kegagalan penerbitan slip payroll yang sah. Slip dibuat setelah publikasi, termasuk run tanpa jurnal. Kehilangan PDF dapat dipulihkan dari edisi terverifikasi. Status nol tidak menampilkan transfer uang. Owner Finance mempertahankan kepemilikan jurnal dan uang. [Instruksi Finance](../../../agents/instructions/finance_accounting/9-leader-remediasi-logika-keuangan-dan-integrasi-instructions.md) · [Bukti audit](../../../agents/audit/finance_accounting/10-2026-11-29-010010-audit-logika-keuangan-dan-integrasi.md)

## Request

### 1. Files

| Aksi | File aktual atau tujuan |
| --- | --- |
| READ / Finance | `backend/services/finance_accounting/internal/service/payroll/service.go:607` |
| READ / Finance | `backend/services/finance_accounting/internal/service/payroll/service.go:891` |
| READ / Finance | `backend/services/finance_accounting/internal/service/payroll/service.go:1062` |
| READ / Finance | `backend/services/finance_accounting/transport/grpc/handler/payroll_handler.go:311` |
| READ / Finance | `backend/services/finance_accounting/internal/service/treasury/payroll_allocation.go` |
| READ / Finance | `backend/services/finance_accounting/contract/proto/finance_accounting/v1/payroll.proto` |
| MODIFY | [backend/services/human_resource/internal/service/payroll/payslip_publication.go:149](../../../backend/services/human_resource/internal/service/payroll/payslip_publication.go#L149) |
| MODIFY | [backend/services/human_resource/internal/service/payroll/payslip_reissue.go:79](../../../backend/services/human_resource/internal/service/payroll/payslip_reissue.go#L79) |
| MODIFY | `backend/services/human_resource/infrastructure/integration/finance/payroll_verifier.go` |

### 2. Functions / Methods / Operations

- file: Gunakan titik masuk pada tabel di atas. Unit test berada pada layer service yang menangani operasi.
- function: PublishPayroll, VerifyPayrollPublication, FindPayrollPublicationByRun, PublishPayslipForResult, ReissueEmployeePayslip
- logic: Gunakan satu dari journal_id atau no_journal_posting_id di publication, lookup dan event. Cocokkan bukti, entity, run dan digest terhadap posting Finance. Untuk settlement, bedakan no_transfer_required, partially_paid, paid, reversed, termasuk outstanding per worker. Pertahankan edisi slip immutable dan reconciliasi DRM.
- error_handling: Tolak pasangan kosong/ganda, mismatch dan event versi lama tanpa bukti terverifikasi. Retry DRM membuat satu versi per publikasi/hasil. Publikasi Finance dapat selesai sebelum cash settlement.
- security: HR mengajukan rincian. Finance memublikasikan dan mencairkan. Employee aktif dapat mencetak sendiri, mantan karyawan melalui HR berwenang dengan bukti audit. Batasi data sensitif per recipient.
- expected_result: Slip dibuat setelah publikasi, termasuk run tanpa jurnal. Kehilangan PDF dapat dipulihkan dari edisi terverifikasi. Status nol tidak menampilkan transfer uang.
- testing: Bukti journal dan no-journal menghasilkan satu PDF per edisi · Digest/owner salah ditolak · Timeout DRM dan retry · Hasil positif, nol, parsial dan reversal menunjukkan status benar · Mantan employee meminta HR, cetak ulang tercatat

### 3. Integration Requirements

Finance menerbitkan fakta publikasi, HRM memverifikasi run dan worker, mengajukan penyimpanan atau pemulihan edisi PDF ke DRM dan menyajikannya sesuai aktor. Bandingkan kontrak publikasi dengan verifier payment yang sudah membawa no_journal_posting_id. Pakai gRPC privat beridentitas workload dan inbox/outbox idempoten. Pertahankan ID, versi dan digest untuk menelusuri respons yang sah saat recipient pulih dari timeout.

## Action

1. Cocokkan versi kontrak dan adapter aktual kedua owner, termasuk data lama.
2. Perbarui service dan adapter milik human_resource bersama penerbit kontrak Finance.
3. Uji service dan panggil jalur pengguna melalui HTTPS GraphQL. Kelola grant uji melalui GraphQL IAM.
4. Laporkan SHA, bukti kasus di bawah, alur retry dan hasil rekonsiliasi ke owner Finance.

| Kasus | Skenario yang harus dibuktikan |
| --- | --- |
| TC-01 | Bukti journal dan no-journal menghasilkan satu PDF per edisi |
| TC-02 | Digest/owner salah ditolak |
| TC-03 | Timeout DRM dan retry |
| TC-04 | Hasil positif, nol, parsial dan reversal menunjukkan status benar |
| TC-05 | Mantan employee meminta HR, cetak ulang tercatat |

## Context

[AGENTS.md](../../../AGENTS.md) · [agents/arsitektur.md](../../../agents/arsitektur.md) · [agents/boundary.md](../../../agents/boundary.md) · [agents/users.md](../../../agents/users.md). [Matriks API nyata dan usulan kontrak](../../../agents/grpc/finance_accounting/7-matriks-operasi-publik-privat-dan-usulan-keuangan.md) · [Panduan gRPC](../../../agents/how_to_code_grpc_managements.md) · [Panduan Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md).
