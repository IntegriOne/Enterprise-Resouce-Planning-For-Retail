# Verifikasi schema HRM pada Gateway lokal

Permintaan tanpa cookie dan token CSRF dikirim ke Caddy HTTPS dan masing-masing port Gateway lokal. ID yang dipakai adalah fixture. Pemeriksaan mengamati pemuatan schema dan respons akses anonim tanpa perubahan data.

| Operasi | Caddy HTTPS | Gateway 5000 | Gateway 5001 |
| --- | --- | --- | --- |
| `hrmWorker` | UNAUTHENTICATED | UNAUTHENTICATED | UNAUTHENTICATED |
| `hrmWorkerAssignments` | GRAPHQL_VALIDATION_FAILED | GRAPHQL_VALIDATION_FAILED | GRAPHQL_VALIDATION_FAILED |
| `hrmDeletePosition` | GRAPHQL_VALIDATION_FAILED | GRAPHQL_VALIDATION_FAILED | GRAPHQL_VALIDATION_FAILED |

Kedua replika sehat dan memakai image `sha256:894cfa82993e671b1c20f2bf1e30175d9da29fdc3de503bfe21e3de5b8424e87`. Dua field terbaru tersedia pada schema checkout dan belum termuat di image tersebut. `go build ./...` Gateway di checkout lulus. Uji positif memerlukan image yang memuat schema baru serta sesi dengan grant IAM resmi.

## Image terbaru pada container uji

Image `sha256:39c7c03ac070f47056ad606963215dd03f945e05ca648037e01ebcb1483f69d5` dibangun dari checkout dan dijalankan sampai sehat pada port loopback sementara.

| Permintaan tanpa sesi | Hasil |
| --- | --- |
| `hrmWorker` | UNAUTHENTICATED |
| `hrmWorkerAssignments` | UNAUTHENTICATED |
| `hrmDeletePosition` tanpa CSRF | FORBIDDEN |

Schema baru termuat dan pemeriksaan akses berjalan. Container uji dihentikan setelah pemeriksaan. Kedua replika aktif masih menggunakan image lama. Uji ini menggunakan HTTP loopback tanpa sesi dan belum membuktikan akses HTTPS dengan grant IAM resmi.

## Rollout lokal dan verifikasi HTTPS

Kedua replika kini sehat dengan image `sha256:39c7c03ac070f47056ad606963215dd03f945e05ca648037e01ebcb1483f69d5` dan integrasi HRM aktif. Image sebelumnya tersedia dengan tag `hrm-before-schema-rollout`. Pengujian HTTPS memakai CA pengembangan dari `backend/deploy/certs/ca.crt`.

| Pemeriksaan | Hasil |
| --- | --- |
| Query assignment tanpa sesi di Caddy | UNAUTHENTICATED |
| Mutation hapus posisi tanpa CSRF di Caddy | FORBIDDEN |
| Guard cuti, payroll, dan filter pekerja via suite e2e | 25 subtes lulus |
| Query pekerja dengan grant badan hukum | Dilewati karena fixture akun dan pekerja lintas badan hukum belum tersedia |
