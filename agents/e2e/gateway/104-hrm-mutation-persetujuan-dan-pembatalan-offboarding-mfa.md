# Uji HTTPS persetujuan independen dan pembatalan offboarding

Database IAM, HRM, Foundation dan Redis uji terisolasi dari development. Superadmin fixture membuat pekerja, rencana, dan tugas. Status aktif pekerja dan verifikasi identitas employee kedua disiapkan lewat SQL fixture yang membatasi operasi pada database `*_test`. Akun employee kedua dibuat, diaktifkan, dan diberi role melalui API IAM. Kedua akun memakai sesi TOTP dan CSRF yang sah.

| Kasus | Hasil |
| --- | --- |
| Pembuat mencoba menyetujui rencana sendiri | `FORBIDDEN` |
| Employee kedua menyetujui draf dengan tugas wajib | Status `approved` |
| Employee kedua membatalkan rencana sebelum dimulai | Status `cancelled` |
| Query detail tugas setelah pembatalan | Status `pending` |
| Pemeriksaan PostgreSQL | Pembuat berbeda dari approver, alasan pembatalan tercatat, dan tugas tetap tersimpan |

Database dan Redis uji dibersihkan setelah layanan development sehat. [Uji pembuatan draf](103-hrm-query-mutation-draf-offboarding-mfa.md).
