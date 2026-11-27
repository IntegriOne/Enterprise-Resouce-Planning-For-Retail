# Regresi HTTPS provisioning akun pekerja

Runner repositori `backend/gateway/internal/e2e/scripts/run_hrm_employee_provisioning.py` lulus pada build checkout aktif dengan Gateway, IAM, Enterprise Foundation, HRM, PostgreSQL dan Redis disposable.

| Kasus | Hasil |
| --- | --- |
| Superadmin memberikan izin dan role uji melalui GraphQL IAM | Berhasil |
| Pengajuan akun tanpa event perekrutan yang memenuhi syarat | Ditolak |
| Aktivasi pekerja mencatat event perekrutan | Pengajuan diterima di antrean IAM |
| ID worker atau jenis GID receipt berbeda dari permintaan | Ditolak |
| Pembuatan akun dan pengulangan mutation dengan receipt yang sama | Satu akun IAM dan satu receipt `linked` |
| Query ulang melalui Relay dan status HRM | `applied` dengan principal tertaut yang sama |
| Pencabutan grant dan penghapusan role uji lewat GraphQL | Berhasil |

Kontrak bertanda tangan di jalur aktivasi menggunakan fixture. Runner memeriksa persistensi kedua service dan menghapus stack disposable setelah selesai.
