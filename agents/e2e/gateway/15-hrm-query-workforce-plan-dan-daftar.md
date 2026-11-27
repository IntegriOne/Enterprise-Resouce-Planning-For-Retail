# Pengujian integrasi Query rencana tenaga kerja

Pengujian memakai endpoint GraphQL Gateway dengan sesi fixture dan Workforce Planning gRPC melalui bufconn.

| Kasus | Hasil |
| --- | --- |
| Detail UUID dengan izin baca rencana berscope badan hukum | ID, kode dan versi rencana |
| Daftar dengan izin yang sama, limit 1 dan offset 1 | Satu item dan total 2 dari gRPC |
| Daftar dengan izin baca pekerja saja | `FORBIDDEN` |

Ketiga kasus lulus dalam `TestHRMQueriesThroughGateway`. Suite Gateway dengan race detector dan `go vet ./...` juga lulus. Penerapan filter badan hukum oleh HRM dengan akun IAM resmi belum diuji lewat HTTPS.
