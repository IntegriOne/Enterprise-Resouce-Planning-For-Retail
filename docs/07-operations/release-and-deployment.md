# Release Checklist, Release Notes dan Deployment

> **Operasional Retail ERP — draft SOP.** Tidak ada perintah deployment produksi atau tes restore yang diklaim berhasil melalui penulisan dokumen. Kebijakan owner dan gate ada di [decisions](../00-governance/decisions.md), [quality gates](../04-engineering/quality-gates.md).

## Release gate
- [ ] Commit SHA/tag yang immutable disetujui owner dan berasal dari `main` dengan bukti build/test.
- [ ] Lingkup phase dan SOW/PRD terverifikasi; tidak ada P0 keamanan/data/uang outstanding.
- [ ] IAM sessions 30m/8h/30s dan no fallback allow, CSRF browser/guest, supplier party isolation diuji.
- [ ] Business owners: PO/receipt/Finance, OMS one-order, WMS holds, Pricing quote, Commerce login/listing/Node sesuai fase aktif.
- [ ] Provider channel resmi, credentials scope, webhook signature/replay, reconciliation, analytics source, market agreement dan token rotation lulus bila channel diaktifkan.
- [ ] Schema/migration production direncanakan dengan backup; baseline `000001` hanya kebijakan pra-rilis. Jangan reset database customer.
- [ ] Restore test, health/readiness, observability, incident/support, privacy/terms/DPA sesuai peran disetujui.
- [ ] Go/No-Go ditandatangani owner/klien/PIC yang benar.

## Deploy & rollback
Gunakan artifact dari commit SHA yang diuji. Simpan konfigurasi/secret di vault per deployment, PostgreSQL credentials per service, mTLS/private network, HTTPS host, Redis recovery gate, migration log. Rollback dengan cara yang menjaga state/revoke efektif dan **tidak mengaktifkan kembali grant lama atau menimpa schema**. Jika perubahan domain sudah membuat data baru, perlu strategi kompatibilitas yang dirancang sebelumnya.

## Release notes template
`version | tag | SHA | deployment | date | features verified | migration | security/user action | known issues | rollback steps | support contact`. Changelog riwayat banyak release, release notes hanya satu. Dokumentasi paket ini tidak membuat rilis baru atau menambah tag.

## Compliance dan access
Penerbitan customer-facing website baru tahap F4 setelah legal/scope/checkout/UAT selesai. Shadow deployment/preview tidak boleh dianggap rilis umum.
