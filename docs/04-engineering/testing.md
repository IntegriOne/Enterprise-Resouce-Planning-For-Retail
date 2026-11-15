# Test Strategy, Plan dan Kasus Penerimaan

> **Pedoman engineering.** Ikuti [AGENTS.md](../../AGENTS.md), keputusan [arsitektur](../../agents/arsitektur.md), [ownership](../../agents/boundary.md) dan [scope](../01-product/vision-and-scope.md). Bukti tes terpisah dari rencana tes.

## Dua lapis sesuai AGENTS
**1. Service tests:** unit, mock dependency, race, benchmark terukur; tanpa SQL/Redis produksi. **2. E2E:** alur nyata dengan Gateway HTTP→gRPC→service owner→PostgreSQL `*_test` dan Redis testing jika diperlukan. Mock/SQL recording/bufconn dapat membuktikan unit/API contract tetapi **tidak** menggantikan E2E lintas service nyata.

## Matriks skenario minimum
| ID | Fase | Skenario dan evidence |
|---|---|---|
| TC-01 | Internal | Admin/staf positif dan staff valid tanpa izin; tidak ada row/outbox saat denied |
| TC-02 | Internal | SKU/variant create/update, version conflict, seller mapping |
| TC-03 | Internal | PO 100, receiving 98 dan Finance matching menolak mismatch |
| TC-04 | Internal | Parallel WMS holds >available; tidak oversell atau membukukan stok palsu |
| TC-05 | Supplier | PIC A tak bisa baca PO milik B via list, count, Node, export |
| TC-06 | Vendor | Scoped contract expiry/revoke; tidak memperoleh role supplier otomatis |
| TC-07 | Marketplace | Webhook bad signature/replay rejected, raw payload tak mengandung secret |
| TC-08 | Marketplace | 100 duplicate/reordered events → satu efek OMS per idempotency key |
| TC-09 | Marketplace | Timeout after OMS accept → caller lookup key sama, bukan order kedua |
| TC-10 | Analytics | Duplicate order, partial refund dan late event → metrics owner cocok |
| TC-11 | Buyer F4 | Guest A tidak melihat cart B, hanya katalog LIVE untuk public |
| TC-12 | Buyer F4 | Quote changed/expired → reject digest lama; OMS one-order |
| TC-13 | Security | 30m idle/8h absolute/30s freshness, effective revoke, Redis/IAM unavailable fail closed |
| TC-14 | Ops | Outbox crash/retry dan DB restore tanpa resurrect unauthorized state |

## Bukti tes
Kolom `test ID | requirement ID | commit SHA | command | test fixture | actual result | service/DB asserted | timestamp | tester`. Rencana tes yang belum dijalankan adalah `NOT RUN`; jika provider sandbox/API official tidak tersedia tulis `BLOCKED` bukan “E2E passed”.

Standar pelaporan per management mengikuti [AGENTS.md](../../AGENTS.md):
- **Laporan Benchmark Service:** `agents/benchmark/{management}/{id}-{benchmark-title}.md` (termasuk metrik `-benchmem`, `ns/op`, `B/op`, `allocs/op`, konfigurasi dataset `mock_db`, dan batasan klaim mock).
- **Laporan E2E (Query/Mutation & Cases):** `agents/e2e/{management}/{id}-{e2e-title}-{query|mutation}-<operation>.md` — template `agents/e2e/{management}/{id}-{e2e-title-{Query***/Mutation***}}.md` — merinci seluruh *cases* (Skenario A/B/C TestAB, CSRF/Origin, error/edge cases, konkurensi/idempotensi, serta verifikasi state & *zero-write* DB `*_test`).

## UAT
Owner business dan PIC pilot menandatangani alur internal, vendor/supplier, marketplace dan akhirnya pembeli secara terpisah. QA tidak mengesahkan deployment dari status test package saja. Penolakan bisnis berbeda dari dependency error; revisi tes dan bug harus mencatat source owner.
