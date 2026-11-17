# Audit penarikan posting rekrutmen HRM

## Temuan dan perubahan

Posting published kini dapat ditarik melalui `WithdrawJobPosting` dengan permission sensitif tersendiri. Alasan, actor dan waktu dari server tersimpan pada posting dan outbox dalam satu transaksi; validator mengunci requisition open, membekukan field lain, dan menolak pengulangan. Migrasi aditif `000003` menambahkan tiga kolom serta constraint untuk mewajibkan metadata penarikan hanya pada status withdrawn. Finder alasan dan actor mengikuti paginasi dan scope service.

## Bukti pengujian

- `buf lint`/`buf generate`, validasi DBML (86 tabel, 9 enum), dan SQL migrasi `000003` up/down/up pada PostgreSQL disposable: lulus.
- `go fmt ./...`, `go vet ./...`, `go test -race ./...`, `go build ./...`: lulus pada modul HRM.
- gRPC `bufconn` → PostgreSQL `_test` dengan `-race`: lulus untuk alasan wajib, penarikan, finder, retry, outbox, constraint SQL, dan penutupan requisition.
- Docker HRM build dan startup healthy; development DB melaporkan migrasi versi 3 dan tiga kolom baru.

## Batasan

Status withdrawn merupakan keputusan HRM, belum konfirmasi tayangan di kanal eksternal sudah dihapus. Tes memakai verifier DRM tiruan dan policy tiruan; grant IAM, Gateway, dan DRM nyata belum diuji. Migrasi menolak baris historis berstatus withdrawn tanpa metadata; pemeriksaan database development sebelum migrasi menunjukkan nol baris tersebut.
