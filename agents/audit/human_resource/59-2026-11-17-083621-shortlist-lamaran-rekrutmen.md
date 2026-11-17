# Audit shortlist lamaran rekrutmen HRM

## Temuan dan perubahan

`ShortlistApplication` menambah jalur RPC, command sensitif `.advance`, dan guard yang hanya mengizinkan recruiter terpasang memindahkan lamaran `screening` ke `shortlisted` dengan alasan manual. Riwayat tahap menyimpan actor dan waktu server; update lamaran serta outbox berada dalam transaksi yang sama. Skor dan versi scorecard tetap kosong karena belum ada rubrik dan bukti evaluasi yang diverifikasi. Penolakan dari `shortlisted` juga dibuka agar lamaran dapat berakhir tanpa memalsukan tahap berikutnya.

Tes PostgreSQL menemukan bahwa constraint `ck_hr_application_rejected` pada migrasi `000004` menerima `current_stage_code=NULL`: ekspresi `CHECK` bernilai `UNKNOWN`. Migrasi aditif `000005` memperketat constraint rejected dan menambah constraint shortlisted dengan `IS NOT NULL` eksplisit. Migrasi lama tetap immutable.

## Bukti pengujian

- `buf lint`/`buf generate`, validasi DBML 86 tabel/9 enum, migrasi `000005` down/up pada database disposable: lulus.
- `go fmt ./...`, `go vet ./...`, `go test -race ./...`, `go build ./...`: lulus pada modul HRM.
- gRPC `bufconn`→PostgreSQL `_test` dengan race: lulus untuk tahap screening→shortlist, larangan actor lain, alasan kosong, retry, riwayat, outbox, SQL `NULL`, dan penolakan setelah shortlist. Database disposable dihapus sesudah tes.
- Docker HRM dibangun ulang dan healthy; database development memiliki migrasi 1–5 serta kedua constraint.

## Batasan

Policy tes memakai actor tiruan. Belum ada grant IAM, E2E Gateway, verifier owner untuk requisition produksi, atau bukti evaluasi screening terverifikasi. Tahap interview/assessment/offer belum aktif. Assertion penutupan requisition pada fixture SQL lama dihapus karena fixture tersebut tidak mempunyai provenance yang diwajibkan workflow.
