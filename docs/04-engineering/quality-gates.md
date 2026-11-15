# Definition of Done (DoD), QA dan Release Gates

> **Pedoman engineering.** Ikuti [AGENTS.md](../../AGENTS.md), keputusan [arsitektur](../../agents/arsitektur.md), [ownership](../../agents/boundary.md) dan [scope](../01-product/vision-and-scope.md). Bukti tes terpisah dari rencana tes.

## DoD per domain
- Scope/ownership jelas, requirement ID serta operasi Query/Mutation/RPC lengkap.
- Auth permission dan scope disepakati dengan IAM; tidak ada role otomatis; test positif-negatif dan deny no-state.
- Migrasi DBML/GORM owner konsisten, baseline `000001` selama pra-rilis, test migration isolated.
- Business invariant, state transition, CAS/idempotency, timeout/error/rollback/outbox sesuai domain.
- Generator proto/gqlgen resmi; build, fmt, vet, race/unit, E2E owner serta contract tests yang relevan benar-benar lulus.
- Observability tanpa secret/PII, readme/handoff/index dan bukti commit main diperbarui.

## Gate proses
| Gate | Exit evidence |
|---|---|
| G0 | Scope/BRD/PRD/ADR/owner dan resource diakui |
| G1 — Internal | Purchase→actual receiving→Finance matching dan IAM staff E2E |
| G2 — Vendor/supplier | PIC scope party/contract, portal, ASN/PO/dokumen dengan negative tests |
| G3 — Marketplace | Provider official approval, signed webhook, event replay/reconcile, analytic correctness |
| G4 — Customer terakhir | Browser opaque customer/guest, checkout/OMS Pricing/WMS, tracking, privacy/UAT |
| G5 — Deploy | Recovery/runbook, security review, backup restore, owner sign-off dan immutable SHA |

## Defect dan stop conditions
P0: data bocor cross-party/tenant, unauthorized write, duplicate OMS order, double Finance receipt, webhook signature bypass, fail-closed berubah menjadi allow. Stop release alur terdampak sampai diperbaiki dan diverifikasi. P1: stock drift, misquote, event unrecoverable; jangan mengaktifkan fitur terdampak. P2/P3 UX/docs bisa dijadwalkan berdasarkan risiko.

## Hasil dan laporan
Status `DECIDED|DESIGNED|IMPLEMENTED|VERIFIED|BLOCKED|DEPLOYED`. QA STR, incident, changelog dan UAT disimpan bersama artefak pelaksanaan/release, bukan diisi angka green tanpa eksekusi. CLI test command dan log link harus mencakup hasil actual, bukan hanya niat.
