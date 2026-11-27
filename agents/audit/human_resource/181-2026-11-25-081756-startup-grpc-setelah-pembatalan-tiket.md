# Startup gRPC setelah pembatalan tiket

Binary HRM dari checkout terbaru dibangun dengan `GOWORK=off CGO_ENABLED=0 go build ./cmd`. Kontainer sementara menjalankan binary dengan listener terisolasi, relay outbox standby, Redis TLS lokal, dan database PostgreSQL disposable. Health check menerima status `SERVING` melalui mTLS. SHA-256 migrasi baseline 000001 yang tercatat di database cocok dengan file checkout. Kontainer serta database sementara dibersihkan setelah pengujian.

Bukti startup melengkapi registrasi RPC, regresi 22 domain, dan tes pembatalan tiket paralel. Verifikasi lintas Gateway dan grant IAM produksi serta koneksi ke Finance dan DRM memerlukan runtime terkait. [Inventaris HRM](../../grpc/hrm/1-inventaris-operasi-grpc-hrm.md) · [Regresi pembatalan](../../e2e/human_resource/48-hrservice-mutation-cancel-service-request.md)
