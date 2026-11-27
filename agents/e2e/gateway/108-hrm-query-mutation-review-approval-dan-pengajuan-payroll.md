# Uji review, persetujuan, dan pengajuan jurnal payroll lewat HTTPS

Dua akun fixture dengan TOTP dan role IAM resmi memakai stack disposable. Run draf dibuat melalui GraphQL. Grup, periode, aturan, profil pekerja, komponen, hasil, dan status kalkulasi disiapkan melalui SQL pada database `*_test` agar alur setelah kalkulasi dapat diuji tanpa kalkulator produksi. Verifikasi identitas employee kedua juga memakai fixture SQL pada IAM disposable.

| Kasus | Hasil |
| --- | --- |
| `hrmSubmitPayrollRunForReview` oleh pembuat | Status `review` |
| `hrmApprovePayrollRun` oleh pembuat | `FORBIDDEN` |
| Persetujuan oleh employee kedua | Status `approved` |
| `hrmRequestPayrollPosting` setelah persetujuan | Berhasil, status HRM tetap `approved` |
| Pengajuan posting kedua | `CONFLICT` |
| `hrmPayrollResults` dan `hrmPayrollResultLines` dengan MFA | Masing-masing mengembalikan satu baris dengan ID yang sesuai |
| Pemeriksaan PostgreSQL | Approver berbeda dari pembuat, waktu pengajuan ada, referensi jurnal Finance masih kosong, dan tepat satu event `hr_payroll_runs.request_posting` tersimpan |

Database dan Redis uji dihapus setelah runtime development kembali sehat. Hasil ini membuktikan pengajuan sampai outbox HRM. Konfirmasi posting, penerbitan payroll, dan transfer tetap tanggung jawab Finance. [Kontrak handoff](../../grpc/hrm/25-kebutuhan-integrasi-finance-payroll.md).
