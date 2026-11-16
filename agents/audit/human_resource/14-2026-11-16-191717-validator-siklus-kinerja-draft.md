# Audit validator siklus kinerja draft

Validator produksi `hr_performance_cycles` kini terdaftar di wiring HRM. Create/update/delete hanya menerima status draft; kode siklus diperiksa unik dan field peluncuran tidak dapat diisi melalui CRUD. Update maupun delete ditolak jika goal atau review sudah mereferensikan siklus. Perubahan ini tidak membuka transisi status `open` atau fase penilaian lain.

Tes unit mencakup duplikasi, perubahan draft, referensi goal/review, serta penolakan perubahan status. Tes integrasi `bufconn` → gRPC handler → service → repository → PostgreSQL `_test` membuktikan create, finder berhalaman, update draft, penolakan duplikasi, penolakan field peluncur dari klien dan penolakan update/delete setelah goal dibuat. Dua event outbox muncul hanya dari mutasi berhasil. `go fmt`, `go vet`, `go test -race`, `go build`, Docker build dan health check lulus.

Batas: policy tes adalah tiruan dengan scope global. Ini belum membuktikan grant IAM, scope parsial, HMAC/mTLS runtime atau Gateway. Goal/review CRUD masih fail-closed sampai validator parent dan workflow disiapkan. Total validator bisnis kini 13 dari 78 resource mutasi; 65 lainnya tetap ditolak.
