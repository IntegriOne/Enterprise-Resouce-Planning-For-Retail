# Risk Register, Stakeholder dan Change Control

> **Manajemen proyek Retail ERP · 12 November 2026.** Draft untuk keputusan owner; bukan SOW ditandatangani, hasil penelitian pasar, UAT lulus atau deployment. [Roadmap](../01-product/roadmap.md) · [DoD](../04-engineering/quality-gates.md).

## Risk register awal (nilai hipotesis; belum hasil observasi)
| ID | Risiko | Prioritas | Mitigasi/PIC |
|---|---|---|---|
| R-01 | Oversell dari website/marketplace | P0 untuk penjualan | WMS hold/version dan OMS idempotency |
| R-02 | Forged/duplikat webhook | P0 keamanan/uang | Official signature, inbox dedupe, audit |
| R-03 | Supplier/customer PII bocor lintas party | P0 | IAM ABAC, Node/list tests dan data minimization |
| R-04 | Harga/pajak/quote tidak konsisten | P1 | Pricing/Finance signed source decision, quote expiry/digest |
| R-05 | Provider API izin tidak disetujui | BLOCKED | POC/contract resmi, scope kecil, no fake event |
| R-06 | Data Finance/stock asal yang tidak sinkron | P0/P1 sesuai dampak | Source validation, reconciliation, stop posting |
| R-07 | Kegagalan Redis/pencabutan tak efektif | P0 security | Fail closed, version fence, recovery gate |
| R-08 | Banyak developer pada `main` | P1 | Koordinasi file bersama, CI, non-force push |
| R-09 | Dokumen/fitur dipasarkan sebelum VERIFIED | P1 reputasi | Release gates dan sign-off UAT |
| R-10 | Perjanjian dan privasi belum diputuskan | BLOCKED launch | Legal/DPA/terms, owner sign-off |

Owner risiko, probabilitas, dampak, residual dan tanggal review harus diisi. Skala P/I tidak boleh dianggap probabilitas terukur.

## Change Request
`CR-ID | requester | baseline affected (BRD/PRD/SOW/ADR) | before→after | impact budget/security/data/API | alternatives | approver | APPROVED/REJECTED/DEFERRED | commit/test`.

Owner menyetujui scope/budget, security D01–D12, ownership domain, tenancy, bank integration dan go-live. Perubahan hanya bisa dilakukan management pemilik. Penanganan P0 dapat memulai containment segera, tetapi tetap memerlukan audit/decision dan incident report. Tidak force-push atau SQL perbaikan langsung ke DB management lain.

## Stakeholder communication
Internal owner, staf, supplier/vendor PIC, pilot merchant, customer, privacy/legal, platform dan partner marketplace memiliki hak keputusan yang berbeda. Update status faktual/tes, jangan berjanji SLA, license atau fitur API sebelum sumber resmi dan kontrak disetujui.
