# Audit cakupan validator dan regresi HRM gRPC

- Kontrak mendeklarasikan 22 service bisnis dan `RegisterAll` memanggil 22 registrasi domain. Ini bukti cakupan registrasi, sementara perilaku RPC dinilai melalui tes masing-masing.
- Tes `TestBuildMutationValidatorsCoversAllCatalogMutatableResources` membangun registry produksi dengan dependency offline. Seluruh 78 resource berkode mutasi terdaftar, ditambah validator untuk ledger cuti. Lima resource baca lainnya tidak mempunyai validator mutasi generik.
- Suite modul `GOWORK=off go test -p 1 -tags=integration -race -count=1 ./...` lulus pada 67 paket. DSN mengarah ke database PostgreSQL disposable berakhiran `_test` dan database dihapus setelah tes.
- Wiring runtime memiliki klien Foundation yang diaktifkan melalui konfigurasi. Tes ini belum membuktikan grant IAM, sesi Gateway, atau verifier DRM dan settlement Finance produksi.
