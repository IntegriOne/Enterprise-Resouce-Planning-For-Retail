# Penerapan penempatan primer

Menambahkan command/RPC `ApplyPrimaryAssignment` dengan permission sensitif. Validator menjaga worker aktif, assignment primer approved yang berlaku, posisi aktif, badan hukum, bukti Foundation, serta berakhirnya penempatan primer lama. Update worker dan outbox atomik; CRUD dan retry ditolak. Unit/mock, gRPC–PostgreSQL disposable termasuk dua permintaan paralel, fmt/vet/race, build Docker dan health HRM lulus. Jalur positif produksi masih menunggu aktivasi worker dan verifier Foundation.
