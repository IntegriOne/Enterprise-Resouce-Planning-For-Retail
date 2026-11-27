# Rollout schema HRM pada Gateway lokal

Mengganti dua replika secara berurutan dengan image yang telah lulus uji startup. Kedua replika sehat dan integrasi HRM aktif. Verifikasi melalui Caddy dengan CA pengembangan membuktikan guard sesi dan CSRF. Suite HTTPS meluluskan 25 kasus guard dan melewati kasus grant resmi karena fixture belum tersedia.
