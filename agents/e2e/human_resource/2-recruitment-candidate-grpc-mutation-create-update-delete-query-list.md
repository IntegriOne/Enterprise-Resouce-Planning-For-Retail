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

## Referensi individu Enterprise Foundation

Pengujian adapter HRM melalui gRPC bertanda tangan menerima individu aktif. Adapter menolak organisasi, individu nonaktif, individu terhapus, ID yang tidak ditemukan, dan respons dengan ID berbeda. Pengujian memakai server EF fixture. Uji transaksi kandidat melalui gRPC dan PostgreSQL membuktikan penyimpanan referensi yang diterima. Build HRM dipasang di Compose development, berstatus sehat, dan lulus uji baca pekerja melalui HTTPS. Operasi kandidat belum memiliki jalur HTTPS pada rangkaian uji ini.

Verifikasi kontak kini memeriksa ID kontak, status aktif, dan ID individu pemilik melalui gRPC EF. Fixture gRPC menolak kontak nonaktif, kontak milik party lain, dan respons dengan ID berbeda. Uji transaksi PostgreSQL menerima pasangan individu dan kontak yang valid serta menolak kontak tanpa referensi individu. Build HRM yang memuat perubahan ini telah dipasang di development. Uji HTTPS superadmin pada query pekerja berhasil setelah rollout.

Regresi penuh setelah perubahan kontak selesai dengan status 0. Sebanyak 68 paket lulus lewat `go test -race -p 1 -count=1 -tags=integration ./...` memakai PostgreSQL disposable dan font PDF pengujian.
