# Uji HTTPS Gateway bersama IAM, Enterprise Foundation, dan HRM

Basis Git e28491c dengan perubahan worktree. Gateway, Enterprise Foundation, dan HRM dibangun ulang dari source aktif. IAM memakai container development yang sudah berjalan. Gateway sementara menggunakan koneksi gRPC HRM aktif dan mengarah ke Enterprise Foundation serta HRM sementara dengan database PostgreSQL `*_test`. HTTPS memakai CA development pada port 18443.

| Operasi | Bukti |
| --- | --- |
| Login superadmin, buat role uji, pasang tiga permission, assign role, login ulang | PASS melalui GraphQL IAM |
| `createLegalEntity` | PASS pada build EF saat ini, badan hukum aktif |
| `hrmCreateWorkforcePlan` | PASS, badan hukum yang dibuat di EF diterima setelah validasi HRM → EF |
| Pengajuan draf duplikat | PASS, `CONFLICT` pada payload |
| `hrmUpdateWorkforcePlan` dan `hrmWorkforcePlan` | PASS, skenario hasil update terbaca kembali |
| Baca PostgreSQL HRM `*_test` | PASS, satu baris `draft\|E2E-UPDATED` |
| Cabut assignment dan izin, hapus role, logout | PASS melalui GraphQL IAM |

Percobaan awal yang memakai Gateway image lama khusus untuk fixture EF menghasilkan `UNAVAILABLE`. Pengujian lulus setelah semua operasi dikirim ke Gateway baru yang sama. Kedua database bisnis dan container uji dibersihkan. Pengujian ini belum menjalankan provisioning akun employee, pengecekan IAM → EF secara langsung, cakupan akses negatif, maupun build IAM terbaru pada runtime uji.

Runner lokal: `python3 /tmp/erp-cross-service.UUJx1n/run-positive.py`.
