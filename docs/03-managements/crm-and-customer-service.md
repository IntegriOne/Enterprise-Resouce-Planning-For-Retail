# CRM dan Customer Service — Lead, Pelanggan dan Tiket

> **Dokumen management retail · 12 November 2026.** Kepemilikan bisnis normatif mengikuti [agents/boundary.md](../../agents/boundary.md). Ini ringkasan, bukan instruksi melakukan perubahan lintas management atau seed IAM otomatis.

## CRM & Sales Management
CRM menyimpan lead/prospect, hubungan pelanggan, pipeline dan kampanye opt-in yang diizinkan. Tidak mengambil alih `Commerce customer account` (akses kanal), `OMS sales order` atau `Finance receivable`. Data pelanggan dari marketplace hanya dapat dipakai sesuai kontrak/terms provider dan dasar privasi.

## Customer Service
Tiket, SLA request, komplain, bantuan order dan eskalasi retur/pembayaran. Akses ke status order yang benar dari OMS, pengiriman WMS, receipt Finance dan dokumen DRM melalui referensi yang diotorisasi; tiket tidak memberi izin memperbaiki transaksi owner secara sepihak.

## F4 (customer terakhir)
Search/FAQ/kontak, laporan masalah, ticket receipt, tracking dan return status; nominasi CRM marketing harus mengikuti consent dan preference. Agen support hanya melihat PII yang diperlukan untuk kasusnya; list/search/customer_id lain discope.

## Implementasi dan dependency
Keduanya termasuk 13 management terpilih namun status awal pada repo lebih banyak planned/scaffold; service bisnis penuh dan GraphQL E2E harus diaudit dahulu. Proses dukungan manual sah saat backend belum siap, tetapi jangan memasukkan data customer ke channel yang tidak aman.

[CRM](../../backend/services/crm_sales_management/README.md) · [Customer Service](../../backend/services/customer_service_management/README.md).
