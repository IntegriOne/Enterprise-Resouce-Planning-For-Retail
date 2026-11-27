# Uji terminasi pekerja melalui HTTPS dengan persetujuan independen

Superadmin fixture dan employee kedua memakai sesi TOTP pada stack disposable. Status awal pekerja aktif dan verifikasi identitas employee kedua disiapkan melalui SQL fixture di database `*_test`. Pembuatan rencana, persetujuan independen, mulai offboarding, dan terminasi memakai API resmi.

| Kasus | Hasil |
| --- | --- |
| `hrmTerminateWorker` dengan rencana berjalan dan tanggal efektif hari pengujian dalam UTC | Pekerja berstatus `terminated` |
| Ulangi mutation untuk pekerja dan rencana yang sama | `CONFLICT` |
| Detail pekerja melalui HTTPS | Status `terminated` |
| Pemeriksaan PostgreSQL | Tanggal terminasi `2026-11-26` dan alasan `resignation` cocok dengan rencana. Rencana tetap `in_progress`. |

Database dan Redis uji dibersihkan setelah layanan development kembali sehat. Penutupan rencana memerlukan prasyarat final payroll dan clearance tersendiri. [Uji mulai offboarding](105-hrm-mutation-mulai-offboarding-dan-pengecualian-tugas.md).
