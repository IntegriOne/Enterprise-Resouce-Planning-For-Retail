# Personas dan Skenario User

> **UX/UI design brief · belum berupa file Figma atau prototipe visual jadi.** Keputusan fase mengikuti [produk](../01-product/roadmap.md) dan kebijakan akses [IAM](../02-architecture/security-and-sessions.md).

## Persona draft, bukan hasil riset pengguna
| Persona | Tujuan utama | UX problem |
|---|---|---|
| Owner UMKM | Melihat cashflow, order, stok dan channel kinerja | Banyak tool dan sumber yang berbeda |
| Staf purchasing | PO supplier dan barang diterima | Kuantitas pesan vs terima tidak sama |
| Staf gudang | Receiving, picking, packing, dispatch | Stock hold dan barcodes tidak sinkron |
| Finance staff | Invoice, matching, receipt, refund | Upload bukti tidak sama dengan dana masuk |
| Supplier PIC | Melihat PO sendiri, membuat ASN | Salah scope dapat membocorkan vendor lain |
| Vendor jasa PIC | Mengakses kontrak/tugas/bukti relevan | Vendor ≠ supplier, izin harus berbeda |
| Channel manager | Link Shopee/TikTok, cek event health dan sync | Token expiry, rate limit, missing events |
| Customer service | Menjawab tiket, cek order/refund terotorisasi | PII tidak boleh bocor di case lain |
| Calon pembeli F4 | Browse barang tanpa login | Butuh foto, harga, variant, availability jelas |
| Customer F4 | Cart→checkout→track/retur | Quote kadaluarsa, pembayaran pending, order unknown |

## Story scenarios
- Internal: daftar SKU/variant, buat PO, lihat receiving parsial, periksa selisih Finance.
- Supplier: login, hanya PO party A, buat ASN dan lihat status; akses PO B ditolak tanpa membuka count.
- Marketplace: connect account resmi, initial backfill, heartbeat, webhook batch, konflik dan reconciliation.
- Buyer: cari varian, masukkan cart guest, checkout dengan quote final, lihat order status; tidak diklaim paid otomatis.

## Riset awal
Wawancara merchant/staf/supplier, amati pekerjaan nyata dengan data sintetis bila mungkin, tes prototipe ke berbagai ukuran layar, keyboard dan pembaca layar. Target sample, conversion benchmark dan pola desain akhir belum diputuskan. Tulis hasil dan consent sebelum menyatakan keputusan UX tervalidasi.
