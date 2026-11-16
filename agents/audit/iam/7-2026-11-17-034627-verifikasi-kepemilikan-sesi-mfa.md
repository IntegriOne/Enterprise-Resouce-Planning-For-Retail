# Audit lanjutan alur step-up MFA IAM

Tanggal: 17 November 2026 WIB. Cakupan audit menyeluruh enam domain IAM dan 81 RPC telah dicatat pada audit nomor 4–6. Pemeriksaan lanjutan ini menelusuri `AuthHandler.VerifySessionMfa` → `user.Service.VerifySessionMFA` → `session.Service` → repository sesi.

## Temuan dan perbaikan

Pemeriksaan handler semula membandingkan `session_id` permintaan hanya jika metadata sesi pemanggil tidak kosong. Konteks yang ditandai terautentikasi namun tanpa ID sesi dapat melewati pemeriksaan ini. Handler kini mewajibkan kedua ID sesi terisi dan sama. Service sesi menyediakan `MarkUserMFAVerified` yang memeriksa pemilik sesi sebelum mutasi; service user juga memverifikasi pemilik hasil mutasi di dalam transaksi sebagai pertahanan tambahan.

Tes handler membuktikan metadata sesi kosong ditolak. Tes service membuktikan sesi milik pengguna lain ditolak tanpa berubah, dan sesi milik pengguna yang benar dapat melalui step-up MFA. `go fmt ./...`, `go vet ./...`, dan `go test -race -count=1 ./...` lulus pada modul IAM.

## Batas bukti yang masih terbuka

- Scope non-global pada assignment role belum dibuktikan terhadap objek Enterprise Foundation.
- Pengiriman email pascacommit belum memiliki outbox tahan gangguan.
- Balapan MFA, rollback grant/revoke, dan propagasi revisi Redis melalui seluruh jalur gRPC belum memiliki matriks PostgreSQL/Redis nyata yang lengkap.
- Fixture izin Gateway lama dan pengujian E2E tanpa pengecualian tetap berada pada pekerjaan pemilik Gateway.

Dengan demikian, seluruh domain masuk audit alur, tetapi seluruh cabang bisnis dan integrasi nyata belum boleh diklaim selesai.
