# Blokir grant scope eksternal tanpa verifikasi owner

IAM sebelumnya menerima UUID scope eksternal hanya berdasar format. Guard domain baru menolak penugasan role pengguna/grup untuk scope luar hingga kontrak pembuktian owner tersedia; hasil gRPC `Unavailable`. Scope global dan self tetap mengikuti aturan IAM. Tes domain, service dan bufconn diperbarui. Kebutuhan kontrak pemilik dicatat di `agents/notes/enterprise_foundation/1-iam.md`, termasuk status objek, jenis, relasi, versi, dan kegagalan dependency. Perubahan tidak membaca database management lain.
