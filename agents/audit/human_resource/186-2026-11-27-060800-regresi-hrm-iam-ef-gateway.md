# Regresi HRM, IAM, Enterprise Foundation, dan Gateway

`go test -count=1 ./...` dan `go vet ./...` lulus di keempat modul. HRM juga lulus `go test -race -p 1 -count=1 -tags=integration ./...` dengan PostgreSQL Docker pada database `hrm_disposable_*_test`, memakai opt-in suite dan dibersihkan setelah selesai.

Percobaan awal menemukan fixture pengajuan akun tanpa status hasil pada peristiwa perekrutan. Fixture kini mengikuti transisi dari command aktivasi dan memverifikasi bahwa server memilih peristiwa tersebut walau ID tidak dikirim. Percobaan awal juga menunjukkan suite succession dan workforce planning memerlukan nama database dan opt-in disposable. Suite terarah serta regresi penuh lulus setelah runner memakai konfigurasi itu.

Tes basis data membuktikan perilaku service dan gRPC pada cakupan suite. Bukti pengujian melalui sesi IAM dan HTTPS dicatat pada [matriks Gateway](../../grpc/hrm/26-status-wiring-gateway-hrm.md).
