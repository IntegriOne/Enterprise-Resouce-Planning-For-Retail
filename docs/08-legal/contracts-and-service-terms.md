# SOW, MSA, NDA dan Kontrak Dukungan — Draft

> **DRAFT UNTUK REVIEW HUKUM — bukan dokumen siap ditandatangani atau dipublikasikan.** Legal/owner/klien wajib memverifikasi regulasi Indonesia yang berlaku, peran pengendali/prosesor dan syarat marketplace resmi. Tidak ada legal review eksternal yang dijalankan pada rekonstruksi dokumentasi ini.

## Statement of Work (SOW)
`Parties/registration numbers, work scope by phase, project deliverables, exclusions (Manufacturing paused, bank automation no-go), source data inputs, implementation milestones, dependencies and acceptance/UAT, fees/tax, warranty, signoffs`. Perubahan pekerjaan memakai CR/ADR, bukan membuat change gratis di management lain.

## Master Services Agreement (MSA)
`Service delivery/licensing, IP ownership, source code, hosting, data ownership, privacy/DPA, subprocessor, backup, security, billing/payment, support/service credits, incident notices, termination, export/offboarding, law/dispute venue, liability constraints`. Model RaaS satu instalasi per company adalah baseline teknis, **bukan penetapan otomatis siapa pemilik source code**.

## Non-Disclosure Agreement (NDA)
`Disclosing/receiving parties, mutual/single scope, definition of confidential information, exclusions, purpose, staff need-to-know, retention/destruction, duration, remedies, signoffs`. NDA tidak menggantikan DPA. Kategori rahasia: source code privat, supplier prices, inventory/Finance, customer PIIs, API credentials dan security keys.

## Client SOW acceptance
| Deliverable | Bukti yang harus tersedia |
|---|---|
| IAM/setup | Principal/scopes ditinjau dan test negatif |
| Internal retail | SKU/PO/WMS/Finance E2E |
| Supplier/vendor | Party-scope/ASN/contract access negatif |
| Marketplace | Official provider credentials, signed webhook, retry/reconciliation dan analytics |
| Buyer F4 | Mobile commerce, checkout, OMS order, Finance/returns, privacy UAT |
| Handover | Immutable SHA/tag, backup restore, user guide, support contact, license/terms |

**Semua angka biaya, tanggal, SLA, lisensi dan pihak penandatangan adalah OPEN.** Penerimaan jasa baru berlaku dari kontrak yang sudah diisi dan disetujui sesuai hukum.
