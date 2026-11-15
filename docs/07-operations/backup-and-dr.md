# Backup, Restore dan Disaster Recovery

> **Operasional Retail ERP — draft SOP.** Tidak ada perintah deployment produksi atau tes restore yang diklaim berhasil melalui penulisan dokumen. Kebijakan owner dan gate ada di [decisions](../00-governance/decisions.md), [quality gates](../04-engineering/quality-gates.md).

## Aset yang harus dilindungi
| Sumber | Strategi yang perlu disahkan | Kriteria restore |
|---|---|---|
| PostgreSQL per management | Encrypted backups, PITR jika tersedia, retention & access controls | Schema, constraints, row counts, outbox/inbox, domain state & revision valid |
| IAM sessions, Redis security | Persistence bukan sumber grant; recovery epoch dan fence otoritatif | Tidak membuka trafik sebelum status IAM source direkonsiliasi |
| DRM/media customer | Versioned encrypted object store, digest dan legal hold | Versi serta ACL sesuai |
| Provider connector | Encrypted token vault, mapping shop, sync checkpoints | Reconnect dan replay hanya event yang sah, no double OMS order |
| Release/infra | Artifact digest/sha, config template, secret rotation, certs | Deploy ulang terisolasi dari backup tanpa memasukkan secrets ke Git |

## RPO/RTO
Target RPO dan RTO belum disetujui atau diukur. Tidak mengklaim zero data loss, seamless HA atau failover auto-safe. Ukur pada restore drill dan paparkan gap yang berakibat order/stock/uang.

## Restore drill
1. Buat environment disposable dengan DB berakhiran `_test`; validasi nama sebelum instruksi destruktif.
2. Restore sample backup terenkripsi, jalankan integrity checks per owner, cek migrasi, PII redaction dan isolation.
3. Lakukan reconcile outbox/inbox, external event mappings, OMS orders, WMS holds/balances, Finance receipts dan channel watermark.
4. Validasi IAM login scoped, revoke effective, Redis recovery gate; tidak menghidupkan sesi yang tak dapat dibuktikan.
5. Re-run service/E2E terpilih pada data sintetis, catat durasi/kehilangan data dan approval.
6. Jangan menghapus backup sebelum retention/legal hold terpenuhi.

## Keputusan disaster
DRP ditetapkan ketika deployment/hosting per customer dipilih. Setiap pelanggan memiliki batas data dan kunci terpisah sesuai keputusan arsitektur, bukan sekadar label “tenant”.
