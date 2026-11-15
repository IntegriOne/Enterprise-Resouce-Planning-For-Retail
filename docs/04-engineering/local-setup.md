# Local Development & Workspace

> **Pedoman engineering.** Ikuti [AGENTS.md](../../AGENTS.md), keputusan [arsitektur](../../agents/arsitektur.md), [ownership](../../agents/boundary.md) dan [scope](../01-product/vision-and-scope.md). Bukti tes terpisah dari rencana tes.

## Runtime source
Root Git memuat `backend/`, `frontend/`, `proxy/`, `agents/`, `docs/`, dan `docker-compose.yml`. Workspace Go berada di **`backend/go.work`**; `backend/gateway/`, `backend/services/`, `backend/contracts/`, `backend/shared/` dan `backend/deploy/` bukan folder langsung di root. Gunakan `backend/`, bukan path lama `back-end/`. Workspace `go 1.27.0` dengan `toolchain go1.27.4`; Gateway module mencantumkan Fiber v3.7.0, gqlgen v0.17.97. Periksa `backend/go.work`, `backend/gateway/go.mod` dan `go version` sebelum build.

```bash
git status --short --branch
git fetch origin
git pull --ff-only origin main
cd backend
go version
go work sync
# Masuk ke module yang akan diuji:
cd gateway
go test ./...
```

**Perhatian:** `go work sync` mengubah dependency metadata; hanya jalankan jika memang tugas sinkronisasi dan tinjau diff file shared. E2E tidak boleh memakai database development/production; gunakan DB disposable `*_test` dan Redis test terisolasi sesuai README modul. Jangan menjalankan `FLUSHALL` pada Redis bersama.

## Docker Compose terpadu (root repository)

Root `docker-compose.yml` meng-include `backend/docker-compose.yml` dan
`frontend/docker-compose.yml`. Backend meneruskan ke Docker dependency dan
dua Gateway HTTP. Versi Compose minimal v2.20 diperlukan.

```bash
bash backend/deploy/scripts/setup-local-env.sh
bash backend/deploy/scripts/generate-dev-certs.sh
docker compose config --quiet
docker compose up -d --build --wait
docker compose ps
```

Port lokal: Gateway 5000/5001, frontend 3333/3334, PostgreSQL 6543, Redis
TLS 6379. Port 5002–5014 dicadangkan untuk microservices dan tidak
otomatis menjadi READY. Gateway Compose awal menonaktifkan koneksi protected
sampai IAM mTLS/opaque dan izin domain terverifikasi. Lihat
`backend/deploy/README.md`. Perintah `docker compose down` tanpa
`-v` menjaga volume lama; jangan memakai `down -v`.

## Menjalankan service
Ikuti perintah dan env di `backend/services/{management}/README.md`; contoh `make proto-iam` dan `go run ./cmd/server` hanya berlaku bila target/konfigurasi benar-benar tersedia. Gateway `/api/graphql` memerlukan service target yang hidup, TLS/identity, dan schema yang aktif. `health`/`NOT_SERVING` tidak memberi bukti bisnis siap.

## Data dan credential
Salin `.env.example` lokal jika tersedia, jangan commit file rahasia. Kredensial IAM MFA, DSN Postgres, token Shopee/TikTok dan signing/encryption keys berasal dari secret manager/env yang tidak terlog. Dev Compose memakai image `latest` sesuai instruksi lama; verifikasi image runtime versi sebenarnya setiap update.

## Minimum verifikasi
Perubahan Go: `gofmt -w <files>`, `go vet ./...`, `go test -race ./...` pada scope, `go build ./...`. Kontrak proto: Buf lint/generate. Gateway: gqlgen generate dengan konfigurasi aktif. E2E nyata mengikuti test harness yang memerlukan setup spesifik, bukan otomatis setiap `go test ./...`. Catat apa yang belum dijalankan.
