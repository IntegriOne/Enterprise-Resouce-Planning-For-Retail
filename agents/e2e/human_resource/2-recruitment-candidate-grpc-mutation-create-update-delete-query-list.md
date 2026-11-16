# Integrasi kandidat rekrutmen gRPC–PostgreSQL

Tes menggunakan `bufconn`, policy tiruan, migrasi HRM dan PostgreSQL disposable `*_test`. Ini bukti integrasi service, bukan E2E IAM/Gateway.

| Operasi | Kasus | Hasil |
| --- | --- | --- |
| Mutation CreateCandidate | Metadata notice dan retensi valid | Berhasil; creator dari actor server |
| Mutation CreateCandidate | Nomor duplikat | `AlreadyExists` |
| Mutation CreateCandidate | Persetujuan notice hilang | `InvalidArgument` |
| Mutation CreateCandidate | Referensi kontak Foundation belum diverifikasi | `Unavailable` |
| Query GetAllCandidates | Limit 1 | Satu item, total 1 |
| Mutation UpdateCandidate | Koreksi nama | Berhasil; outbox bertambah |
| Mutation UpdateCandidate | Ubah versi notice | `FailedPrecondition` |
| Mutation DeleteCandidate | Kandidat tanpa lamaran | Berhasil |

Perintah terkini: `go test -tags integration -race -count=1 ./transport/grpc/handler/recruitment/test -run TestRecruitmentDraftRPCWithPostgreSQL -timeout=2m`. Hasil: lulus pada 17 November 2026. Database sementara dihapus sesudah tes.
