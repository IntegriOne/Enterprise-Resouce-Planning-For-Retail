# Pemrosesan satu pengajuan akun dengan dua permintaan bersamaan

Uji Caddy HTTPS memakai stack, fixture kontrak, katalog uji, dan setup grant GraphQL seperti pada [uji aktivasi](135-hrm-mutation-aktivasi-dan-provisioning-akun.md). Setelah command aktivasi dan pengajuan berhasil, dua klien HTTPS membuka koneksi berbeda dan mengirim `createEmployeeUser` secara bersamaan. Keduanya memakai sesi MFA superadmin yang sama, ID pengajuan dan pekerja yang sama, serta username dan email berbeda.

| Skenario | Hasil |
| --- | --- |
| Dua mutation dikirim bersamaan | Keduanya sukses dan mengembalikan ID akun yang sama. |
| Mutation diulang setelah kedua respons | ID akun tetap sama. |
| Status dilihat dari kedua management | Antrean terbuka IAM kosong. Detail IAM `linked` dan HRM `applied`. |
| Verifikasi PostgreSQL | Tersimpan satu akun fixture dan satu penerimaan IAM. Referensi principal HRM cocok dengan akun IAM. |

Runner lokal `/tmp/erp-cross-service.UUJx1n/run-activation-concurrent.py` keluar dengan status 0. Klaim oleh dua principal berbeda dan pemulihan setelah respons gRPC terputus masih perlu diuji.

## Regresi runner repositori

`python3 backend/gateway/internal/e2e/scripts/run_hrm_employee_provisioning.py` kini menjalankan dua browser HTTPS serentak memakai sesi MFA, receipt, worker, dan kredensial fixture yang sama. Kedua mutation dan pengulangan setelahnya mengembalikan ID akun yang sama. Pemeriksaan PostgreSQL membuktikan tepat satu akun dan satu receipt tertaut. Runner mengelola grant melalui GraphQL dan membersihkan stack disposable. Cabang konflik pada pembacaan ulang diuji terpisah pada adapter karena urutan balapan live bergantung pada penjadwalan.
