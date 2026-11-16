# Audit draf rencana suksesi HRM

Tanggal: 17 November 2026 WIB. Cakupan: validator `hr_succession_plans`, policy sensitif, dan wiring HRM.

## Perubahan

- Validator mengunci posisi HRM dan mewajibkan posisi aktif pada tanggal review saat membuat/mengubah plan. Kode+versi harus unik per posisi. ID, posisi, kode dan versi tidak dapat dipindah melalui update.
- CRUD hanya menerima draf `highly_restricted` tanpa approver, waktu approval atau succession risk. Approval dan penilaian risk memerlukan workflow tersendiri. Plan dengan kandidat tidak dapat diubah atau dihapus; plan tanpa kandidat dapat dihapus walau posisi kemudian ditutup.
- Read dan mutasi plan/kandidat ditandai sensitif untuk step-up. Registry produksi memasang validator plan; kandidat tetap fail-closed.

## Bukti

- Tes unit validator mencakup create/update/delete draf, duplikasi, kandidat terhubung, posisi nonaktif/hilang, approval dari klien, perubahan posisi, penurunan kerahasiaan, dan context dibatalkan.
- `go fmt ./...`, `go vet ./...`, `go build ./...`, `go test -race -count=1 ./...` lulus pada modul HRM.
- `docker compose build human-resource` dan `docker compose up -d --no-deps --wait human-resource` lulus; container berstatus healthy.

## Batas bukti

Belum ada uji positif gRPC→PostgreSQL untuk suksesi atau E2E IAM/Gateway. Posisi HRM belum dapat dibuat melalui validator produksi, sehingga jalur positif memerlukan fixture posisi sampai workflow posisinya tersedia. Persetujuan plan, validasi kandidat/readiness dan mandat panel belum diaktifkan. Policy unit tidak membuktikan grant IAM nyata.
