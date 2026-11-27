# Pembatalan penempatan mendatang

Terminasi menandai penempatan mendatang dengan waktu pembatalan dan menjaga riwayat persetujuan. Penempatan batal dikecualikan dari kapasitas dan constraint benturan periode primer, serta tidak dapat dipakai sebagai penempatan aktif. Query Personnel membawa waktu pembatalan. Tes gRPC–PostgreSQL dengan race detector mencakup pembatalan, audit data lama, dan rollback jika waktu persetujuan berada di masa depan. Baseline berubah dan database development lama perlu penanganan checksum. [Laporan](../../e2e/human_resource/41-personnel-mutation-terminate-worker.md)
