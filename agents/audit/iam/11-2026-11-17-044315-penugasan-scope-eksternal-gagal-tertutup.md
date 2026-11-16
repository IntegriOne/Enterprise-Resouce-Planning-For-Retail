# Audit grant scope eksternal IAM

Tanggal: 17 November 2026 WIB. Audit sebelumnya menemukan penugasan role/grup dapat menerima UUID `legal_entity`, `organization_unit`, `facility`, `supplier_party`, atau `vendor_party` tanpa pembuktian objek pemilik. Ini bisa membuat grant berscope tidak sah tampak efektif.

## Perbaikan

- `domain/service.RequireScopeOwnerVerification` mengizinkan hanya scope `global` dan `self` yang dapat diverifikasi lokal oleh IAM. Scope eksternal valid secara format mengembalikan `iam.scope.owner_verification_unavailable` dengan kind `Unavailable` sebelum transaksi dan INSERT assignment.
- `role.Service.AssignToUser` dan `group.Service.AssignRole` menerapkan guard yang sama. Aturan entity tetap mengikat `self` ke pengguna yang bersangkutan dan menolak `self` pada role grup.
- Tes domain, service, dan bufconn membuktikan `global` tetap berjalan, scope eksternal ditolak, serta respons gRPC berkode `Unavailable`.
- Kebutuhan kontrak owner dicatat di `agents/notes/enterprise_foundation/1-iam.md`; `facility` memerlukan keputusan kepemilikan bersama WMS. IAM tidak membuat lookup lintas database.

## Batas

Grant scope eksternal sengaja belum dapat dibuat hingga kontrak verifikasi pemilik tersedia. Ini adalah status gagal tertutup, bukan fitur scope yang selesai. Email outbox, pagination enam list relasi, dan E2E semua alur IAM tetap terbuka sesuai audit nomor 10.

## Verifikasi

`go fmt ./...`, `go vet ./...`, `go build ./...`, dan `go test -race -count=1 ./...` lulus pada modul IAM. `docker compose build iam` dan `docker compose up -d --no-deps --wait iam` lulus; container IAM healthy.
