# Uji HTTPS guard dan provisioning kandidat publikasi

Gateway, HRM, IAM, dan Enterprise Foundation memakai binary kandidat yang dibangun dengan `GOWORK=off`. Caddy, Redis, dan database disposable menjalankan pengujian HTTPS secara terisolasi.

| Kasus | Hasil |
| --- | --- |
| Kontrak SDL kandidat | 24 tipe Node dan 24 Connection lolos pemeriksaan. |
| Seluruh query HRM tanpa sesi | 52 dari 52 mengembalikan `UNAUTHENTICATED`. |
| Seluruh mutation HRM tanpa header CSRF | 55 dari 55 mengembalikan `FORBIDDEN` karena header CSRF tidak ada. |
| Provisioning setelah jendela rate limit | Superadmin berhasil memasang izin melalui GraphQL IAM. Satu akun dan satu receipt tertaut tersimpan, dan pengulangan tidak menghasilkan duplikat. |
| Pembersihan | Role uji dicabut melalui GraphQL. Stack dan database disposable dibersihkan. |

Percobaan pertama menjalankan provisioning segera setelah guard dan mencapai rate limit. Pengulangan yang berhasil memberi jeda antarsuite dengan batas rate limit tetap aktif. [Rincian provisioning](169-iam-mutation-create-employee-user-kandidat-publikasi.md).
