# Verifikasi digest outbox setelah pembacaan JSONB

PostgreSQL mengubah format serialisasi payload JSONB. Relay sebelumnya membandingkan byte hasil baca dengan digest dari JSON map yang ditulis Go, sehingga event yang sah dapat dianggap rusak. Relay kini menyerialisasi ulang objek sesuai format penulis sebelum menghitung SHA-256. Tes PostgreSQL dan race detector membuktikan kegagalan sebelum perubahan dan kelulusan sesudahnya. Tes payload dengan digest salah tetap lulus dengan penolakan.
