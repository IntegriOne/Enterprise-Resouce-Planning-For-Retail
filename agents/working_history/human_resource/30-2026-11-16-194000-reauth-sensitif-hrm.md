# Kesegaran re-autentikasi sensitif HRM

Policy HRM kini memeriksa waktu autentikasi efektif sesi IAM pada `IssuedAtUnix` untuk semua permission yang sensitif menurut grant IAM atau katalog lokal. MFA dan umur maksimum 10 menit wajib; timestamp hilang/masa depan ditolak. Tes tabel mencakup grant sensitif, command sensitif, sesi lama, MFA hilang, dan read biasa.

Verifikasi: `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, `go build ./...`, `docker compose build human-resource`, dan `docker compose up -d --wait human-resource` lulus; container healthy. Tidak ada grant IAM baru atau perubahan shared authsession. Scope parsial dan E2E IAM/Gateway tetap pekerjaan lanjutan.
