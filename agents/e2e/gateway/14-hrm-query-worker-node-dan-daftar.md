# Pengujian integrasi Query pekerja HRM melalui Gateway

Gateway menjalankan schema dan middleware yang sama dengan entrypoint aplikasi. Tes mengirim HTTP POST ke `/api/graphql` dengan penyimpanan sesi fixture dan Personnel gRPC melalui bufconn.

| Kasus | Hasil |
| --- | --- |
| Daftar tanpa cookie sesi | `UNAUTHENTICATED` |
| Daftar dengan sesi employee tanpa izin baca | `FORBIDDEN` |
| Detail berdasarkan UUID dengan grant badan hukum | GID pekerja dan nomor pekerja fixture |
| Node berdasarkan GID HRM dengan grant badan hukum | Tipe `HRMWorker` dan nomor pekerja fixture |
| Daftar dengan grant badan hukum, limit 1 dan offset 1 | Satu item, total 2, limit dan offset diteruskan |

Kelima kasus lulus melalui `go test -race ./internal/server -run TestHRMQueriesThroughGateway -v`. `go test -race ./...` dan `go vet ./...` Gateway juga lulus. Pengujian memverifikasi pemetaan GraphQL dan penerusan RPC. Verifikasi akun IAM resmi dan filter badan hukum HRM melalui jalur HTTPS masih tersisa.
