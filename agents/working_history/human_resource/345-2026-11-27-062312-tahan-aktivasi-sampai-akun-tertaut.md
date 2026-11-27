# Aktivasi akun menunggu penautan HRM

IAM kini menolak aktivasi manual dan aktivasi melalui verifikasi identitas jika penerimaan provisioning akun belum berstatus linked. Akun lama tanpa penerimaan mengikuti alur yang sudah ada. Tes modul IAM lulus. Simulasi HTTPS membuktikan activateUser ditolak ketika receipt masih processing dan berhasil setelah retry penautan selesai.
