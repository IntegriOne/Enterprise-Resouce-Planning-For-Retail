# Regresi HRM setelah pembatalan tiket

Regresi modul HRM pada PostgreSQL disposable lulus dengan `GOWORK=off go test -p 1 -tags=integration -race -count=1 ./...`. Sebanyak 68 paket tes lulus, termasuk paket handler 22 domain bisnis. `HRM_PAYSLIP_TEST_FONT` memakai DejaVu Sans agar jalur renderer PDF juga dijalankan.

`GOWORK=off go build ./...`, `GOWORK=off go vet ./...` dan `buf lint` lulus. Rekonsiliasi proto, handler, tipe masukan dan keluaran, klasifikasi serta izin mencocokkan 1.007 operasi dengan 669 Query dan 338 Mutation. [Inventaris](../../grpc/hrm/1-inventaris-operasi-grpc-hrm.md)

Pengujian CancelServiceRequest memakai policy fixture. Aktivasi grant IAM dan resolver GraphQL merupakan pekerjaan integrasi. Finance dan DRM produksi diperlukan untuk menguji penerbitan dan akses slip gaji lintas service. [Kasus pembatalan tiket](../../e2e/human_resource/48-hrservice-mutation-cancel-service-request.md)
