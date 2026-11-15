# SOW, MSA, NDA, RFI/RFP, RFC dan ADR — Template

> **Template pengelolaan proyek, status DRAFT/NOT RUN** sampai persetujuan atau pengujian aktual. Rujukan: [Charter](project-charter.md), [Requirements](../01-product/requirements.md), [DoD](../04-engineering/quality-gates.md).

> **Dokumen hukum di bawah hanya kerangka; bukan kontrak siap ditandatangani.** Legal reviewer dan pihak berkepentingan harus menyusun versi sah di luar placeholder.

## SOW (Statement of Work)
`SOW-ID, provider legal name, client legal name, project phase, modules, implementation deliverables, exclusion, input client, schedule, milestone acceptance/UAT, commercial price/tax, warranty, support, signature`. Contoh scope harus spesifik: internal SKU→PO→receipt terlebih dahulu; buyer website bila F4 telah siap.

## MSA (Master Services Agreement)
`Contract parties; service definitions; priority vs SOW; IP/software licensing; billing; support/SLA; security; data ownership; privacy/DPA; vendor/provider API terms; warranties/liability; termination/exit; applicable law; dispute resolution`. Biaya/ownership code tidak boleh diasumsikan dari kata RaaS saja.

## NDA (Non-Disclosure Agreement)
`Parties; disclosing vs receiving; confidential information; purpose and exclusions; need-to-know; duration; permitted disclosure; protection; return/deletion; remedies; governing law; signatures`. NDA **tidak menggantikan DPA atau Privacy Policy**.

## RFI dan RFP
**RFI:** inventory SKU, legal entity, supplier, gudang, jumlah pengguna, channel marketplace, payment methods, volume, data migrations, retention. **RFP:** minta proposal detail API security, feature coverage, demo E2E, privacy, backup, pricing, SLA dan compatibility, lalu nilai secara eksplisit.

## RFC dan ADR
`RFC-ID | problem | evidence source | proposal/alternative | auth/ownership impact | API/schema change | migration/rollback | acceptance tests | review comments`. Jika diterima menjadi `ADR-ID | status | context | decision | considered alternatives | consequences | owner | effective date | evidence`. Jangan menyebut D01–D12 berubah tanpa owner sign-off.

## Nota kerja
Vendor/supplier/customer agreement per merchant memerlukan data bisnis dan pihak nyata. Dokumentasi template tidak mengotorisasi mengirim dokumen, menerima pembayaran, atau membuat pelanggan baru di produk.
