# Konkurensi pembuatan draf suksesi

Tanggal: 17 November 2026 WIB. Tes integrasi suksesi kini mengirim dua `CreateSuccessionPlan` secara bersamaan melalui gRPC `bufconn` ke PostgreSQL disposable untuk pasangan posisi–kode–versi yang sama. Tepat satu permintaan berhasil dan satu mendapat `AlreadyExists`; permintaan duplikat berikutnya juga ditolak. Setelah update dan delete draf, jumlah event outbox untuk agregat itu tetap tiga (create, update, delete), sehingga percobaan duplikat tidak menambah event.

`GOWORK=off go test -race -tags integration ./transport/grpc/handler/succession/test -run TestSuccessionPlanRPCWithPostgreSQL -count=10 -v` lulus sepuluh kali. `go fmt`, `go vet`, `go build`, dan `go test -race -count=1 ./...` lulus pada modul HRM. Database disposable dihapus; tidak ada yang tersisa pada pola `hrm_disposable_*_test`.

Tes memakai policy tiruan dan posisi fixture. Ini bukti konkurensi untuk satu constraint draf plan, bukan bukti nominasi kandidat, mandat panel, IAM/Gateway, atau seluruh mutasi suksesi.
