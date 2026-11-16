# Audit draf requisition HRM

## Temuan

`hr_requisitions` memiliki RPC CRUD dan entity, tetapi tidak mempunyai validator Integrity. Semua mutasi gagal tertutup; tes lamaran menyisipkan requisition `open` langsung lewat SQL. DBML mewajibkan unit Foundation serta hiring manager, sementara approval dan pembukaan belum mempunyai command.

## Perbaikan

- Validator draf terdaftar pada wiring HRM. Create/update memeriksa manager sebagai worker aktif dengan tautan IAM tunggal, job aktif, posisi opsional yang cocok, tanggal target, nomor unik, dan bukti unit/lokasi dari `PositionReferenceVerifier`.
- Field approval/open/close tidak dapat diisi lewat CRUD. Plan, anggaran Finance, dan recruiter yang belum terverifikasi ditolak. Jenis `confidential`/`evergreen` belum dibuka. Update/delete draf yang sudah dipakai aplikasi atau posting ditolak.
- Registry produksi menginjeksi verifier nil: create/update tetap `Unavailable`; delete draf tanpa child dapat berjalan. Tidak ada akses SQL ke Foundation/IAM atau grant IAM baru.

## Bukti dan batas

Unit mock memeriksa jalur sah, duplikasi, owner unavailable, manager tidak aktif, posisi berbeda unit, anggaran tak terverifikasi, child, dan context cancellation. Tes gRPC `bufconn` → PostgreSQL disposable `_test` dengan policy/verifier tiruan membuktikan create/update/list/delete, larangan membuka melalui CRUD, serta tiga outbox. `go fmt`, `go vet`, `go build`, seluruh `go test -race`, dan Docker HRM healthy lulus.

Belum ada adapter Foundation produksi, bukti mandat IAM hiring manager, approval/open command, atau E2E Gateway/IAM. Worker manager dan job pada tes memakai fixture SQL; `hr_workers` sendiri belum memiliki validator mutasi. Tes lamaran masih memakai requisition `open` fixture SQL. Validator terdaftar 37/78, tetapi pendaftaran tidak berarti operasi produksi sudah tersedia.
