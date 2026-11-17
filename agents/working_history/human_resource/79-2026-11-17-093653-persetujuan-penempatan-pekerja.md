# Persetujuan penempatan pekerja

Menambahkan command/RPC `ApproveWorkerAssignment`, permission sensitif, dan validator SoD, status aktif, FTE temporal serta kapasitas posisi dengan lock parent dan outbox atomik. Menghindari urutan lock berlawanan pada proposal kompensasi. Memperluas unit/mock, katalog permission, dan tes gRPC–PostgreSQL untuk penolakan, retry, serta dua approval paralel. `go generate`, fmt, vet, seluruh tes race, build Docker, dan health HRM lulus. Penerapan penempatan primer dan aktivasi posisi produksi belum tersedia.
