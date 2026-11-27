# Uji HTTPS Query hrmPayrollRuleSets pada Gateway hasil isolasi

Gateway dibangun dari salinan `origin/main` dengan kontrak dan wiring HRM. Caddy sementara menggunakan sertifikat development yang sama dan meneruskan permintaan HTTPS port 18443 ke build tersebut. Gateway terhubung ke IAM dan HRM development yang sedang berjalan.

| Kasus | Hasil |
| --- | --- |
| Health melalui HTTPS dengan verifikasi CA | `health: ok` |
| Login superadmin dan pengambilan CSRF | Berhasil |
| Cari permission `human_resource.payroll_rule_sets.read`, buat role sementara, pasang permission, dan assign role ke superadmin | Berhasil melalui GraphQL IAM |
| Login ulang dan panggil `hrmPayrollRuleSets` | Berhasil, `total_count: 0` |
| Cabut penugasan role, cabut permission, hapus role, dan logout | Berhasil melalui GraphQL IAM |

Hasil query mencakup respons daftar kosong. Setelah uji, kedua container sementara dihentikan.

## Uji ulang setelah penyelarasan MutationResult

Build isolasi menjalankan `HRM_GRPC_ENABLED=true`. Kedua replica Gateway development ditemukan memakai `false` saat pengujian ulang. Percobaan yang mewarisi pengaturan itu menghasilkan `UNAVAILABLE` untuk Query HRM. Pengujian dengan HRM diaktifkan pada Gateway sementara membuktikan kasus berikut.

| Operasi | Hasil |
| --- | --- |
| Query `hrmPayrollRuleSets` | PASS, `total_count: 0` |
| Mutation `hrmUpdateWorker` dengan input kosong | PASS, `was_successful: false`, `errors[0].type: ARGUMENT`, `worker: null`. Ditolak saat validasi Gateway sebelum panggilan gRPC. |
| Penyiapan dan cleanup role uji | PASS melalui GraphQL IAM untuk kedua permission uji |

Pengujian ini memakai build lokal yang belum memiliki SHA publikasi. Kasus sukses Mutation dengan perubahan state memerlukan pengujian terpisah.
