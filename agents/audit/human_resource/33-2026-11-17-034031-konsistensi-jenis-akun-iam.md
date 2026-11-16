# Audit konsistensi jenis akun IAM pada policy HRM

Policy HRM sebelumnya membandingkan user ID, session ID, auth epoch, revisi izin dan recovery epoch, tetapi tidak membandingkan `user_kind`. IAM memproyeksikan jenis akun ke token sesi dan snapshot; Gateway mengirim jenis dari token melalui konteks gRPC. Policy kini mewajibkan ketiganya sama dan bernilai `employee` sebelum evaluasi permission. Nilai kosong, mismatch, serta kategori customer/vendor/supplier/system ditolak secara fail-closed.

Tes unit/race policy meliputi employee valid, jenis caller/token/snapshot yang berbeda, jenis caller kosong, serta customer konsisten yang tetap ditolak. `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, dan `go build ./...` lulus pada modul HRM. Build Docker lulus dan container HRM kembali `healthy`. Tes menggunakan proyeksi tiruan; belum ada E2E positif IAM→Gateway→HRM atau dukungan actor workload HRM.
