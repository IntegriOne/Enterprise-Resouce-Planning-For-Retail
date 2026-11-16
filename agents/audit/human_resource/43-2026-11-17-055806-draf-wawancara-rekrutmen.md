# Audit draf wawancara rekrutmen HRM

Tanggal: 17 November 2026 (WIB). Scope: `hr_interviews` pada HRM.

## Temuan dan perbaikan

RPC dan repository wawancara sudah ada, tetapi `Integrity` belum mendaftarkan validator sehingga seluruh mutasi tertutup. Validator draf sekarang mengunci application, candidate, dan requisition dalam transaksi; menerima status lamaran yang belum terminal, kandidat aktif dengan retensi sah, serta lowongan terbuka yang disetujui. Ronde/jenis unik dan identitas pembuat/parent tidak dapat diganti. Jadwal memerlukan start/end masa depan dan zona waktu IANA; tanggal selesai tidak boleh melewati retensi kandidat. Update/delete draf ditolak setelah panelis atau feedback terhubung. Referensi lokasi dan rapat luar tetap `Unavailable`; status selain `planned` tidak dapat ditulis melalui CRUD.

## Bukti

- `go fmt ./...`, `go test -race -count=1 ./...`, `go vet ./...`, `go build ./...`, dan `git diff --check` lulus pada modul HRM.
- Tes gRPC `bufconn` → PostgreSQL disposable `_test` lulus untuk create, duplikasi, zona waktu tidak valid, retensi, referensi eksternal, finder/pagination, reschedule, penolakan perubahan status, panelis tanpa validator, pembekuan saat panelis terhubung, delete, serta tiga event outbox.
- `docker compose build --quiet human-resource` dan `docker compose up -d --wait --no-deps human-resource` lulus; container healthy.
- Tes menggunakan policy tiruan dan requisition `open` fixture SQL. Belum ada grant IAM/Gateway, workflow approval requisition, atau undangan kandidat.

## Sisa risiko

Hanya draf wawancara yang aktif. Transisi status lamaran, jadwal final, panelis terverifikasi, feedback, keputusan, dan E2E lintas owner tetap harus diimplementasikan. Total 36/78 resource mutasi memiliki validator terdaftar; posisi create/update masih fail-closed tanpa verifier Foundation.
