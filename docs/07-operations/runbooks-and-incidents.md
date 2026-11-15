# Runbook Gangguan, Pemulihan dan Postmortem

> **Operasional Retail ERP — draft SOP.** Tidak ada perintah deployment produksi atau tes restore yang diklaim berhasil melalui penulisan dokumen. Kebijakan owner dan gate ada di [decisions](../00-governance/decisions.md), [quality gates](../04-engineering/quality-gates.md).

## 1. Marketplace webhook gagal/tertunda
**Deteksi:** signature failure, token expired, backlog/lag, request quota, missing checkpoints. **Tindakan:** hentikan event tidak terverifikasi, periksa token tanpa mencetak secret, persist event valid ke inbox, lakukan retry/backoff sesuai provider; reconciler membandingkan provider authoritative data dengan mapping owner. **Verifikasi:** lag turun, event gaps dijelaskan, OMS/WMS/Finance facts sesuai. Provider outage bukan izin membuat order palsu.

## 2. Double order atau mismatch stok
**Deteksi:** dua OMS IDs dari source shop/external order ID, atau stok WMS dan holds tidak sesuai. **Containment:** pause command inbound kanal terdampak, karantina event konflik; jangan update langsung tabel owner lain. **Diagnosa:** event id, key, revision, digest, replay history, WMS movements. **Pemulihan:** owner OMS/WMS/Finance membuat koreksi bisnis yang sah, dengan audit/compensating events; cek tidak ada dua receipt.

## 3. IAM/Redis failure
**Deteksi:** 503/auth dependency unavailable, revision stale atau revoke pending. **Tindakan:** stop admission protected jika security state tidak dapat divalidasi; jangan fallback ke JWT legacy ketika opaque gagal. Redis restore memerlukan recovery epoch/snapshot reconcile/readiness gate sebelum membuka layanan. Audit E2E dua akun sesudah pemulihan.

## 4. PostgreSQL dan migration
Periksa DSN dan database owner, locks, storage, replication/backup, migration log. Jangan `DROP DATABASE`/reset production. Dalam pra-rilis, migrasi baseline `000001` sesuai AGENTS; setelah live gunakan kebijakan versioned migration/rollback yang disetujui owner.

## 5. Credential marketplace bocor
Containment connector dan toko terdampak, revoke/rotate provider tokens lewat prosedur resmi, audit akses API, tinjau data pribadi yang mungkin terpapar, eskalasi legal/privacy sesuai kewajiban. Jangan paste token ke chat, issue atau log.

## Incident / postmortem template
`INC-ID | severity | detected/started/resolved UTC | owner/commander | impact merchants/orders | timeline with evidence | containment | root cause | data correction | preventive action + owner + due | verified recovery | client notification decision`. Insiden ditutup setelah state owner dan metrics diverifikasi, bukan setelah server hidup saja.
