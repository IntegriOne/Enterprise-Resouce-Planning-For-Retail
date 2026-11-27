# Regresi HRM lintas management

## Pemeriksaan checkout

| Modul | `go test ./...` | `go vet ./...` |
| --- | --- | --- |
| Human Resource | Lulus, 52 paket dengan tes | Lulus |
| IAM | Lulus, 26 paket dengan tes | Lulus |
| Enterprise Foundation | Lulus, 21 paket dengan tes | Lulus |
| Gateway | Lulus, 9 paket dengan tes | Lulus |

Pemeriksaan modul menggunakan checkout aktif. Hitungan paket mengecualikan paket tanpa berkas tes. `git diff --check` lulus.

## Bukti operasi

[Matriks Gateway](../../grpc/hrm/26-status-wiring-gateway-hrm.md) memuat 52 Query dan 61 Mutation domain. Matriks HTTPS mencatat 102 jalur positif dan 11 operasi yang masih memerlukan bukti jalur positif dengan dependency runtime. Referensi Finance dan DRM tetap dalam handoff pemilik. Tes paket tidak membuktikan wiring produksi keduanya.

Regresi terpisah membuktikan provisioning HTTPS melalui IAM dan EF serta terminasi otomatis dan clearance IAM melalui mTLS. [Provisioning serentak](../../e2e/gateway/136-iam-mutation-create-employee-serentak.md) · [Terminasi otomatis](../../e2e/gateway/182-hrm-mutation-terminasi-pencabutan-sesi-iam.md)

Pemeriksa SDL dan guard `check_hrm_schema_guards.py` lulus pada HTTPS development untuk 52 dari 52 Query dan 61 dari 61 Mutation. Pemeriksa juga mencatat 24 tipe Node dan 24 koneksi Relay. Pemeriksaan ini menggunakan permintaan tanpa sesi atau bukti CSRF sesuai skenario guard.

## Integrasi gRPC dan PostgreSQL

Perintah `go test -race -v -count=1 -tags integration -p 1 ./...` lulus di modul HRM. Database baru berakhiran `_test` dipakai dengan opt-in disposable, lalu dibersihkan. Renderer slip memakai DejaVu Sans. Output mencatat 71 paket dengan tes, 294 tes induk lulus, dan tanpa tes berstatus skip.
