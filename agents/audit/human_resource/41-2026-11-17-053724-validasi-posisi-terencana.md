# Validasi posisi terencana HRM

Tanggal: 17 November 2026 WIB. Validator `hr_positions` terpasang pada wiring. Create/update hanya menerima status `planned`, job HRM yang aktif pada masa berlaku posisi, parent posisi yang cocok employer dan masa berlaku, serta kode unik. Identitas job/unit/employer/parent tidak boleh berubah lewat CRUD. Update/delete posisi dengan assignment, requisition, suksesi atau child position ditolak. Cost center Finance dan status `active` tidak dapat dipasang melalui CRUD. Create/update posisi mewajibkan step-up.

Referensi legal entity, unit dan lokasi harus dibuktikan oleh `PositionReferenceVerifier`. Runtime belum memiliki adapter Foundation dan menginjeksi nil; create/update tetap gagal tertutup dengan `Unavailable`. Delete draf tanpa referensi lokal tetap dapat diproses. Ini persiapan integrasi, bukan klaim posisi siap dipakai sebagai parent penempatan resmi.

## Bukti

- Unit test memeriksa draf sah dengan verifier tiruan, duplikasi, assignment terhubung, owner gagal, relasi Foundation salah, status aktif, parent berubah, cost center, job nonaktif, employer parent salah dan context cancellation.
- gRPC `bufconn` → PostgreSQL disposable `_test` dengan verifier tiruan dan race detector lulus untuk create/update/list/delete, duplikasi/status aktif dan tiga outbox event.
- `go fmt ./...`, `go vet ./...`, `go build ./...`, `go test -race -count=1 ./...` lulus pada modul HRM. Bukti IAM/Gateway/EF nyata belum ada.
