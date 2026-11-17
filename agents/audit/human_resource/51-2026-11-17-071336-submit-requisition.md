# Audit submit requisition HRM

## Temuan

Migrasi `000002` menyimpan provenance pembuat/pengusul, tetapi belum ada jalur bisnis yang menulis status `submitted`. CRUD sengaja menolak status dan metadata workflow. Approval/open membutuhkan status submit yang sah dan pembuktian owner lebih lanjut.

## Perbaikan

- RPC privat `SubmitRequisition` memakai command dengan permission sensitif `human_resource.requisitions.submit`; jumlah kode runtime menjadi 329. Tidak ada grant IAM baru.
- Command mengunci draf, mewajibkan actor manusia sama dengan pembuat, mengisi waktu/principal submit dari server, memeriksa ulang referensi melalui Integrity, dan menulis outbox dalam transaksi yang sama.
- Validator menolak perubahan field lain selama submit dan tetap melarang transisi lewat CRUD. Submit ulang, update, dan delete setelah submit ditolak.

## Bukti dan batas

Tes unit/mock lulus untuk actor lain, actor nonhuman, rollback integrity, dan retry. Tes gRPC `bufconn`→PostgreSQL `_test` lulus untuk submit, metadata server, dua event outbox, serta penolakan CRUD/retry. `buf lint`, generator proto resmi, `go fmt ./...`, `go vet ./...`, `go test -race ./...`, `go build ./...`, build Docker dan healthcheck HRM lulus. Tes memakai policy/verifier Foundation tiruan; registry produksi masih memakai verifier nil dan menolak submit dengan `Unavailable`. Approval/open, grant IAM, dan E2E Gateway belum tersedia.
