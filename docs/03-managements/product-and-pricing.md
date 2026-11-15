# PLM dan Pricing — Produk Jadi, Varian, Harga

> **Dokumen management retail · 12 November 2026.** Kepemilikan bisnis normatif mengikuti [agents/boundary.md](../../agents/boundary.md). Ini ringkasan, bukan instruksi melakukan perubahan lintas management atau seed IAM otomatis.

## PLM pemilik item
Product Lifecycle Management mengelola item/SKU, varian, item identity, revision, spesifikasi dan eligibility. Fase retail cukup bagian katalog; tidak mewajibkan EBOM, PFMEA, routing manufaktur. Field storefront (nama kampanye, foto channel, slug/SEO, koleksi dan jadwal publikasi) milik Commerce, bukan master varian PLM.

## Pricing pemilik keputusan komersial
Pricing & Revenue Management menentukan price lists, price tiers, promotion, voucher, discount allocation, currency dan rounding policy/expiry. Commerce boleh menyimpan **snapshot quote yang telah disetujui buyer**, tetapi bukan keputusan harga kedua.

## Kontrak minimal
`(seller/channel, SKU+revision, qty, customer eligibility, currency, coupon code, as_of) → price/unit, discount allocation, tax-hand-off, valid_until, source_revision, calculation_version`. Nama function/RPC harus disepakati developer Pricing dan dibuktikan pada proto aktif, bukan dibuat fiktif dalam dokumen ini.

## SKU dan kanal
Satu produk dapat dipasarkan dengan external listing/SKU berbeda di Shopee dan TikTok; mapping stabil per provider+shop+variant, tidak boleh menyamakan nama produk sebagai identity. Data price expired diklasifikasi STALE/UNKNOWN; rate pihak ketiga tidak mengubah source ledger PLM atau Pricing.

## Status
PLM memiliki service bisnis Go dan DBML; Pricing sampai snapshot dokumen yang diperiksa masih planned/scaffold. Jangan mengaktifkan perhitungan fallback permanen di Commerce jika kontrak pricing wajib belum tersedia.

[PLM](../../backend/services/product_lifecycle_management/README.md) · [Pricing](../../backend/services/pricing_revenue_management/README.md) · [Commerce](../../backend/services/ecommerce_commerce/README.md).
