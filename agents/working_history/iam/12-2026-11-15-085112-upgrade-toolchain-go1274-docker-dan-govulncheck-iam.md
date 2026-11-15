# Upgrade Toolchain Go `1.27.4` & Rebuild Docker IAM

- **Waktu:** `2026-11-15 08:51 WIB`
- **Cakupan:** [`backend/go.work`](file:///home/muriarfad/Documents/projects/erp-retail/backend/go.work), [`shared/{authcontext,authsession,authz}/go.mod`](file:///home/muriarfad/Documents/projects/erp-retail/backend/shared/authsession/go.mod), [`services/iam/go.mod`](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/go.mod), [`services/iam/Dockerfile`](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/Dockerfile).
- **Perubahan:** Meningkatkan pin toolchain ke `go1.27.4` beserta SHA256 resmi `amd64`/`arm64` di Dockerfile agar lingkungan Docker/produksi dan lokal identik.
- **Verifikasi:**
  - `go fmt`, `go vet`, `GOWORK=off go test -race -count=1 ./...` (`go1.27.4`) `PASS` di `shared/{authcontext,authsession,authz}` (`1.005s–1.007s`) dan seluruh `20` package bertest `services/iam`.
  - `govulncheck` (`go1.27.4`): `0` kerentanan pada kode & package terimpor (hanya tersisa 1 peringatan modul non-impor `GO-2026-5932` pada `golang.org/x/crypto/openpgp`).
  - Image `integrione-retail-iam:dev` dibangun (`34.3s`), `/usr/local/bin/iam-server` terverifikasi `go1.27.4`, seluruh `8` container `healthy`, E2E HTTPS `71` test / `140` subtest `PASS`, dan re-index `codebase-memory-mcp` berstatus `indexed`.
