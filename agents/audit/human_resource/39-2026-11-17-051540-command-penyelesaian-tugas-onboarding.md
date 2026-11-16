# Command penyelesaian tugas onboarding HRM

Tanggal: 17 November 2026 WIB. `CompleteOnboardingTask` kini tersedia sebagai RPC privat dan command aplikasi terpisah dari CRUD. Permission `human_resource.onboarding_tasks.complete` sensitif dan terpisah dari `.update`. Command memerlukan actor manusia yang sama dengan principal pekerja yang ditugaskan, tugas belum selesai, bukti versi dokumen DRM yang terverifikasi dan immutable, serta rencana `in_progress`. Validator mengunci plan dan worker, menjaga semua field tugas selain metadata penyelesaian, dan menolak tugas worker tanpa link IAM yang cocok. Update generik tetap menolak perubahan `completed_at`; pengulangan command ditolak. Update task dan outbox berada dalam satu transaksi.

## Bukti

- Unit/mock menguji verifier tidak tersedia, bukti tidak sah/mutable/versi salah, penyelesaian sah, pengulangan, dan ketiadaan event pada penolakan.
- Tes validator menguji hanya tugas worker pada plan berjalan, tidak pada plan draf atau saat field lain berubah.
- Tes `bufconn` → PostgreSQL disposable `_test` dengan race detector menguji RPC completion, audit actor/waktu/bukti, plan completion sesudah tugas wajib, penolakan pengulangan, dan jumlah outbox. Tes ini memakai policy dan verifier DRM tiruan.
- `go generate ./...` pada kontrak, `go fmt ./...`, `go vet ./...`, `go build ./...`, dan `go test -race -count=1 ./...` lulus. Image `human-resource` dibangun ulang; container sehat setelah restart.

## Batas

Adapter DRM produksi belum tersedia; registry tidak memasang verifier, sehingga RPC produksi gagal tertutup dengan `Unavailable`. Tes tiruan tidak membuktikan keabsahan dokumen nyata, grant IAM, scope non-global, atau E2E Gateway. Tugas PIC selain worker tetap belum dapat diselesaikan sampai relasi principal/mandatnya terverifikasi. Kebutuhan kontrak DRM ada pada `agents/notes/document_records_management/1-human_resource.md`. Lifecycle onboarding belum dapat dinyatakan selesai.
