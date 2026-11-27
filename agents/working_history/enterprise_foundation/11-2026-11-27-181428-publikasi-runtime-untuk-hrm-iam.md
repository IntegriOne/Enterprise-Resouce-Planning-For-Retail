# Publikasi runtime untuk HRM dan IAM

Commit `db79aea` terverifikasi di `origin/main`. Enterprise Foundation menyediakan gRPC pemilik badan hukum, unit organisasi, lokasi, alamat, dan party. Publikasi menyertakan baseline PostgreSQL, image, init database baru, dan konfigurasi koneksi IAM melalui Compose.

Kandidat terisolasi lulus 88 tes dengan race detector dan PostgreSQL disposable, termasuk rollback migrasi, alur aplikasi, dan outbox. Buf lint, vet, build Go, build Docker, dan pemeriksaan deklarasi service Compose lulus. Database aktif tidak diubah. Wiring Gateway dan bukti HTTPS lintas management dari main masih diperlukan.
