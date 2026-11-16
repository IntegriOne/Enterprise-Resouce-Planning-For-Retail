# Audit semua domain dan sanitasi error gRPC IAM

Pada 17 November 2026, enam kelompok/81 RPC IAM diperiksa dari kontrak, handler, service, port domain, repository, hingga wiring. ErrorMapper ditemukan meneruskan status gRPC terbungkus beserta detail internal; tes regresi mereproduksi kebocoran dan perbaikan menutup jalur itu. `go test -race`, vet, build dan diff check lulus. Audit juga mengoreksi jumlah RPC daftar relasi tanpa pagination menjadi tujuh karena `ListPermissionsByRole` terlewat. Batas bukti PostgreSQL/Redis dan temuan terbuka dicatat pada audit IAM nomor 13.
