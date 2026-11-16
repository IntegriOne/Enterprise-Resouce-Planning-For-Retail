# Audit worker preboarding HRM

## Temuan

RPC dan repository `hr_workers` sudah tersedia, tetapi Integrity tidak mendaftarkan validator. Mutasi worker gagal tertutup; tes domain lain memakai fixture SQL. Worker adalah sumber hubungan kerja, sedangkan person party dan legal entity milik Foundation, identitas dan grant milik IAM.

## Perbaikan

- Validator menerima create `preboarding` dan koreksi tanggal hire/service sebelum worker dipakai kontrak, assignment atau onboarding. Nomor unik, identitas, kategori, legal entity, pembuat dan timestamp buat dijaga.
- Status aktif, link IAM, primary assignment, terminasi dan delete tidak dapat dilakukan lewat CRUD. Delete dilarang agar riwayat hubungan kerja tidak hilang.
- `WorkerReferenceVerifier` memerlukan bukti party orang serta legal entity Foundation. Wiring produksi menginjeksi nil dan create/update tetap `Unavailable` sampai adapter owner tersedia. Tidak ada write IAM atau Foundation dari HRM.

## Bukti dan batas

Tes unit/mock membuktikan jalur sah, perubahan status/link, duplikasi, child, owner outage/nil dan cancellation. Tes gRPC `bufconn` → PostgreSQL disposable `_test` dengan policy/verifier tiruan membuktikan create, update, penolakan delete dan dua outbox event. `go test -race ./...`, `go vet`, `go build`, build Docker, dan health container HRM lulus.

Belum ada adapter Foundation produksi, bukti party/entitas nyata, workflow aktivasi/link IAM/assignment, atau E2E Gateway/IAM. Keunikan hubungan kerja aktif per party/legal entity dan konkurensi create juga belum dibuktikan; itu harus diselesaikan sebelum adapter Foundation diaktifkan. Validator terdaftar 38/78; pendaftaran tidak membuktikan operasi produksi aktif.
