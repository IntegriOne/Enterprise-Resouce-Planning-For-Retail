# Audit transisi awal screening lamaran HRM

## Cakupan dan temuan

Lamaran yang dibuat pada status `received` sebelumnya tidak mempunyai command untuk pindah tahap. `UpdateApplication` sudah menolak perubahan status, sedangkan tabel riwayat tahap hanya dibaca. Kondisi ini membuat status screening tidak dapat dicapai melalui runtime resmi.

## Perbaikan

- `StartApplicationScreening` menambah RPC privat dan permission sensitif `human_resource.applications.start_screening` (katalog aktif 328 kode), tanpa grant IAM.
- Command mengunci lamaran, mewajibkan actor manusia, dan hanya menerima transisi `received` ke `screening`. Recruiter diambil dari actor tepercaya. Validator memeriksa ulang kandidat aktif, notice/retensi, requisition terbuka, dan melarang perubahan field lain.
- Riwayat tahap dan outbox ditulis dalam transaksi perubahan lamaran. CRUD generik tetap tidak boleh mengubah status; pengulangan command ditolak.

## Bukti

Tes gRPC `bufconn` → PostgreSQL disposable `_test` dengan race detector lulus: kandidat ditarik menolak command tanpa riwayat, jalur sah menghasilkan status/recruiter/riwayat/outbox, retry ditolak. `GOWORK=off go test -race -count=1 ./...`, `go vet ./...`, `go build ./...`, `buf generate`, dan Docker build/up `human-resource` healthy lulus.

## Batas

Tes masih memakai policy tiruan dan requisition `open` dari fixture SQL. Belum ada grant IAM nyata, E2E Gateway, workflow membuka requisition, atau tahap lamaran setelah screening. Mutasi lain yang belum memiliki validator tetap gagal tertutup.
