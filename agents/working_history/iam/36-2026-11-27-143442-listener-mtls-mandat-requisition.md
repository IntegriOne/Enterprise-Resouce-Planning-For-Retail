# Listener mTLS mandat requisition

IAM dapat membuka listener privat yang hanya menyediakan verifikasi mandat. Konfigurasi mengharuskan alamat serta material TLS lengkap. Tes TCP membuktikan handshake tanpa sertifikat gagal, workload lain ditolak, dan sertifikat HRM mendapat keputusan. Race test dan go vet lulus. Sertifikat deployment dan klien HRM masih belum dihubungkan.
