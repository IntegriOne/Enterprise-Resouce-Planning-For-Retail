# Publikasi Gateway HRM

Commit `cefcf5f` terverifikasi pada `origin/main`. Gateway memuat SDL Relay, resolver dan adapter IAM, EF, HRM, serta konfigurasi runtime terkait. Kredensial lokal tidak ikut commit.

Regenerasi SDL, vet, seluruh tes race Gateway, dan build Go serta Docker lulus. Pada stack HTTPS disposable dari main, antrean provisioning menolak akses anonim. Login superadmin belum lulus karena main masih memakai bootstrap lama. Pengujian alur GraphQL terautentikasi, publikasi bootstrap katalog, dan verifikasi Finance/DRM masih terbuka.
