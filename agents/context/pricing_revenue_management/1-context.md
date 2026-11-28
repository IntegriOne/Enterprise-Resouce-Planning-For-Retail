# Konteks Pricing & Revenue Management

## Apa yang sudah dikerjakan

- Scope produk menetapkan Pricing sebagai pemilik keputusan harga jual, promo, diskon, mata uang, pembulatan, dan masa berlaku. PLM memegang identitas dan revisi SKU, Finance menghitung pajak, dan Web Commerce menyimpan snapshot harga pada quote. [Batas management](../../boundary.md#pricing_revenue_management) · [Desain produk dan harga](../../../docs/03-managements/product-and-pricing.md)
- Pada source lokal Web Commerce yang diperiksa 29 November 2026, `PricingPort.PriceCart` menerima kanal, kategori pembeli, mata uang, kode promo, dan baris item. Hasilnya membawa rincian harga, `PricingSnapshotDigest`, dan `TTL`. `PrepareCheckoutQuote` menolak kalkulasi bila port tidak terpasang. Consumer inbox mengenali `PriceRevisionPublished` dan menandai listing serta quote terkait sebagai `stale`. Bukti lokal ada di `backend/services/web_commerce/domain/service/ports.go`, `services/checkout/service.go`, dan `services/integration/service.go` dalam modul yang sama.

## Apa yang belum dikerjakan

- Checkout belum memuat modul Pricing di `backend/services/` atau entri Pricing di `backend/go.work`. Wiring `backend/services/web_commerce/cmd/wiring.go` juga belum memasang port Pricing. Kontrak gRPC dari pemilik, skema, katalog izin, wiring Gateway, dan bukti E2E Pricing belum tersedia dalam cakupan yang diperiksa.

## Apa yang harus dikerjakan selanjutanya

- Saat tahap Pricing dibuka, sepakati kontrak evaluasi harga dengan Web Commerce, PLM, dan Finance. Tetapkan referensi SKU dan revisi, aturan eligibility, mata uang dan pembulatan, versi keputusan, digest, expiry, dan event invalidasi sebelum mengaktifkan checkout yang bergantung pada Pricing.
- Buktikan jalur positif serta promo tidak memenuhi syarat, quote kedaluwarsa, dan dependensi gagal. Pertahankan penolakan checkout ketika keputusan harga yang sah belum tersedia.
