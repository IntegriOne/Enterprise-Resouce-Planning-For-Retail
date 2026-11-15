# Konteks Aktif GraphQL Gateway (`backend/gateway`)

## Apa yang sudah dikerjakan

- **Penegakan Modul `*self`, Auto-Assignment pada Pembuatan Akun, Anti-IDOR, & Penjagaan `PENDING`/Unverified:**
  - Menegakkan `RequireSelfPermission` dan `RequireSelfSensitive` pada [authorizer.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/middleware/authorizer.go) serta helper `requireSelfOrPerm` pada [adapter.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/adapter/iam/adapter.go) untuk operasi mandiri (`iam.users.{readself,updateself,deleteself}`, `iam.sessions.{readself,deleteself}`, `ScopeSelf{ID: userID}`).
  - Memastikan melalui [pending_self_guard_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/middleware/pending_self_guard_test.go) dan [iam_pending_self_guard_e2e_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/iam_pending_self_guard_e2e_test.go) bahwa seluruh `5` alur pembuatan akun (`createUser` publik & admin, `createEmployeeUser`, `createVendorUser`, `createSupplierUser`) otomatis memperoleh role `customer`/`iam_selfservice` (`SELF`), namun tetap diblokir dari login maupun seluruh query/mutation `*self` selama berstatus `PENDING` atau identitas belum terverifikasi, dan baru terbuka setelah `verifyEmail` + status `ACTIVE`.
  - Memastikan jalur lintas-pengguna (`node`, `user`, nested resolvers, `sessions`, `revokeSession`) mewajibkan izin staf IAM beserta *zero-oracle* dan pencegahan eskalasi delegasi di [rbac_adapter.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/adapter/iam/rbac_adapter.go).
- **Verifikasi Penuh:**
  - `go fmt`, `go vet -tags=e2e ./...`, dan `GOWORK=off go test -race -count=1 ./internal/...` lulus pada seluruh `6` package bertest `backend/gateway` (termasuk [pending_self_guard_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/middleware/pending_self_guard_test.go)).
  - Suite E2E HTTPS lengkap pada 8 container Compose sehat lulus **`76`** top-level test, **`155`** subtest (`148` level-1 + `7` level-2), **`0` fail**, **`0` skip** dalam `68.048s` (laporan terbaru: [9-iam-pending-and-unverified-self-service-guard-https-e2e.md](file:///home/muriarfad/Documents/projects/erp-retail/agents/e2e/gateway/9-iam-pending-and-unverified-self-service-guard-https-e2e.md)).

## Apa yang belum dikerjakan

- Wiring skema dan adapter domain non-IAM menunggu kesiapan service masing-masing management owner.

## Apa yang harus dikerjakan selanjutanya

1. Pertahankan pemisahan ketat permission `*self` (`ScopeSelf`) vs permission lintas-pengguna, auto-assignment peran baseline disertai blokir eksekusi selama `PENDING`/unverified, `MutationResult`, *auth error masking*, batas AST GraphQL, dan aturan *pre-auth CSRF* pada wiring domain berikutnya di Zona 1–3 Gateway.
