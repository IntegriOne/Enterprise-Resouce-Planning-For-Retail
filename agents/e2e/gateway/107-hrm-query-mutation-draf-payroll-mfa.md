# Uji draf payroll dan prasyarat pengajuan melalui HTTPS

Superadmin fixture memakai sesi TOTP dan role IAM khusus payroll. Badan hukum aktif dibuat lewat API Enterprise Foundation. Grup, periode, dan aturan payroll dimuat sebagai baris sintetis melalui fixture SQL yang dibatasi ke database `*_test`. Referensi artefak dan snapshot hanya dipakai untuk uji draf.

| Kasus | Hasil |
| --- | --- |
| `hrmCreatePayrollRun` dengan referensi fixture | Berhasil, status `draft` |
| Pengulangan input dengan idempotency key yang sama | `CONFLICT` |
| `hrmCalculatePayrollRun` tanpa kalkulator runtime | `UNAVAILABLE` |
| `hrmSubmitPayrollRunForReview` sebelum kalkulasi | `CONFLICT` |
| `hrmRequestPayrollPosting` sebelum persetujuan | `CONFLICT` |
| `hrmPayrollRun` dan pemeriksaan PostgreSQL | Tetap `draft`, waktu kalkulasi dan pengajuan posting kosong |

Database dan Redis uji dibersihkan setelah layanan development kembali sehat. Jalur kalkulasi positif memerlukan kalkulator runtime. Finance tetap memiliki penerbitan payroll dan pembayaran. [Handoff kalkulator](../../grpc/hrm/27-kebutuhan-kalkulator-payroll-runtime.md) · [Handoff Finance](../../grpc/hrm/25-kebutuhan-integrasi-finance-payroll.md).
