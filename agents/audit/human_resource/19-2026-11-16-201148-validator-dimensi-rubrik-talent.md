# Audit validator dimensi rubrik Talent HRM

## Temuan dan perbaikan

`hr_talent_dimensions` sebelumnya memiliki RPC dan service CRUD, tetapi mutasi ditolak karena tidak ada validator produksi. Parser finalisasi asesmen juga baru memeriksa keberadaan level 1–5, sedangkan bentuk rubrik dapat memuat tingkat tambahan. Validator baru mengunci framework dalam transaksi, memastikan kode dimensi unik dalam framework, memeriksa tepat lima jangkar teks yang tidak kosong, dan menolak create/update/delete bila framework sudah direferensikan siklus review. Pemindahan dimensi ke framework lain ditolak.

## Bukti

Tes unit domain dan validator lulus. Tes `bufconn` → service → repository → PostgreSQL `_test` membuktikan CRUD dimensi sebelum siklus, GetByID, finder dua halaman, kode duplikat, `InvalidArgument` untuk jangkar tidak lengkap, dan penolakan ketiga mutasi setelah siklus dibuat. Fixture framework disisipkan langsung pada database disposable; database tersebut dihapus setelah tes. `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, `go build ./...`, `docker compose build human-resource`, dan `docker compose up -d --wait --no-deps human-resource` lulus; container healthy.

## Batas

Framework belum bisa dibuat lewat RPC bisnis karena bukti versi dokumen DRM belum diverifikasi. Policy integrasi masih tiruan; belum ada E2E grant IAM/Gateway maupun pembuktian mTLS/HMAC pada mutasi ini. Validator produksi kini mencakup 15 dari 78 resource berkode mutasi; sisanya fail-closed.
