# Status Implementasi, Bukti dan Batas Klaim

> **Dokumentasi Integrione · 12 November 2026.** Penetapan owner, desain, implementasi, verifikasi dan deployment adalah status berbeda. Kebijakan normatif mengikuti [AGENTS.md](../../AGENTS.md), [arsitektur keamanan](../../agents/arsitektur.md), dan [batas management](../../agents/boundary.md).

## Kategori status yang wajib
| Label | Boleh diklaim jika |
|---|---|
| `DECIDED` | Owner/otoritas menyetujui kebijakan/scope |
| `DESIGNED` | Dokumen, DBML atau API contract ada |
| `IMPLEMENTED` | Source dan wiring operasional ada |
| `VERIFIED` | Test aktual di scope yang tepat telah dijalankan dan hasilnya tersedia |
| `BLOCKED` | Kontrak, izin, dependency atau tes wajib belum terpenuhi |
| `DEPLOYED` | Versi yang benar benar berjalan di lingkungan target telah diperiksa |

**Bukti tidak boleh dinaikkan otomatis:** unit test ≠ integrasi PostgreSQL; E2E ke stub IAM ≠ IAM nyata; API katalog ≠ aktif di gqlgen; health process ≠ business RPC ready; source event ≠ uang masuk; commit ke main ≠ deploy.

## Checkpoint 12 November 2026 (bukan audit runtime baru)
- **IAM:** sesi opaque, snapshot, scope dan worker tersedia/teruji pada paket tertentu. Gateway memiliki **pilot HTTP opt-in** `GATEWAY_AUTH_OPAQUE_BROWSER_ENABLED`, default false menurut audit IAM. Cutover produksi umum dan E2E browser→IAM→management lengkap belum dibuktikan. Lihat [audit](../../agents/audit/iam_management/3-2026-11-12-164500-integrasi-opaque-csrf-reference-gateway.md).
- **Commerce:** empat operasi Stage A kanal telah dihubungkan; E2E publik GraphQL→gRPC mTLS→PostgreSQL tercatat pada test tertentu. **Akses staf Commerce IAM/Foundation, katalog/grant positif dan fase B–H belum diakui selesai**. Lihat [audit Commerce](../../agents/audit/ecommerce_commerce/3-2026-11-12-165500-e2e-public-dan-gerbang-iam.md).
- **Pricing, OMS, CRM, Customer Service:** direktori dan rancangan tidak membuktikan service bisnis sudah aktif; cek source saat menjalankan tugas.
- **Marketplace:** API partner, skema signature, quota dan izin Shopee/TikTok Shop **belum diverifikasi**.
- **Manufacturing:** roadmap fitur baru PAUSED dan di luar 13 management prioritas. Direktori `backend/services/manufacturing_operations/` tidak ada pada tree `main` yang diperiksa 13 November 2026; referensi implementasi 2026-11-12 pada audit lama bersifat historis, bukan bukti service saat ini.

## Cara memperbarui status
Sertakan `commit SHA`, file/RPC/GraphQL yang dicek, perintah test, nama database `*_test`, aktor/scope, waktu, output dan batas uji. Jika belum ada hasil, tulis `NOT RUN` atau `BLOCKED`. Status pada halaman ini adalah snapshot bertanggal, bukan pengganti source terbaru.

## Keputusan GraphQL-only — 13 November 2026

- `DECIDED`: hanya `/api/graphql` sebagai HTTP API publik.
- `IMPLEMENTED` (lapisan ingress): Caddy menolak `/api/*` lain; liveness
  Gateway berpindah ke Query GraphQL, bukan `/healthz`.
- `BLOCKED / NOT VERIFIED` (browser lifecycle): GraphQL pre-auth CSRF,
  opaque login, session/rotate/logout belum terintegrasi menggantikan seluruh
  REST handler internal dan JWT legacy. Jangan klaim migrasi auth selesai
  tanpa test browser IAM/Redis nyata dan dua actor scope.

## Gateway code milestone — 13 November 2026

- `IMPLEMENTED (source)`: Gateway hanya mendaftarkan `/api/graphql`;
  `/api/auth/csrf`, route browser REST, dan `/healthz` tidak lagi
  didaftarkan walaupun mode opaque diaktifkan.
- `IMPLEMENTED (source)`: bootstrap CSRF **legacy** melalui GET Query GraphQL
  `{ health }` (header `X-CSRF-Token` dan cookie double-submit) untuk
  menghindari endpoint REST sambil mempertahankan login GraphQL historis.
  Bootstrap dibatasi ke origin browser tepercaya; mode opaque tidak menerima
  cookie CSRF legacy tambahan.
- `BLOCKED`: migrasi GraphQL IAM opaque untuk login/session/rotate/logout dan
  Query bootstrap IAM yang sesungguhnya tetap belum ada; akses browser opaque
  end-to-end belum boleh diklaim siap.
- `NOT RUN`: Go build/tests/race dan E2E IAM/Redis/dua Gateway perlu dijalankan
  di checkout lokal dengan toolchain proyek.
- Mode opaque browser menolak cookie JWT legacy dan resolver login/logout tidak
  memintakan sesi JWT legacy ketika opaque dipilih. Fail-closed ini bukan bukti
  GraphQL opaque sudah tersedia.
