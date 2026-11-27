# Penautan akun employee terverifikasi

Commit `a337d67` terverifikasi pada `origin/main`. Runtime HRM menerima bukti penautan yang mengikat pekerja, party, principal dan operator IAM. Index unik parsial membatasi satu hubungan kerja yang masih berlaku untuk satu akun.

Tes PostgreSQL disposable lulus untuk alur gRPC penautan, penolakan bukti yang salah, penolakan update biasa, pembacaan tautan hasil penyimpanan, dan migrasi riwayat yang dikenali. Buf lint, race suite, vet, dan build lulus. Database aktif tidak diubah. Publikasi Enterprise Foundation dan Gateway serta E2E HTTPS dari main masih perlu diselesaikan.
