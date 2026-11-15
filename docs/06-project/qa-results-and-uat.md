# Test Plan, STR, Bug Report dan UAT — Template

> **Template pengelolaan proyek, status DRAFT/NOT RUN** sampai persetujuan atau pengujian aktual. Rujukan: [Charter](project-charter.md), [Requirements](../01-product/requirements.md), [DoD](../04-engineering/quality-gates.md).

## Test Plan
`PLAN-ID | phase gate | scope domain | environment | test accounts/scopes | fixtures | test case IDs | responsible | entry/exit criteria | schedule`. Strategi E2E dan 14 kasus awal ada di [testing](../04-engineering/testing.md). Kriteria: jangan melabeli mock sebagai E2E, database berakhir `_test`, provider official permission perlu diuji.

## Software Test Results (STR)
| Test ID | Requirement | Command/SHA | Expected | Actual | Status |
|---|---|---|---|---|---|
| TC-01 | BR-01 | [belum] | Scoped positive/deny | [belum] | NOT RUN |
| TC-03 | BR-01 | [belum] | Partial receipt & Finance mismatch | [belum] | NOT RUN |
| TC-08 | BR-03 | [belum] | 100 duplicates→satu efek OMS | [belum] | BLOCKED provider/OMS |
| TC-12 | BR-05 | [belum] | Quote lama ditolak, OMS one-order | [belum] | BLOCKED F4 |

Catat total pass/fail/blocked/skipped, lingkungan/fixture, evidence log ber-redaksi, DB assertions, source version dan hasil rerun. **Angka pass tidak boleh diisi dari rencana.**

## Bug report
`BUG-ID | severity P0–P3 | source SHA | actor/scope | input API sanitized | steps reproduce | expected/actual | owner DB/outbox effect | logs | fix commit | regression test | verification`. P0 security leak/duplicate financial booking/OMS order/unauthorized write memblokir release alur terdampak.

## User Acceptance Test
Skenario: internal merchant staff SKU→PO→receipt; supplier own PO/ASN; vendor contract restricted; marketplace official channel event sync + dashboard; buyer F4 product→checkout→tracking. `UAT-ID | persona | goal | steps | business expected | actual | defects | client/PIC signoff | date`. Gagal/kekurangan harus dilaporkan, bukan dihapus.

## Review dan closure
QA/owner menyetujui kesiapan setelah E2E nyata dan komersial/legal siap. UAT fase internal tidak otomatis membuktikan readiness marketplace atau customer.
