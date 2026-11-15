# Market Requirements (MRD), RFI/RFP dan Proof of Concept

> **Manajemen proyek Retail ERP · 12 November 2026.** Draft untuk keputusan owner; bukan SOW ditandatangani, hasil penelitian pasar, UAT lulus atau deployment. [Roadmap](../01-product/roadmap.md) · [DoD](../04-engineering/quality-gates.md).

## Hipotesis pasar
UMKM brand/reseller memiliki data katalog, supplier, stok, order dan keuangan yang tersebar; website sendiri dan marketplace mengakibatkan input ulang serta risiko oversell. **Belum ada wawancara atau audit live Yoorasarah** pada sumber ini. Segmen, ukuran pasar, willingness to pay dan kategori produk harus divalidasi.

## MRD discovery
Wawancara owner/staf pembelian/gudang/Finance, vendor/supplier serta calon pembeli; petakan volume SKU, kanal aktif, masalah stok/order, kebijakan retur, biaya sekarang, kebutuhan data dan hosting. Catat sample, metode, consent, observasi, bukti dan bias. Jangan mengklaim product-market fit tanpa tes.

## RFI/RFP
**RFI** mengumpulkan kebutuhan calon pelanggan: legal entity, SKU/varian, gudang, supplier, pengguna/role, kanal Shopee/TikTok, volume order, catatan pajak, jenis pembayaran, migrasi dan support. **RFP** membandingkan kemampuan calon penyedia/partner: keamanan, contract, rate limit, backup, API docs, SLA, biaya dan demonstrasi. Tidak ada supplier/partner terpilih otomatis.

## POC: percobaan reversible
Satu merchant synthetic: IAM positif/negatif → supplier party→PO→WMS receipt→Finance match. Satu adaptor marketplace dengan **official sandbox** untuk signed webhook, 100 duplicate events→satu OMS effect, lost response recovery, analytics SKU dari fakta owner. Kriteria GO: tidak ada cross-scope leak, no duplicate order/receipt, source reconciliation, backup restore dan operator mampu menjelaskan kesalahan.

## Hasil yang harus dicatat
`POC-ID | hypothesis | environment | source docs version | fixture | steps | actual pass/fail | evidence URL/SHA | residual risk | GO/REVISE/NO-GO | approver`. **Belum ada POC baru yang dijalankan oleh dokumen ini**. Missing provider permission adalah BLOCKED, bukan validasi.
