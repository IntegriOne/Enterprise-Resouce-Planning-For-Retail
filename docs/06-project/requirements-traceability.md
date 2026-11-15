# Requirements Traceability Matrix

> **Template pengelolaan proyek, status DRAFT/NOT RUN** sampai persetujuan atau pengujian aktual. Rujukan: [Charter](project-charter.md), [Requirements](../01-product/requirements.md), [DoD](../04-engineering/quality-gates.md).

## Kontrak penelusuran
Satu requirement mempunyai ID stabil dan sumber bisnis; implementasi, API, test dan keputusan penerimaan harus dapat ditelusuri tanpa membaca chat awal.

| BRD | PRD | SRS | Domain owner | Test minimal | Status draft |
|---|---|---|---|---|---|
| BR-01 | P-01/P-02 | S-01/S-03 | IAM/EF/PLM | Dua user, SKU varian, scope seller | DESIGNED/PARTIAL |
| BR-01 | P-03 | S-04 | Procurement/WMS/Finance | Partial receipt+matching, row assertion owner | BLOCKED integrated E2E |
| BR-02 | P-05/P-06 | S-01/S-02 | IAM/Procurement/DRM | Supplier A vs B, vendor contract expiry | DESIGNED |
| BR-03 | P-07 | S-05/S-07 | Integration/OMS/WMS | Provider signed webhook, duplicate and lost response | BLOCKED provider |
| BR-04 | P-08 | S-08 | Analytics/OMS/Finance | Dedupe, refund correction, watermark | DESIGNED |
| BR-05 | P-09/P-10 | S-06/S-09 | Commerce/Pricing/OMS/IAM | Guest vs buyer, digest quote, one OMS order | BLOCKED F4 readiness |

## Field per implementasi
`requirement ID | function/API symbol | file path | domain owner | business invariant | permission/scope | test ID | command | SHA | time | PASS/FAIL/BLOCKED | UAT approval`. Jangan merekayasa nama RPC atau hasil tes; ambil dari kode dan test runtime.

## Change impact
Jika schema GraphQL, provider webhook, sumber harga, API OMS, atau policy IAM berubah, identifikasi consumer dan testcase terdampak. Update RFC/ADR/RTM bersama revisi owner, jalankan ulang test; tidak mengklaim backward compatibility hanya dari berhasil compile.
