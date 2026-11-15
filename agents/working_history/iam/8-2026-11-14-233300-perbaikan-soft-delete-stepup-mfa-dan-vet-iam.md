# 8 — Perbaikan Transaksi SoftDelete, Proyeksi Step-Up MFA, Otorisasi Reauth Mandiri, dan Go Vet IAM

Menyelesaikan temuan integrasi E2E HTTPS IAM–Gateway pada `backend/services/iam`:

- Mengubah urutan transaksi [SoftDelete](../../backend/services/iam/services/user/service.go) agar mencabut sesi & menaikkan `auth_epoch` sebelum `deleted_at` terisi, serta menyediakan fallback snapshot `in-tx` berstatus `disabled` pada [runtime_sync.go](../../backend/services/iam/services/session/runtime_sync.go).
- Memperbarui [projector.go](../../backend/services/iam/infrastructure/redisstore/projector.go) agar `IssuedAtUnix` Redis memakai timestamp efektif `MFAVerifiedAt` saat lebih baru dari `AuthenticatedAt`.
- Mendukung `RequireSensitiveReauth: true` tanpa kode izin RBAC di [access_policy.go](../../backend/services/iam/domain/service/access_policy.go) beserta unit test di [access_policy_test.go](../../backend/services/iam/domain/service/test/access_policy_test.go), dan memperbaiki *self-assignment* di [snapshot_query.go](../../backend/services/iam/repository/user/snapshot_query.go).
- Memverifikasi `go fmt`, `go vet`, `GOWORK=off go build`, dan `go test -race -count=1 ./...` seluruh paket lulus.
