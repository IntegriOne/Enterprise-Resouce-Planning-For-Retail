# Uji browser pemrosesan pengajuan akun employee

Basis Git `e28491c` dengan perubahan worktree. Frontend Next.js dibangun dengan Docker Compose, kemudian dijalankan bersama Gateway, IAM, Enterprise Foundation, dan HRM pada port HTTPS uji `18443`. PostgreSQL dan Redis memakai lingkungan disposable. Katalog izin berasal dari build IAM normal.

| Kasus | Hasil |
| --- | --- |
| Persiapan role dan izin untuk superadmin | PASS melalui GraphQL IAM. |
| Pengajuan worker dari HRM masuk antrean IAM | PASS melalui `hrmRequestEmployeeAccount`. |
| Browser membuka Create Employee dengan sesi superadmin yang sudah MFA | PASS. Form memuat dan menampilkan nomor pekerja dari antrean. |
| Pilih pengajuan dan isi identitas akun pada form | PASS. Request browser membawa `provisioning_request_id` dan `worker_id` dari opsi terpilih. |
| Kirim form | PASS. Modal tertutup dan pengajuan keluar dari antrean terbuka. |
| Ulangi Mutation dan baca status HRM | PASS. IAM mengembalikan akun yang sama dan HRM menampilkan `applied`. |
| PostgreSQL | PASS. Tepat satu akun dan satu penerimaan tertaut. |

Login dan MFA dijalankan lewat HTTPS GraphQL sebelum sesi uji dipasang ke profil Chrome sementara. Uji browser berfokus pada interaksi form dan payload Mutation. Aktivasi pekerja memakai fixture PostgreSQL HRM disposable. Browser, container, dan database sementara dibersihkan setelah eksekusi.

Runner lokal: `python3 /tmp/erp-cross-service.UUJx1n/run-browser.py`.

## Uji ulang detail yang terlihat di form

Browser memilih pengajuan dan melihat nama pekerja dari data pihak Enterprise Foundation, nomor pekerja HRM, serta nama dan kode badan hukum. Pemeriksaan teks form tidak menemukan `gid://`. Browser tetap mengirim pasangan ID yang benar melalui GraphQL. Akun tertaut dan hasil Query HRM tetap `applied`.

## Pesan kesalahan tanpa referensi internal

Chrome lewat HTTPS menerima satu kegagalan simulasi saat memuat antrean dan satu konflik simulasi saat mengirim formulir. Kedua respons menyertakan `gid://` di pesan sumber. Dialog menampilkan pesan lokal sesuai tindakan, tanpa ID tersebut. Setelah memuat ulang dan mencoba kembali, operator berhasil memproses pengajuan melalui Gateway, IAM, dan HRM. Pemeriksaan database menemukan satu akun dan satu penerimaan.
