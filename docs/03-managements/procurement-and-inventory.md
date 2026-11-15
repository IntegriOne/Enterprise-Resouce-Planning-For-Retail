# Procurement dan WMS — Barang Jadi dari Supplier sampai Gudang

> **Dokumen management retail · 12 November 2026.** Kepemilikan bisnis normatif mengikuti [agents/boundary.md](../../agents/boundary.md). Ini ringkasan, bukan instruksi melakukan perubahan lintas management atau seed IAM otomatis.

## Batas penting
- **Procurement:** supplier sourcing/qualification, purchase requisition, PO, kontrak, ASN dan status komitmen pembelian.
- **WMS:** lokasi gudang/bin, actual received quantity, stock transaction/balance, putaway, reserve/hold, pick/pack/dispatch, return receiving dan stock count.
- **Finance:** AP invoice/three-way matching, pembayaran yang benar-benar dibuktikan, jurnal.

## F1: pembelian contoh
PO memesan 100, supplier ASN 100, WMS menerima 98 unit baik. Procurement tetap mengetahui 100 ordered/98 received, WMS menambah hanya 98 saldo aktual, Finance menahan matching jika PO/receipt invoice 100 tidak cocok. Tidak ada service selain WMS yang langsung mengubah `inventory_balances`.

## F2: supplier portal
Supplier PIC harus login sebagai principal manusia aktif yang diikat party Foundation, dan melihat PO scope party sendiri. ASN duplikat dengan key identik harus idempoten; payload berbeda dengan key sama harus conflict. Vendor jasa bukan supplier barang secara implisit.

## F3: marketplace fulfill
OMS dan Commerce meminta availability/hold WMS dengan qty, expiry, version, key. Webhook order yang diulang tidak mengurangi stok dua kali. Marketplace outbound available quantity hanya proyeksi WMS dan perlu reconciliation untuk oversell.

## Safety dan test
Parallel holds >available harus gagal/partial sesuai kontrak; wrong-party PO tidak bocor via list, count atau Node. Mutasi DENY tidak menambah stock/PO/outbox. Lost response memerlukan lookup key sama, bukan retry dengan key baru.

[Procurement](../../backend/services/procurement_sourcing/README.md) · [WMS](../../backend/services/inventory_warehouse_management/README.md).
