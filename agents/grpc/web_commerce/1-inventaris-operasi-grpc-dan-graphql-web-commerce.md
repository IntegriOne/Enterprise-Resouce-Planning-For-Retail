# List operasi Web Commerce

Rancangan untuk satu server privat `web_commerce.v1`. **27 Query + 27 Mutation = 54 operasi root yang diusulkan.** Setiap baris pada kedua matriks memetakan satu field Gateway ke satu RPC Commerce. RPC privat dan worker dicatat terpisah. Nama RPC di sini adalah target implementasi. Operasi yang bergantung pada layanan lain baru dapat dinyatakan terintegrasi setelah ada bukti dari pemiliknya. Jalankan [instruksi pengerjaan](../../instructions/web_commerce/1-leader-implementasi-grpc-dan-gateway-web-commerce-instructions.md).

## RBAC dan cara membaca matriks

- **Publik** berlaku hanya untuk storefront pada host aktif yang dikenali Gateway. **Cart** berarti pemegang cookie guest sah untuk kanal itu atau pembeli dengan cart sendiri. **Self** berarti sesi IAM customer aktif, principal terikat buyer lokal, kanal dan resource miliknya. **B2B** juga memerlukan mandat representasi party EF yang masih aktif. Satu akun dalam party tidak otomatis berhak atas semua dokumen party.
- Untuk staf, kode pendek dalam kolom akses adalah usulan izin dengan awalan `web_commerce.` pada katalog IAM. Semua grant dibatasi legal entity dan channel yang terhubung dengan unit/facility bila diperlukan. Kode seperti `catalog.publish` dan `refund.request` terpisah dari wewenang mengedit katalog atau membaca order. Petakan pula tindakan Self ke izin self yang disetujui IAM. Feature flag hanya mengatur ketersediaan alur.
- Untuk seluruh pasangan di bawah, service mengambil aktor dan scope terverifikasi dari konteks. Untuk perintah yang memicu efek, sertakan kunci idempotensi dan versi yang relevan. List memakai cursor, filter scope sebelum paginasi, dan urutan stabil. Field detail, nested, dan Relay node memakai pemeriksaan objek yang sama. Kolom terakhir menetapkan hasil yang diuji, termasuk penolakan khusus operasi.
- Prioritas pengembangan adalah konfigurasi dan staf F1, kewenangan eksternal yang relevan F2, integrasi dan analitik F3, serta storefront dan buyer F4. Label Publik, Cart dan Self di bawah menjelaskan akses akhir setelah gate F4.

## Butuh GraphQL Query

| Field GraphQL | RPC Commerce | Akses | Input, hasil dan aturan bisnis yang diuji |
|---|---|---|---|
| `storefrontSettings` | `GetStorefrontSettings` | settings.read | Entity dari scope → mata uang, presisi, TTL dan flag. Konfigurasi internal tertutup bagi publik. |
| `salesChannels` | `ListSalesChannels` | channel.read | Entity → kanal, nama, host dan status. Kanal cabang lain terfilter sebelum menghitung halaman. |
| `storefront` | `GetStorefront` | Publik | Host terverifikasi → identitas tampilan seller/kanal aktif. Host asing atau nonaktif tidak memilih kanal default. |
| `deliveryOptions` | `ListDeliveryOptions` | Cart | Kanal, tujuan minimum/pickup dan cart → metode eligible, mata uang, estimasi, TTL. Tarif kedaluwarsa dikutip ulang sebelum sealing. |
| `managedListings` | `ListManagedListings` | catalog.read | Pencarian/filter status di kanal staf termasuk draft. Tidak membocorkan draft ke koneksi publik. |
| `storefrontListings` | `ListPublishedListings` | Publik | Kanal dari host, search/cursor → kartu SKU yang terbit dan layak ditampilkan. Batasi cost dan field publik. |
| `storefrontListing` | `GetPublishedListing` | Publik | Slug dalam kanal → nama, media aman, opsi/varian yang diizinkan. Slug unpublished tidak ditemukan. |
| `storefrontVariant` | `ResolveStorefrontVariant` | Publik | Listing + kombinasi opsi → SKU/revisi/UOM cocok dan availability untuk ditampilkan. Kombinasi invalid ditolak, stok unknown tidak dianggap tersedia. |
| `myBuyerProfile` | `GetMyBuyerProfile` | Self | Profil buyer dan status representasi untuk sesi ini. Identitas login tetap berasal dari IAM. |
| `myDeliveryAddresses` | `ListMyDeliveryAddresses` | Self | Alamat aktif milik buyer. Cursor buyer lain ditolak, snapshot alamat order lama tetap utuh. |
| `shoppingCart` | `GetShoppingCart` | Cart | Cart aktif milik sesi/guest dalam kanal → baris, versi, perkiraan. Query bebas efek samping termasuk saat cart kosong. |
| `checkoutQuote` | `GetCheckoutQuote` | Self | Quote milik buyer → breakdown SKU, diskon, freight, tax, total, masa berlaku, versi. Pembacaan tidak memperpanjang TTL. |
| `checkoutSubmission` | `GetCheckoutSubmission` | Self | Referensi submission → progres, alasan aman, nomor OMS setelah diterima. Timeout owner terlihat sebagai pending. |
| `mySalesQuotations` | `ListMySalesQuotations` | B2B | Daftar penawaran yang boleh diwakili principal, nomor, revisi, expiry dan status. |
| `mySalesQuotation` | `GetMySalesQuotation` | B2B | Satu penawaran, rincian dan syarat yang dibekukan. GID/dokumen di luar mandat tidak membuka harga kontrak. |
| `salesQuotationQueue` | `ListSalesQuotationQueue` | quotation.read | Antrean permintaan B2B per unit/kanal untuk staf penjualan. Filter party juga diperiksa di service. |
| `myStoreOrders` | `ListMyStoreOrders` | Self/B2B | Daftar pembelian buyer berdasarkan proyeksi OMS, nomor, nilai, mata uang dan waktu pembaruan. Gunakan label pesanan toko di UI. |
| `myStoreOrder` | `GetMyStoreOrder` | Self/B2B | Order projection → baris snapshot dan status order, funding, shipment, return, refund secara terpisah; cek akses tiap relasi. |
| `storeOrderQueue` | `ListStoreOrderQueue` | order.read | Filter periode/status → order dalam mandat staf dengan ringkasan minim PII. |
| `orderPaymentInstructions` | `GetOrderPaymentInstructions` | Self/B2B | Instruksi sesi Finance yang cocok order, buyer, currency dan expiry. Hanya tujuan bayar provider yang tervalidasi, tidak ada secret. |
| `orderShipments` | `ListOrderShipmentProjections` | Self/B2B atau fulfillment.read | Order → qty reserved, dispatched, transferred, returned per baris/site dari WMS. Tampilan buyer hanya data pengiriman yang relevan. |
| `orderCancellationRequests` | `ListOrderCancellationRequests` | Self/B2B atau order.read | Order → qty diminta, qty disetujui, alasan hasil OMS. Pending bukan izin release sepihak. |
| `orderReturnRequests` | `ListOrderReturnRequests` | Self/B2B atau return.read | Order → status OMS, penerimaan fisik WMS, penyelesaian. Tampilkan dua status saat hasilnya berbeda. |
| `orderRefunds` | `ListOrderRefundProjections` | Self/B2B atau refund.read | Order → entitlement, pending, selesai, sisa refund terverifikasi dan currency. Pemohon tidak melihat data instrumen. |
| `orderCommercialDocuments` | `ListOrderCommercialDocuments` | Self/B2B atau evidence.read | Order → jenis, nomor tampilan, nilai dan status Finance; akses file invoice melalui izin objek DRM sendiri. |
| `storefrontSalesSummary` | `GetStorefrontSalesSummary` | reporting.read | Rentang terbatas + kanal → conversion, checkout, GMV, order, retur. Cantumkan as-of dan definisi. Revenue/COGS aktual hanya dari Finance. |
| `storefrontReconciliationCases` | `ListStorefrontReconciliationCases` | reconciliation.read | Scope, tipe selisih, umur → kasus pending/unknown/dead letter, owner dan kode aman tanpa payload mentah. |

## Butuh GraphQL Mutation

| Field GraphQL | RPC Commerce | Akses | Input, hasil dan aturan bisnis yang diuji |
|---|---|---|---|
| `updateStorefrontSettings` | `UpdateStorefrontSettings` | settings.manage | Entity, expected version dan opsi terbatas. Credit mensyaratkan B2B. Tolak aktivasi prepaid sebelum gate Finance/WMS/OMS. |
| `createSalesChannel` | `CreateSalesChannel` | channel.manage | Kaitkan kode, nama, entity, host dan lokasi EF yang sah. Host normalisasi harus unik pada mapping aktif. |
| `updateSalesChannel` | `UpdateSalesChannel` | channel.manage | Ubah atribut kanal/aktif dengan version check; penonaktifan menutup transaksi baru, pesanan lama tetap dapat dilayani sesuai mandat. |
| `configureDeliveryMethod` | `ConfigureDeliveryMethod` | channel.manage | Kanal, metode, pickup/location atau rate → revisi. Pickup gratis dengan site sah; provider quote memakai reference konfigurasi. |
| `createStorefrontListing` | `CreateStorefrontListing` | catalog.edit | Draft lokal per channel/slug; validasi SKU/revisi terhadap PLM dan unik slug per kanal. |
| `updateStorefrontListing` | `UpdateStorefrontListing` | catalog.edit | Revisi judul, media, varian, urutan dan deskripsi dengan sanitasi serta versi. Perubahan material wajib melalui validasi publikasi. |
| `publishStorefrontListing` | `PublishStorefrontListing` | catalog.publish | Semua SKU sellable, revisi/UOM dan media valid, kanal aktif. PLM unavailable menunda publikasi. |
| `unpublishStorefrontListing` | `UnpublishStorefrontListing` | catalog.publish | Tarik dari seluruh pembacaan publik dan checkout baru; simpan reason, actor dan snapshot terdahulu. |
| `openBuyerProfile` | `OpenBuyerProfile` | Self | Sesi customer aktif → satu profil principal. Untuk B2B verifikasi party/mandat EF. Replay tidak membuat customer baru. |
| `saveDeliveryAddress` | `SaveDeliveryAddress` | Self | Buat atau perbarui alamat milik buyer, validasi tujuan dan versi. ID alamat pihak lain tidak dapat ditimpa. |
| `removeDeliveryAddress` | `RemoveDeliveryAddress` | Self | Nonaktifkan alamat tersimpan dengan ownership check. Quote aktif diperiksa ulang, snapshot historis tidak berubah. |
| `openShoppingCart` | `OpenShoppingCart` | Cart | Buka atau kembalikan cart aktif per kanal; guest baru mendapat cookie terikat host melalui Gateway. Retry menghasilkan cart yang sama untuk identitas yang sama. |
| `commerceClaimGuestCart` | `ClaimGuestCart` | Self + cookie guest | Klaim sekali setelah login kanal yang sama; bila ada dua cart aktif, gabung SKU sesuai presisi/limit, naikkan versi, invalidasi cookie lama. |
| `addShoppingCartLine` | `AddShoppingCartLine` | Cart | Listing item + quantity + versi → baris dan versi baru. Pastikan SKU aktif dan kanal cocok. |
| `changeShoppingCartQuantity` | `ChangeShoppingCartQuantity` | Cart | Pada versi yang diharapkan, set qty positif sesuai UOM. Dua penulis pada versi sama memberi satu konflik. |
| `removeShoppingCartLine` | `RemoveShoppingCartLine` | Cart | Hanya baris di cart sendiri. Harga/tax/quote dihitung ulang saat checkout berikutnya. |
| `prepareCheckoutQuote` | `PrepareCheckoutQuote` | Self/B2B | Cart+versi, alamat, metode, kode promo opsional. Segel hasil Pricing, tarif dan Finance tax pada digest + TTL; input harga dari browser ditolak. |
| `submitCheckout` | `SubmitCheckout` | Self/B2B | Buyer menerima quote_id, digest, cart_version, stable key. Buat submission pending + outbox atomik; satu OMS order per submission. Quote stale perlu persetujuan baru. |
| `requestSalesQuotation` | `RequestSalesQuotation` | B2B | Party berwenang, kanal, line dan qty → permintaan penawaran. Credit terms tetap menunggu verifikasi Finance. |
| `issueSalesQuotation` | `IssueSalesQuotation` | quotation.issue | Staf berwenang membekukan revisi Pricing/tax/terms dan expiry; diskon diskresioner memerlukan otorisasi Pricing. |
| `acceptSalesQuotation` | `AcceptSalesQuotation` | B2B | Tanda terima revisi dan digest tepat, TTL aktif. Konversi melalui saga submission dengan key stabil; satu quotation hanya dapat menghasilkan satu order. |
| `rejectSalesQuotation` | `RejectSalesQuotation` | B2B | Hanya penawaran issued yang direpresentasikan; simpan alasan aman dan cegah konversi setelah ditolak. |
| `requestOrderPaymentSession` | `RequestOrderPaymentSession` | Self/B2B | OMS accepted, hold cocok dan belum kedaluwarsa → minta sesi Finance. Bila outcome unknown cari sesi lama, hindari tagihan aktif kedua. |
| `requestStoreOrderCancellation` | `RequestStoreOrderCancellation` | Self/B2B atau order.cancel_request | Order lines, qty dan reason → permintaan ke OMS; tunggu fence WMS dan qty aktual sebelum melepas hold atau meminta pengembalian uang. |
| `requestStoreOrderReturn` | `RequestStoreOrderReturn` | Self/B2B atau return.request | Baris dispatch asli, qty dan reason → permintaan ke OMS. Batasi qty kumulatif, window/policy, tidak menerima retur atau menambah stok dari browser. |
| `requestStoreOrderRefund` | `RequestStoreOrderRefund` | refund.request | Staf meminta eksekusi atas entitlement yang sah, receipt asli dan sisa. Approval SoD dan pembayaran pada Finance; timeout tetap mengunci sisa. |
| `requestStorefrontReconciliation` | `RequestStorefrontReconciliation` | reconciliation.request | Case terpilih + reason → pekerjaan lookup/replay idempoten, audit dan correlation. Tidak menyediakan pengubahan manual status uang/stok. |

## Penjelasan RPC privat dan integrasi

Pisahkan handler browser dari consumer dengan allowlist identitas workload mTLS per produsen, tujuan, tipe event, versi dan scope. Nama consumer berikut usulan privat. Tidak perlu menerbitkan GraphQL untuk menjalankannya.

| Fungsi privat | Bukti yang boleh diterima dan akibatnya |
|---|---|
| `ApplyProductRevisionEvent`, `ApplyPriceRevisionEvent` | Dari PLM/Pricing. Tandai proyeksi atau quote terdampak stale, cegah checkout baru memakai versi lama, pertahankan snapshot order. |
| `DispatchOrderSubmission`, `ApplyOrderDecision` | Kirim submission ke OMS dengan key stabil. Hanya keputusan OMS yang mengisi order ref/status, termasuk cancellation/return approval. |
| `ApplyReservationEffect`, `ApplyFulfillmentEffect` | Dari WMS. Cocokkan order, line, revision, qty, site, effect ID, authorization dan reversal asal. Bedakan reservasi, pengiriman, transfer kontrol dan penerimaan retur. |
| `ApplyFinanceEffect` | Dari Finance. Bedakan receipt, applied funding, credit approval, refund dan reversal. Validasi order, customer, entity, currency, amount, original receipt dan source version. |
| `IssueCommercialEvidence`, `ApplyPostingConfirmation` | Evidence sale/credit unik untuk transfer kontrol/koreksi yang sah per baris. Finance memvalidasi dan memposting per dokumen parsial, WMS mengirim stock issue ke Finance untuk COGS. |
| `ExpireCheckoutAttempt`, `ReconcileUnknownOutcome`, `RelayOutbox` | Worker dengan lease dan backoff. Lookup owner sebelum retry, kompensasi bagian sukses saja, akhiri dengan outcome owner atau kasus rekonsiliasi yang dapat ditelusuri. |

## Inventaris yang harus dipelihara agen

Saat coding, anotasi **setiap nama** dengan fase, path proto/service/adapter/SDL/resolver, kode IAM aktual, kebijakan objek, penanganan error, contoh hasil untuk UI, skenario tes, bukti dan status. Tambahkan RPC/consumer yang ditemukan dan alasan untuk tiap operasi privat atau ditunda. Hitung usulan, terpasang, diuji HTTPS dan tertahan secara terpisah. Field global `node`, nested resolver, workload dan worker tidak ikut hitungan root domain. Cantumkan operasi di bawah judul `## Penjelasan {operation_name}`, `## Butuh GraphQL Query/Mutation` dan `## RBAC` saat memecah katalog per subdomain di `agents/grpc/web_commerce/{id}-{title}.md`.
