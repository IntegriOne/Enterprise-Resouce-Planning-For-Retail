# Regresi dan startup HRM terkini

- `GOWORK=off go test -p 1 -tags=integration -race -count=1 ./...` lulus dengan PostgreSQL disposable. Ada 68 paket tes lulus, termasuk handler 22 domain bisnis. Renderer PDF dijalankan dengan font DejaVu Sans.
- `GOWORK=off go build ./...`, `GOWORK=off go vet ./...`, `buf lint`, dan pemeriksaan whitespace pada kontrak, kode gRPC dan handoff lulus.
- Binary checkout menerima `SERVING` melalui mTLS di kontainer terisolasi. SHA-256 baseline migrasi PostgreSQL yang tercatat cocok dengan checkout.
- Rekonsiliasi dokumentasi dengan proto, handler, service, dan katalog izin mencocokkan 1.010 RPC, terdiri dari 669 Query dan 341 Mutation. [Inventaris](../../grpc/hrm/1-inventaris-operasi-grpc-hrm.md)
- Bukti pengujian lokal menggunakan fixture untuk alur yang memerlukan verifikasi lintas service. Handoff mencatat operasi yang menunggu adapter produksi dan katalog IAM/Gateway yang resmi. [Learning](../../grpc/hrm/11-operasi-learning.md) · [Finance](../../grpc/hrm/25-kebutuhan-integrasi-finance-payroll.md) · [DRM](../../grpc/hrm/24-kebutuhan-integrasi-drm.md)
