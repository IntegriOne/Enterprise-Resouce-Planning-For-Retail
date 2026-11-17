# Publikasi posting rekrutmen

Menambahkan command dan RPC `PublishJobPosting`, izin sensitif tersendiri, serta validator transisi draf ke published. Command meminta bukti versi/digest immutable DRM dan menulis status, waktu server, serta outbox atomik. Tes unit/mock, gRPC–PostgreSQL disposable, fmt/vet/race/build, dan Docker healthy lulus. Runtime produksi tetap gagal tertutup karena adapter DRM belum tersedia. Close/withdraw posting dan E2E IAM–Gateway–DRM masih perlu dikerjakan.
