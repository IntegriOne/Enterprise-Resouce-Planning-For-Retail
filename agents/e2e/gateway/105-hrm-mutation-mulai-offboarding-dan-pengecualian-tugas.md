# Uji mulai offboarding dan pengecualian tugas melalui HTTPS

Kedua employee fixture memakai MFA TOTP serta grant IAM resmi pada stack disposable. Seperti uji persetujuan sebelumnya, pekerja aktif dan verifikasi identitas employee kedua disiapkan dengan fixture SQL khusus database `*_test`. Pembuat draf dan employee yang menyetujui rencana berbeda.

| Kasus | Hasil |
| --- | --- |
| Employee kedua memanggil `hrmStartOffboardingPlan` setelah persetujuan | Status `in_progress` |
| `hrmCancelOffboardingPlan` pada rencana yang sudah berjalan | `CONFLICT` |
| `hrmWaiveOffboardingTask` dengan alasan untuk tugas HRM yang pending | Status `waived` |
| Query detail tugas dan pemeriksaan PostgreSQL | Tugas tetap `waived`, alasan dan principal tercatat |
| Pemeriksaan PostgreSQL untuk rencana | Tetap `in_progress`, alasan pembatalan kosong |

Database dan Redis uji dibersihkan setelah layanan development kembali sehat. [Uji persetujuan dan pembatalan sebelum mulai](104-hrm-mutation-persetujuan-dan-pembatalan-offboarding-mfa.md).
