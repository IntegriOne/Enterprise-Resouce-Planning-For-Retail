# Project Management Plan, Roadmap, Estimasi dan RACI

> **Manajemen proyek Retail ERP · 12 November 2026.** Draft untuk keputusan owner; bukan SOW ditandatangani, hasil penelitian pasar, UAT lulus atau deployment. [Roadmap](../01-product/roadmap.md) · [DoD](../04-engineering/quality-gates.md).

## Tahap kerja dan dependency
| Gate | Prasyarat | Bukti hasil |
|---|---|---|
| G0 Scope | Charter, BRD/PRD, ownership/ADR | Owner memutuskan scope dan risiko |
| G1 Internal | IAM/PLM/Foundation/Procurement/WMS/Finance | E2E SKU/PO/receipt/invoice dan negative grants |
| G2 Mitra | Supplier/vendor party & IAM access | PO/ASN/doc scopes, No cross-party |
| G3 Marketplace | Official partner auth, OMS, Pricing/WMS mapping | Signed webhook, inbox/outbox, retry, analytics watermark |
| G4 Buyer | Commerce full, IAM CSRF, Pricing/OMS, privacy | Customer/public/guest/checkout/track/return E2E dan UAT |
| G5 Release | QA, security, backup, legal & support | Owner production sign-off pada SHA tervalidasi |

## RACI ringkas
Owner Accountable untuk scope/biaya/launch; developer management Responsible untuk domain/tes; IAM & Gateway Consulted untuk auth transport; integration owner bertanggung jawab atas provider contract; QA & legal berwenang pada gate yang sudah diberi mandat. Tidak ada izin lintas management hanya karena dicantumkan dalam plan.

## Estimasi
Gunakan **PERT**: `(O + 4M + P)/6` untuk optimistic, most-likely, pessimistic. Example `O=2 M=4 P=8` → 4,33 hari adalah ilustrasi metode, **bukan estimasi proyek**. Resource planning mencatat orang/jam tersedia, dependency, shared file conflict dan buffer reviewer. Tanggal milestone, budget, Gantt dan SLA butuh data/keputusan owner.

## Kerja sprint
Backlog ID, acceptance, owner, dependency dan DoD; WIP ideal satu vertical slice/domain per developer sampai E2E. Sprint review menunjukkan demo nyata, bukan jumlah file. Velocity menghitung story **ACCEPTED**, burndown menampilkan scope change, bukan jam kerja sebagai bukti kualitas.

## Komunikasi
Review dependency/risiko/keputusan berkala sesuai jadwal yang disepakati. P0 cross-party leak, order ganda, security incident atau receipt salah langsung dieskalasikan, jangan menunggu sprint review. Semua commit hanya `main` dan file shared dikoordinasikan.
