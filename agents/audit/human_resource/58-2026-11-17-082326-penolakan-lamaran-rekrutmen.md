# Audit penolakan lamaran rekrutmen HRM

## Temuan dan perubahan

Command `RejectApplication` dan RPC privat memproses status `received` atau `screening` menjadi `rejected`. Screening hanya dapat ditolak oleh recruiter yang tercatat. Alasan, actor dan waktu keputusan berasal dari command server; stage history serta outbox ditulis dalam transaksi yang sama. Validator mengunci kandidat dan requisition, membekukan field lain, dan menolak pengulangan. Migrasi aditif `000004` menambah constraint status rejected wajib memiliki metadata keputusan. Tes penutupan requisition kini menolak lamaran lewat RPC, bukan pembaruan status SQL. Fixture mock riwayat memakai ID berurutan agar dapat menguji lebih dari satu transisi.

## Bukti pengujian

- `buf lint`/`buf generate`, validasi DBML (86 tabel, 9 enum), dan migrasi SQL `000004` up/down/up pada PostgreSQL disposable: lulus.
- `go fmt ./...`, `go vet ./...`, `go test -race ./...`, `go build ./...`: lulus pada modul HRM.
- gRPC `bufconn` → PostgreSQL `_test` dengan `-race`: lulus untuk penolakan dari received dan screening, larangan actor lain, retry, riwayat, outbox, constraint SQL, serta close requisition.
- Docker HRM build dan startup healthy; development DB versi migrasi 4 dengan constraint terpasang.

## Batasan

Tes policy memakai actor tiruan; belum ada grant IAM, Gateway, atau bukti owner Foundation/Finance nyata. Command shortlist/interview/assessment/offer dan keputusan kandidat eksternal belum tersedia. Migrasi menolak data historis rejected tanpa metadata; pemeriksaan database development sebelum migrasi menemukan nol baris tersebut.
