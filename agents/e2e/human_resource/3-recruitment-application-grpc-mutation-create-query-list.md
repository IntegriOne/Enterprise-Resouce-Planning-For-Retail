# Integrasi lamaran rekrutmen gRPC–PostgreSQL

Tes `TestRecruitmentDraftRPCWithPostgreSQL` memakai `bufconn`, policy tiruan dan PostgreSQL disposable `*_test`. Requisition terbuka disiapkan sebagai fixture SQL, sehingga bukti ini belum mencakup workflow pembukaan lowongan atau IAM/Gateway.

| Operasi | Kasus | Hasil |
| --- | --- | --- |
| Mutation CreateApplication | Kandidat aktif, notice/retensi sah dan requisition terbuka | `received` tersimpan |
| Mutation CreateApplication | Kandidat–requisition sama dengan nomor baru | `AlreadyExists` |
| Query GetAllApplications | Limit 1 | Satu item, total 1 |
| Mutation UpdateApplication | Transisi ke `screening` lewat CRUD | `FailedPrecondition` |
| Mutation DeleteApplication | Hapus tanpa workflow retensi | `FailedPrecondition` |
| Mutation DeleteCandidate | Kandidat punya lamaran | `FailedPrecondition` |
| Outbox | Aplikasi berhasil | Satu event; mutasi ditolak tidak menambah event |

Perintah: `go test -tags integration -race -count=1 ./transport/grpc/handler/recruitment/test -run TestRecruitmentDraftRPCWithPostgreSQL -timeout=2m`. Lulus pada 17 November 2026; database sementara dihapus sesudah tes.
