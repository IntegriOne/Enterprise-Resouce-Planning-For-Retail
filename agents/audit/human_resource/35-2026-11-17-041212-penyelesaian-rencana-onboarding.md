# Audit command penyelesaian rencana onboarding HRM

Tanggal: 17 November 2026 WIB. Cakupan: `hr_onboarding_plans` dan `hr_onboarding_tasks` dari domain, repository, service, policy/validator, proto, handler gRPC, dan Docker.

## Perubahan

- `CompleteOnboardingPlan` memakai command transaksional pada plan yang dikunci. Actor harus manusia, plan harus `in_progress` dan memiliki metadata mulai. Minimal satu tugas harus ada; semua tugas wajib harus mempunyai `completed_at`. Status, waktu dan principal penyelesaian berasal dari server. Mutasi dan outbox ditulis dalam transaksi yang sama; pengulangan ditolak.
- Validator plan menerima hanya transisi `in_progress` → `completed` dengan kolom lain tetap. Perubahan atau penghapusan plan yang sudah dimulai tetap ditolak. Pemeriksaan tugas diulang dalam transaksi validator.
- RPC gRPC dan permission sensitif `human_resource.onboarding_plans.complete` ditambahkan. Katalog aktif sekarang 326 kode dan 410 pemetaan operasi.

## Bukti

- `go fmt ./...`, `go vet ./...`, `go build ./...`, dan `go test -race -count=1 ./...` pada modul HRM lulus.
- Tes validator dan service mock membuktikan penolakan plan draft, tugas wajib tertunda, perubahan kolom lain, penghapusan plan berjalan, serta pengulangan tanpa outbox tambahan.
- Tes gRPC `bufconn` → PostgreSQL disposable `_test` dengan race detector membuktikan penolakan sebelum tugas wajib selesai, keberhasilan sesudah fixture tugas selesai, metadata actor/waktu dan satu outbox tambahan. Database uji dihapus setelah tes.
- `go generate ./contract` dan `docker compose build human-resource` lulus. `docker compose up -d --no-deps --wait human-resource` menghasilkan container `healthy` pada port privat `127.0.0.1:9003`.

## Batas bukti

Fixture PostgreSQL menulis tanda selesai dan UUID bukti langsung untuk menguji command plan. **Ini bukan bukti bahwa tugas diselesaikan melalui layanan DRM.** `CompleteOnboardingTask` belum ada; kontrak DRM versi spesifik dan pembuktian PIC belum tersedia. Grant IAM dan E2E Gateway positif belum diuji. Penyelesaian plan tidak membuat role IAM.
