# Konteks Customer Relationship & Sales

## Apa yang sudah dikerjakan

Checkout memuat [DBML](../../../backend/services/customer_relationship_sales/db/schema/customer_relationship_management.dbml) dan [usulan logika](../../../agents/logic/customer_relationship_sales/logic.md). Instruksi pembangunan server dan Gateway beserta daftar file tersedia pada [instruksi 1](../../../agents/instructions/customer_relationship_sales/1-leader-implementasi-server-grpc-dan-gateway-instructions.md). Targetnya 35 query, 54 mutation dan 10 RPC privat. [Katalog operasi](../../../agents/grpc/customer_relationship_sales/1-katalog-operasi-relasi-pelanggan-dan-penjualan.md). Kode Commerce telah mempunyai reference opportunity opsional pada quote B2B.

## Apa yang belum dikerjakan

Pada baseline checkout 29 November 2026, folder service CRM berisi DBML. Modul Go, migrasi, handler, manifest, SDL dan E2E CRM masih perlu dibangun. Kesiapan owner untuk bukti order OMS, revenue Finance, izin/scan DRM dan handoff Customer Service perlu dibuktikan oleh developer saat wiring.

## Apa yang harus dikerjakan selanjutanya

Mulai dengan kontak terscope dari DB sampai HTTPS. Lanjutkan lead, pipeline, consent dan assignment, lalu koordinasi quotation/attribution serta operasi komunikasi dan performance. Hubungkan setiap owner sesuai [instruksi 2](../../../agents/instructions/customer_relationship_sales/2-leader-wiring-lintas-management-instructions.md). CRM memberi konteks relasi kepada CS. Tiket dan SLA tetap di CS. Riwayat penamaan yang lebih lama tersimpan pada [konteks CRM historis](../../../agents/context/crm_sales_management/1-context.md).
