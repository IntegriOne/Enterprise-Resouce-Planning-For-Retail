# Audit transisi mulai onboarding HRM

Tanggal: 17 November 2026 WIB. `StartOnboardingPlan` sekarang merupakan command dan RPC terpisah. Permission `human_resource.onboarding_plans.start` sensitif, sehingga membutuhkan grant, MFA, dan reautentikasi yang sah. Service mengunci plan, memeriksa actor manusia dan status draf, tanggal efektif, serta keberadaan tugas. Integrity memeriksa worker dan membatasi perubahan hanya pada status beserta metadata mulai. Penulisan plan, audit actor/waktu server, dan outbox berada dalam transaksi yang sama; pengulangan ditolak.

Baseline `000001` menambahkan metadata mulai/selesai plan dan pelaku penyelesaian task, dengan CHECK untuk kombinasi status, waktu, actor, dan bukti. Perubahan dibuat lewat DBML dan generator resmi. Database HRM development dibangun ulang **setelah seluruh tabel bisnis diperiksa kosong**; tabel migrasi dan kolom baru terverifikasi, container kembali healthy.

## Bukti

- `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, `go build ./...`: lulus pada modul HRM.
- `bash backend/services/human_resource/db/test/verify_schema.sh`: baseline up, kasus integritas termasuk penolakan metadata onboarding yang tidak lengkap, down, dan up lulus.
- Tes integrasi `TestOnboardingDraftRPCWithPostgreSQL` dengan tag `integration` dan race detector: lulus pada database PostgreSQL `_test` sementara; membuktikan start tanpa tugas ditolak, start sah, respons audit, pengulangan ditolak, dan jumlah outbox.
- `docker compose build --quiet human-resource` dan `docker compose up -d --wait --no-deps human-resource`: lulus, container healthy.

## Batas

Command penyelesaian tugas/rencana, validasi versi bukti DRM, relasi PIC selain worker, grant IAM nyata, serta E2E positif lewat Gateway belum tersedia. Direktori runtime `backend/services/document_records_management/` tidak ada pada checkout ini, sehingga UUID bukti dari klien belum dapat dibuktikan ke owner DRM dan penyelesaian tugas harus tetap fail-closed. Tes integrasi memakai policy tiruan, sehingga belum membuktikan otorisasi lintas management pada runtime.
