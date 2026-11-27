# Uji pembatalan payroll saat review melalui HTTPS

Run dibuat melalui Gateway. Hasil kalkulasi, referensi payroll, dan verifikasi identitas employee kedua menggunakan fixture SQL pada database disposable. Pembuat dan employee kedua menerima role melalui IAM dan login dengan TOTP.

| Kasus | Hasil |
| --- | --- |
| Pembuat mengajukan run terkalkulasi untuk review | Status `review` |
| Pembuat mencoba membatalkan sendiri | `FORBIDDEN` |
| Employee kedua menjalankan `hrmCancelPayrollRun` dengan alasan | Status `cancelled`, waktu pembatalan terisi |
| Ulangi pembatalan | `CONFLICT` |
| Coba ajukan jurnal setelah pembatalan | `CONFLICT` |
| Pemeriksaan PostgreSQL | Pelaku pembatalan berbeda dari pembuat, alasan tersimpan, waktu pengajuan kosong, satu event pembatalan dan nol event pengajuan jurnal |

Database dan Redis uji dibersihkan setelah runtime development kembali sehat. [Uji payroll jalur persetujuan](108-hrm-query-mutation-review-approval-dan-pengajuan-payroll.md).
