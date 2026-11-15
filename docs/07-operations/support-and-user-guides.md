# Quick Start, User/Admin Manual, FAQ dan Support Flow

> **Operasional Retail ERP — draft SOP.** Tidak ada perintah deployment produksi atau tes restore yang diklaim berhasil melalui penulisan dokumen. Kebijakan owner dan gate ada di [decisions](../00-governance/decisions.md), [quality gates](../04-engineering/quality-gates.md).

## Panduan internal F1
Login akun IAM dengan MFA sesuai policy→setup legal entity/location→buat item SKU/varian→daftarkan supplier party→Procurement PO→WMS menerima aktual→Finance memeriksa invoice→dashboard status. Jika 403, jangan memakai akun Super Admin untuk bypass. Status PENDING tidak dianggap sukses final.

## Portal vendor/supplier F2
Supplier PIC terverifikasi login→pilih party authorised→baca PO sendiri→ASN/bukti→track menerima atau selisih; vendor jasa hanya melihat kontrak/tugas/versi dokumen yang di-grant. Expiry/revoke membatasi sesi, tidak menghapus arsip kontrak/PO.

## Admin channel F3
Hubungkan shop via API resmi yang diizinkan dan secret vault→initial sync/backfill→pantau webhook/inbox/outbox, token expiry, shop mapping, reconciliation, data freshness dan analytics provenance. Provider 200 ACK tidak berarti order OMS/stock WMS accepted. Jangan menggandakan request dengan key baru ketika status UNKNOWN.

## Buyer F4 (belum dibuka sekarang)
Browse katalog LIVE, pilih varian, cart, checkout konfirmasi quote final, tunggu keputusan OMS, lacak pengiriman/retur, ajukan bantuan. Status pembayaran berdasarkan Finance atau provider resmi sesuai kontrak. UX/help harus dapat dipakai mobile, accessible, dan privasi/terms ditinjau.

## FAQ
**PO sudah approved tapi stok belum bertambah?** Only WMS actual receipt increases physical stock. **Bukti transfer ada tapi belum PAID?** Finance harus memverifikasi. **Marketplace lebih lambat dari website?** Lihat source timestamp, sync lag, event backlog dan reconciliation; jangan menganggap webhook always delivered. **403?** Periksa role/scope, bukan isi cookie. **503?** Health dependency, gunakan retry idempoten.

## Support dan aftersales
L1 triage/knowledge base, L2 domain owner, L3 security/platform/provider. Tiket memuat ID, waktu UTC, source trace ID, subjek/scope yang aman, business impact, status dan owner. Jangan meminta password/token/PII berlebihan. Jam respons SLA, contact support, warranty dan kontrak belum diisi. Tutup tiket hanya sesudah koreksi bisnis diverifikasi.
