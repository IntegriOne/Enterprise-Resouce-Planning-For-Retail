# Project Charter dan Business Case

> **Manajemen proyek Retail ERP · 12 November 2026.** Draft untuk keputusan owner; bukan SOW ditandatangani, hasil penelitian pasar, UAT lulus atau deployment. [Roadmap](../01-product/roadmap.md) · [DoD](../04-engineering/quality-gates.md).

## Mandat
Membangun ERP retail + website Commerce bagi UMKM yang menjual produk jadi. Sponsor/approver akhir: **Owner Integrione**. Leader menjaga domain boundary dan arsitektur, developer bertanggung jawab atas management masing-masing, QA memvalidasi, ops menyiapkan environment/release, legal/PIC merchant menyetujui kontrak dan kebijakan publik.

## Sasaran
Alur internal SKU→PO→receipt→Finance yang auditable; akses supplier/vendor terisolasi; sinkronisasi marketplace aman dan analitik SKU; storefront pembeli sebagai fase terakhir. Keberhasilan diukur dengan satu sumber fakta/order/stok/uang, E2E positif-negatif, serta petunjuk biaya/lag nyata.

## Ruang lingkup
13 management terpilih pada [vision](../01-product/vision-and-scope.md). Manufacturing baru **PAUSED**. Bank transfer otomatis/shared multitenancy/partner API yang belum diizinkan bukan ruang lingkup yang otomatis disetujui.

## Stakeholder dan resource
| Peran | Akuntabilitas |
|---|---|
| Owner sponsor | Scope, budget, ADR keamanan/tenancy, launch, change material |
| Product lead/leader | Rencana, dependency, konsistensi requirement dan handoff |
| Domain owner developer | Implementasi + verification + katalog IAM dalam boundary |
| QA/security | Negatif scope, idempotency, E2E dan produk aman |
| Pilot merchant | Kebutuhan riil, UAT dan data yang disetujui |
| Marketplace partner | API scopes, contract, signature, quota dan terms |
| Operations/legal | Hosting/backup, SOP, privacy, SOW/DPA/SLA |

## Belum ditetapkan
Pilot merchant, jenis produk, biaya, tanggal target, SLA, pricing paket, hosting per pelanggan, provider channel pertama, partner API eligibility. Keputusan ini jangan dipalsukan untuk mengisi dokumen. Charter berlaku setelah owner menandatangani scope/roadmap, bukan karena file ada.
