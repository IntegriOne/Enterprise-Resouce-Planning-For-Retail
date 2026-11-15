# Switch Backend IAM Time Handling ke Waktu Lokal OS

- Mengganti seluruh penggunaan `time.Now().UTC()`, `.UTC()`, dan `time.UTC` pada backend Go IAM menjadi waktu lokal OS (`time.Now()` / `time.Local`).
- Mengubah `gorm.Config.NowFunc` di `backend/services/iam/db/init.go` ke `time.Now()`.
- Mengubah `SystemClock.Now()` di `backend/services/iam/domain/service/ports.go` ke `time.Now()`.
- Mengganti normalisasi timestamp di 8 file `backend/services/iam/domain/entity/*.go` ke `normalizeLocalTime` (`time.Local` sambil menjaga nilai zero) serta memutakhirkan seluruh 9 file test `domain/*/test/*_test.go`.
- Verifikasi: `rg -n -i 'utc' backend` nol hasil; `go fmt ./...`, `go vet ./...`, `go build ./...`, dan `go test -race -count=1 ./...` lulus untuk `domain/*` dan unit test `db/` (`TestEnsureAndMigrateDevAndTestDatabases` di-skip tanpa DSN live PostgreSQL).
