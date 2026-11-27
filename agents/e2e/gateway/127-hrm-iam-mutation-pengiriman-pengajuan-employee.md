# Uji HTTPS pengiriman pengajuan akun employee ke IAM

Basis Git `e28491c` dengan perubahan worktree. Gateway dan IAM dibangun ulang dari source aktif. HRM dan Enterprise Foundation dijalankan pada stack uji privat dengan PostgreSQL `*_test` dan Redis sementara. IAM uji memuat katalog permission fixture. Superadmin menyiapkan dan mencabut role/permission lewat GraphQL saja.

| Kasus | Hasil |
| --- | --- |
| `hrmRequestEmployeeAccount` untuk worker dan event lifecycle fixture | PASS, payload berisi ID pengajuan HRM dan ID penerimaan IAM. |
| Kirim ulang input yang sama | PASS, kedua ID sama dengan hasil awal. |
| PostgreSQL IAM setelah pengiriman ulang | PASS, hanya satu penerimaan tersimpan. |
| Validasi target badan hukum di EF | PASS untuk target aktif dan penolakan target acak pada uji scope. |
| Regresi `createEmployeeUser` dengan MFA | PASS, relasi akun dan worker cocok. Jalur ini masih menggunakan Worker ID dan belum mengklaim antrean. |

Aktivasi worker dan event lifecycle menggunakan fixture SQL pada database HRM disposable. Percobaan pertama terkena timeout ke alamat DNS EF sementara. Uji lulus setelah runner memakai alamat jaringan container EF yang baru. Semua container dan database uji dibersihkan.

Runner lokal: `python3 /tmp/erp-cross-service.UUJx1n/run-ingress.py`.
