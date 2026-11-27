# Regresi HRM setelah penguatan workflow

Regresi seluruh modul menjalankan `go test -v -p 1 -tags=integration -race -count=1 ./...` dengan PostgreSQL sementara dan Redis TLS terisolasi. Sebanyak 67 paket dan 276 tes tingkat atas lulus tanpa skip. `go vet -tags=integration ./...` dan `go build ./...` juga lulus.

Pemeriksaan mencakup pergantian penempatan primer, rollback promosi, nominasi suksesi, benturan panelis lintas lamaran, waktu penyelesaian asesmen, serta keterikatan bukti dan kronologi tugas. Transfer umum atomik, persetujuan suksesi, lifecycle wawancara, verifikasi DRM dan clearance Finance aktual, serta E2E IAM–Gateway masih terbuka.
