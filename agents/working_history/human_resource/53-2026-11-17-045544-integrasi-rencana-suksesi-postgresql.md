# Integrasi rencana suksesi PostgreSQL

Ditambah tes integrasi `bufconn` gRPC untuk draf rencana suksesi pada PostgreSQL disposable. Kasus meliputi create, duplikasi, pagination, update tanggal, penolakan field workflow, kandidat fail-closed, delete, dan tiga event outbox. Fixture posisi dibuat langsung karena validator produksi posisi belum ada. Tes race integrasi, format, vet, build, dan seluruh tes race modul HRM lulus. Database uji dihapus. Policy tiruan hanya membuktikan alur database; IAM/Gateway serta workflow kandidat tetap belum selesai. Detail pada audit HRM nomor 37.
