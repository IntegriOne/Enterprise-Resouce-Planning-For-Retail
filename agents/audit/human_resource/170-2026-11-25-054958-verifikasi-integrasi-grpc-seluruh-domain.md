# Verifikasi integrasi gRPC seluruh domain HRM

Seluruh 22 paket handler yang memiliki tes integrasi PostgreSQL dijalankan dengan `go test -race -tags integration -count=1 -v`. Setiap paket memakai database PostgreSQL sementara tersendiri. Runner selesai dengan exit code 0.

Sebanyak 38 tes tingkat atas lulus, termasuk 30 tes gRPC–PostgreSQL, tanpa skip atau kegagalan. Domain yang diperiksa meliputi access request, analytics, benefits, career, compensation, competency, employee relations, engagement, HR service, learning, leave, offboarding, onboarding, payroll, performance, personnel, policy, promotion, recruitment, succession, talent, dan workforce planning.

Cakupan ini membuktikan skenario yang diuji di tiap domain. IAM dan dependency pemilik bukti pada tes tertentu memakai adapter tiruan. Integrasi produksi dan pengujian dari `/api/graphql` masih terbuka. Kebutuhan Finance dicatat untuk sinkronisasi. [Inventaris Gateway](../../grpc/hrm/1-inventaris-operasi-grpc-hrm.md) · [Payroll](../../grpc/hrm/15-operasi-payroll.md)
