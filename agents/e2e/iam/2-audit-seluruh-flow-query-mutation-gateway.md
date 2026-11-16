# E2E Query/Mutation IAM setelah audit seluruh flow

Tanggal: 17 November 2026 WIB. IAM baru dibangun dan direstart secara terisolasi; status Docker `healthy`.

| Case | Hasil | Bukti |
| --- | --- | --- |
| Login, browser session dan CSRF melalui Gateway | Lulus | Termasuk pada matriks Query/Mutation di bawah; `GetSessionCsrf` kini menuntut sesi caller yang sama |
| User, identity, auth/MFA, session, role, group, permission Query/Mutation selain filter resource lama | Lulus | `go test -tags=e2e ./internal/e2e -run '^Test(Query|Mutation)' -skip '^TestQueryPermissions$' -count=1 -timeout=10m` pada Gateway: `ok`, 37,063 detik |
| `TestQueryPermissions/FilterByResourceSensitiveAndPagination` | Belum lulus | Fixture Gateway masih mengirim `resource="users"`; katalog IAM sekarang menyimpan `iam.users` |

Matriks ini berjalan pada HTTPS `/api/graphql` nyata dengan Gateway dan IAM development. Pengujian tidak membuktikan validitas target scope Enterprise Foundation ataupun fault injection email/DB.
