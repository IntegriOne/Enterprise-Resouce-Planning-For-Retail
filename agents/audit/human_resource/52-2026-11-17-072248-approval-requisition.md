# Audit approval requisition HRM

## Temuan

Status `submitted` dan provenance tersedia, tetapi approval belum mempunyai command yang memisahkan pengusul dari penyetuju serta meminta bukti headcount dan anggaran. Rekaman budget hanya dapat dipercaya bila berasal dari owner, bukan input CRUD. Kontrak owner Foundation/Finance belum terpasang pada runtime produksi.

## Perbaikan

- RPC privat `ApproveRequisition` memakai permission sensitif `human_resource.requisitions.approve`; katalog runtime menjadi 330 kode tanpa grant IAM baru.
- Service mengunci requisition, menolak self approval, memanggil port bukti owner berdasarkan snapshot requisition, memeriksa referensi anggaran UUID, lalu menyimpan status, budget, actor/waktu dan outbox dalam satu transaksi.
- Validator menolak perubahan lain saat approval, approval tanpa provenance/proof, delete setelah submit, dan pembaruan workflow lewat CRUD. Gangguan verifier serta bukti tidak sah gagal tertutup.

## Bukti dan batas

Unit/mock membuktikan self approval, outage owner, proof tak sah, rollback dan retry. Tes gRPC `bufconn`→PostgreSQL `_test` dengan policy/verifier tiruan membuktikan approval sah, penolakan pengusul, retry, serta dua approval paralel dengan satu hasil sukses dan satu event outbox. `buf lint/generate`, `go fmt ./...`, `go vet ./...`, `go test -race ./...`, `go build ./...`, Docker build dan healthcheck HRM lulus. Registry produksi belum menginjeksi verifier Foundation/Finance, sehingga command tetap `Unavailable`. Bukti owner nyata, grant IAM, pembukaan lowongan, dan E2E Gateway belum tersedia.
