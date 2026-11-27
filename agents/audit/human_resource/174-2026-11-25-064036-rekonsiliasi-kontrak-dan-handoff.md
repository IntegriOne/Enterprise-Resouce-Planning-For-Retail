# Rekonsiliasi kontrak dan handoff gRPC HRM

Pencocokan per service antara nama RPC di proto dan baris handoff menemukan 1000 RPC pada 22 service. Nama cocok tanpa operasi hilang atau duplikat. Tabel handoff mengklasifikasikan 667 Query dan 333 Mutation. Angka pemeriksaan terdahulu di indeks diselaraskan dengan hasil ini.

`go vet ./...` dan `go test -count=1 ./...` lulus pada checkout HRM. Pemeriksaan ini menggunakan tes tanpa tag integrasi. `AcceptOffer` dan `DeclineOffer` sudah memiliki RPC dan handler. Jalur produksi masih membutuhkan verifier respons kandidat.

Celah workflow yang dikonfirmasi di kontrak Personnel adalah ketiadaan RPC terminasi dan penutupan kontrak atau penempatan. Penutupan offboarding mensyaratkan status dan tanggal terminasi pekerja. Alur terminasi perlu diuji melalui RPC hingga penutupan offboarding. [Handoff Personnel](../../grpc/hrm/17-operasi-personnel.md#terminasi-pekerja)
