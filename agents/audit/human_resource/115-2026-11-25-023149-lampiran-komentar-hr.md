# Lampiran komentar tiket HR

Tes gRPC dengan PostgreSQL sementara mereproduksi penyimpanan UUID lampiran tanpa verifikasi DRM. Validator kini mengembalikan Unavailable untuk komentar berlampiran sampai verifier versi dokumen pada event tersedia. Komentar tanpa lampiran dapat dibuat, termasuk dengan idempotency key yang dipakai percobaan berlampiran yang ditolak. Lampiran historis tidak diubah.
