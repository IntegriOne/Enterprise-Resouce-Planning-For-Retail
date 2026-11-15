# Deployment, RaaS dan Go Workspace

> **Baseline 12 November 2026.** Kebijakan keamanan/izin harus cocok dengan [agents/arsitektur.md](../../agents/arsitektur.md); ownership ditentukan oleh [boundary management](../../agents/boundary.md). Status runtime perlu dibuktikan dari kode/tes terbaru.

## Workspace aktual yang harus menjadi sumber
Workspace `backend/go.work` memakai `go 1.27.0` dan `toolchain go1.27.4` (snapshot 12 Nov 2026); module `backend/gateway/go.mod` mengikat Fiber v3.7.0, gqlgen v0.17.97 dan gRPC v1.85.0. **Verifikasi langsung tiap kali dependency berubah.** Direktori backend yang aktif bernama `backend/` (bukan `back-end/`); Git root juga memuat `frontend/`, `proxy/` dan `docker-compose.yml`.

```text
erp/
├── AGENTS.md
├── agents/                  # kebijakan, boundary, audit, handoff
├── docs/                    # dokumentasi kanonik
├── docker-compose.yml       # orkestrasi root
├── backend/
│   ├── go.work              # workspace Go
│   ├── contracts/           # protobuf dan binding
│   ├── gateway/             # GraphQL / HTTP ingress
│   ├── shared/              # utilitas teknis
│   ├── services/{management}/ # hanya 13 management aktif
│   └── deploy/              # konfigurasi development
├── frontend/                # website/UI
└── proxy/                   # Caddy ingress
```

Setiap management memiliki DB logis/kredensial isolasi; jangan menuntut tiap service memiliki server fisik sendiri. Development Compose menyediakan PostgreSQL/Redis; produksi butuh TLS, vault, backups, recovery dan approval terpisah.

## Model bisnis pelanggan
Baseline arsitektur saat ini **single-company RaaS per installation**. Banyak UMKM sebagai pelanggan jasa **tidak otomatis shared multi-tenancy**. Pilihan dedicated per merchant vs shared tenant membutuhkan ADR baru, costing, breach isolation, IAM realms, migrations, queries, backups dan kontrak legal. Dilarang menambahkan `tenant_id` secara diam-diam atau mencabut realm dari IAM tanpa keputusan.

## Pengembangan dan deployment
Developer commit hanya ke `main`, tidak membuat branch/force-push; kontrol file generated/shared harus dikoordinasikan. Build dari SHA teruji dan rilis produksi hanya setelah UAT, security, backup/restore dan owner sign-off. Fitur yang belum siap diberi gate OFF, bukan fake success. Tag Docker `latest` pada dev dapat berubah; catat versi actual setiap pull.

## Kesiapan operasional
Health/readiness berbeda dari layanan bisnis siap; khusus commerce kanal perlu IAM audience, grant dan Foundation yang benar sebelum staf aktif. Redis failover tidak auto reconnect tanpa recovery gate. Pelanggan/support dapat mengekspor data sesuai kontrak dan retensi.
