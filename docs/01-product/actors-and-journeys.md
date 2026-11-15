# Aktor dan Journey Bisnis

> **Integrione Retail ERP · 12 November 2026 · status: rancangan produk.** Sumber keputusan produk: [governance](../00-governance/decisions.md). Kebutuhan target bukan bukti implementasi.

## Identity tidak sama dengan peran bisnis
Satu `principal` IAM untuk satu manusia dapat berelasi ke pekerja HRM, representative supplier, vendor jasa, dan buyer Commerce. Role/grant efektif harus disetujui serta dibatasi party/legal entity/source dan tidak diambil dari `principals.attributes`. Vendor jasa dan supplier barang **dua relasi berbeda**. Akun workload connector tidak otomatis dapat mengaku sebagai user.

| Aktor | Journey utama | Akses minimum |
|---|---|---|
| Owner/admin internal | Konfigurasi perusahaan, periksa dashboard, delegasi staf | IAM grant yang sah untuk tiap management; superadmin IAM bukan wildcard transaksi |
| Product/catalog staff | Daftar SKU, varian, dan konten channel | PLM item write vs Commerce listing publish terpisah |
| Purchasing | Supplier → PO → revisi → ASN status | Procurement scope legal entity/party |
| Warehouse | Receipt aktual, stock adjustment, pick/pack/dispatch | WMS scope gudang; tidak posting Finance |
| Finance staff | AP matching, AR receipt, ledger, laporan | Finance permission dan recheck aksi sensitif |
| HRM | Worker profile, penempatan, usulan akses | Tidak langsung membuat IAM grant |
| Supplier PIC | Login→lihat PO sendiri→ASN/bukti→status | IAM scoped supplier party + asset owner |
| Vendor PIC | Login→lihat kontrak/tugas/bukti yang diizinkan | IAM scoped vendor party/contract |
| Marketplace workload | Link akun resmi→webhook→sync/reconcile | Auth API resmi + workload identity, account scope |
| Customer service | Kelola tiket/account/order referensi | Case-scoped PII; tidak mengubah OMS/Finance |
| Calon pembeli (F4) | Browse kategori/varian publik | Allowlist listing yang LIVE |
| Pembeli (F4) | Guest/cart→quote→submit→track/return | Ownership cart/account dan OMS data view yang dibolehkan |

## Critical journeys dan negative cases
**Internal:** vendor/supplier master → approve PO → receive partial qty → WMS balances → Finance matching. Pembelian 100 tetapi penerimaan 98 bukan stok 100. 

**Mitra:** PIC supplier A melihat hanya PO A, bukan PO B, termasuk query list/count/Node. Retrying ASN key sama tidak membuat dua ASN.

**Multichannel:** 1 external order = 1 OMS order; provider webhook duplikat/out-of-order tidak mengubah stok atau membuat invoice kedua. Jika provider down, UI menampilkan freshness/DEGRADED, bukan data realtime palsu.

**Buyer F4:** publik hanya listing LIVE, guest A tidak melihat cart guest B, quote digest lama ditolak setelah perubahan harga/qty, OMS state PENDING tidak dianggap paid atau accepted.
