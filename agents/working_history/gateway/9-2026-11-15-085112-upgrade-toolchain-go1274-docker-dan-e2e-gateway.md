# Upgrade Toolchain Go `1.27.4` & Rebuild Docker Gateway

- **Waktu:** `2026-11-15 08:51 WIB`
- **Cakupan:** [`backend/gateway/go.mod`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/go.mod), [`backend/gateway/Dockerfile`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/Dockerfile).
- **Perubahan:** Meningkatkan `go 1.26.8` menjadi `go 1.27.4` di `go.mod` dan `ARG GO_VERSION=1.27.4` beserta SHA256 resmi `amd64`/`arm64` di `Dockerfile` agar image produksi memakai `go1.27.4`.
- **Verifikasi:**
  - `go fmt`, `go vet`, `GOWORK=off go test -race -count=1 ./...` (`go1.27.4`) `PASS` di seluruh `6` package bertest `backend/gateway`.
  - `govulncheck` (`go1.27.4`): `0` kerentanan pada kode & package terimpor (tersisa 1 peringatan modul non-impor `GO-2026-5932`).
  - Image `integrione-retail-gateway:dev` dibangun (`37.6s`), `/usr/local/bin/gateway-server` terverifikasi `go1.27.4`, `gateway-1` & `gateway-2` berstatus `healthy`, dan suite E2E HTTPS (`71` test top-level, `140` subtest, `0` fail, `0` skip, `57.033s`) seluruhnya `PASS`.
