# Verifikasi jalur runtime IAM–Gateway–HRM

Probe aktif `docker exec integrione-retail-human-resource-1 /usr/local/bin/grpc-healthcheck` berakhir dengan kode 0. Status container HRM adalah `healthy`. IAM, kedua replika Gateway, dan Enterprise Foundation juga berstatus `healthy` pada saat pemeriksaan. Ini membuktikan kesiapan proses yang sedang berjalan dan belum memverifikasi image hasil build dari perubahan terkini.

Konfigurasi gqlgen pada checkout memuat `graphql/**/*.graphql`. Direktori skema memuat IAM dan Enterprise Foundation. HRM belum memiliki skema maupun wiring klien di Gateway checkout ini, sehingga permintaan HRM yang autentik melalui `/api/graphql` belum dapat diuji. DRM dan Finance juga belum tersedia sebagai service pemilik di bawah `backend/services/`. Tidak ada grant atau data bisnis yang diubah selama verifikasi.
