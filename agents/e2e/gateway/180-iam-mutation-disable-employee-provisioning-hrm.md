# Uji HTTPS deaktivasi employee hasil provisioning HRM

Stack disposable menjalankan build Gateway, IAM, HRM, dan Enterprise Foundation. Superadmin memasang role dan izin lewat GraphQL IAM. HRM membuat pengajuan akun dan IAM membuat satu akun tertaut. Status worker aktif berasal dari command HRM dengan fixture kontrak bertanda tangan. Verifikasi identitas login memakai fixture SQL pada database IAM disposable.

| Kasus | Hasil |
| --- | --- |
| Aktivasi dan login employee tertaut | Berhasil |
| Superadmin menjalankan `disableUser` | Akun menjadi `DISABLED` |
| Sesi employee lama membaca `me` | `UNAUTHENTICATED` |
| Employee login lagi dengan kredensial yang sama | Ditolak |
| Pembersihan role dan permission lewat GraphQL IAM | Berhasil |

Percobaan awal memakai receipt yang telah diproses oleh tes Go dengan kredensial berbeda. Replay mengembalikan akun yang sama. Uji berikutnya membuat akun sekali dengan kredensial fixture yang dipakai untuk login. Jalur ini membuktikan deaktivasi manual IAM melalui HTTPS. Pengiriman event terminasi otomatis dari HRM masih terbuka.

Skenario sudah menjadi bagian dari runner repositori. Jalankan `python3 backend/gateway/internal/e2e/scripts/run_hrm_employee_provisioning.py` dari root. Runner membagikan kredensial fixture yang sama ke subtes Go dan Python serta memakai stack disposable terpisah dari konfigurasi mandat IAM development.
