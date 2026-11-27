# List operasi

| Operasi | Kebutuhan Finance |
| --- | --- |
| `CompleteOffboardingPlan` | Verifikasi clearance Finance untuk pekerja dan hasil payroll terminasi yang tepat. |

## Penjelasan CompleteOffboardingPlan

HRM memeriksa rencana yang masih berjalan, terminasi pekerja, tugas wajib, dan hasil payroll terminasi sebelum meminta verifikasi. Finance perlu membuktikan bahwa referensi clearance yang diajukan terikat pada rencana offboarding, pekerja, tanggal terminasi, badan hukum, mata uang, run payroll final, hasil individual, dan digest kedua snapshot. Jika HRM sudah menyimpan referensi pembayaran, Finance perlu mencocokkan referensi itu dengan hasil individual.

Kontrak Finance perlu menetapkan kapan clearance boleh diberikan ketika payroll sudah disetujui HRM, sudah diterbitkan Finance, atau sudah dibayar. Ketiganya merupakan status yang dapat diperiksa oleh kode HRM hari ini. Bukti clearance harus membedakan status pembayaran individual dan status agregat run. Penerbitan, transfer dan rekonsiliasi tetap berada di Finance Accounting.

## Butuh GraphQL Mutation

Gateway sudah menyediakan `completeOffboardingPlan(id: ID!, final_payroll_run_id: ID!, finance_clearance_reference_id: ID!)`. Verifikasi menggunakan kontrak privat antarlayanan. UUID clearance dikirim sebagai referensi ke keputusan Finance yang sudah ada dan harus diverifikasi oleh service pemilik.

## RBAC

Operator HRM memerlukan `human_resource.offboarding_plans.complete` dan MFA. Finance mengautentikasi workload HRM untuk membaca bukti. IAM memverifikasi clearance akses untuk principal pekerja secara terpisah. [Handoff IAM dan workflow](13-operasi-offboarding.md).

## Verifikasi integrasi

- Uji clearance sah yang cocok dengan seluruh snapshot dan penolakan saat pekerja, badan hukum, hasil payroll, digest atau referensi pembayaran berbeda.
- Uji clearance yang belum terbit, dicabut, dan gangguan Finance. Kegagalan verifikasi harus mempertahankan status rencana serta outbox.
- Konfirmasi status pekerja yang pembayaran individualnya sudah berhasil ketika run masih berstatus `posted`.
- Sinkronkan aturan clearance Finance dan bukti pencabutan akses IAM sebelum uji positif lintas layanan. [Implementasi HRM](../../../backend/services/human_resource/internal/service/offboarding/complete_plan.go).

Adapter `NewOwnerClearance` di HRM mewajibkan verifier IAM dan Finance. Setelah IAM menyetujui konteks akses pekerja, HRM meneruskan referensi dan digest payroll ke verifier Finance. Principal IAM tidak diteruskan ke Finance. Nominal, digest, dan referensi payroll tidak diteruskan ke IAM. Adapter ini diuji melalui gRPC dan PostgreSQL disposable dengan verifier fixture. Integrasi runtime masih memerlukan verifier dari pemilik masing-masing. [Adapter clearance](../../../backend/services/human_resource/internal/service/offboarding/owner_clearance.go)
