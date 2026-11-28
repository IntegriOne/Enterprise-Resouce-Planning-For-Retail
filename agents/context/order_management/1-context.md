# Konteks Order Management (OMS)

## Apa yang sudah dikerjakan

- Scope produk menetapkan OMS sebagai pemilik sales order dan keputusan menerima, menolak, membatalkan, atau menyetujui retur. Web Commerce menerima pengajuan dan menampilkan status, WMS memegang stok dan fulfillment, dan Finance memegang fakta pembayaran. [Batas management](../../boundary.md#order_management) · [Desain OMS dan Commerce](../../../docs/03-managements/oms-and-commerce.md)
- Pada source lokal Web Commerce yang diperiksa 29 November 2026, `OMSPort` mendefinisikan `SubmitOrder` dan `LookupOrderBySubmission`. `SubmitCheckout` menyimpan percobaan dalam status `pending` dan event `CommerceCheckoutSubmitted` dalam satu transaksi. Penggunaan kembali kunci dengan payload berbeda ditolak. Wiring saat ini membiarkan port OMS kosong, dan pengaturan prepaid checkout membutuhkan port OMS serta kesiapan Finance. Bukti lokal ada di `backend/services/web_commerce/domain/service/ports.go`, `services/checkout/service.go`, `services/storefront/service.go`, dan `cmd/wiring.go` dalam modul yang sama.

## Apa yang belum dikerjakan

- Checkout belum memuat modul OMS di `backend/services/` atau entri OMS di `backend/go.work`. Kontrak producer OMS, database sales order, pemrosesan keputusan, wiring Gateway, katalog izin, dan bukti E2E lintas pemilik belum tersedia dalam cakupan yang diperiksa.

## Apa yang harus dikerjakan selanjutanya

- Saat tahap OMS dibuka, sepakati kontrak submit dan lookup dengan Web Commerce, keputusan pembatalan dan retur, serta penerusan state berversi. Gunakan kunci pengajuan yang stabil untuk percobaan ulang setelah timeout dan sertakan korelasi sumber untuk order marketplace.
- Buktikan satu sales order untuk replay identik, konflik untuk key dengan payload berbeda, penolakan akses lintas scope, dan rekonsiliasi WMS serta Finance sebelum mengaktifkan checkout lintas layanan.
