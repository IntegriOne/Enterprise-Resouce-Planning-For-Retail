# Finalisasi Pending Email, Default Locale/Timezone, dan Ketahanan Pool Redis IAM

- Menambahkan konstanta `DefaultUserLocale` (`id-ID`) dan `DefaultUserTimezone` (`Asia/Jakarta`) di [`enums.go`](../../../backend/services/iam/domain/entity/enums.go), menginisialisasinya pada `entity.NewUser`, menjaga nilai non-kosong pada `User.UpdateProfile`, serta menghilangkan duplikasi di [`mappers.go`](../../../backend/services/iam/repository/user/mappers.go).
- Meningkatkan `MaxRetries: 4` serta backoff `10ms–100ms` di [`redis_client.go`](../../../backend/shared/authsession/redis_client.go) agar soket pool usang (`broken pipe`/`EOF`) pasca-restart Redis terkuras otomatis secara *fail-closed*.
- Memperluas unit test di [`user_test.go`](../../../backend/services/iam/domain/entity/test/user_test.go) lalu memverifikasi `go fmt`, `go vet`, `-race`, PostgreSQL 18 `iam_test`, dan TLS Redis (`1/1` package `shared/authsession` dan `19/19` package `services/iam` lulus `PASS`, `0` skip/fail).
