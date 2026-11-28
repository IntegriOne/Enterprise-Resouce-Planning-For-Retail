# Konteks Customer Service Management

## Apa yang sudah dikerjakan

- Scope produk menetapkan Customer Service sebagai pemilik tiket, komplain, prioritas, dan eskalasi pelanggan. Status order, pengiriman, pembayaran, dan dokumen dibaca sesuai izin dari OMS, WMS, Finance, dan Document & Records Management. [Batas management](../../boundary.md#customer_service_management) · [Desain CRM dan Customer Service](../../../docs/03-managements/crm-and-customer-service.md)
- Pada source lokal Web Commerce yang diperiksa 29 November 2026, `RequestStoreOrderCancellation` dan `RequestStoreOrderReturn` mencatat permintaan transaksi purnajual dengan status `requested`. Bukti ini memberi titik integrasi untuk eskalasi, sementara tiket dan status penyelesaiannya menjadi tanggung jawab Customer Service. Sumber lokal ada di `backend/services/web_commerce/services/aftersales/service.go`.

## Apa yang belum dikerjakan

- Checkout belum memuat modul Customer Service di `backend/services/` atau entri Customer Service di `backend/go.work`. Lifecycle tiket, penugasan petugas, kontrak akses ke transaksi pemilik, wiring Gateway, katalog izin, dan bukti E2E Customer Service belum tersedia dalam cakupan yang diperiksa.

## Apa yang harus dikerjakan selanjutanya

- Saat tahap Customer Service dibuka, tetapkan lifecycle tiket, hak akses pemohon dan petugas yang ditugaskan, serta data minimum untuk eskalasi. Gunakan referensi logis dan kontrak pemilik ketika tiket terkait order, pengiriman, pembayaran, atau dokumen.
- Uji tiket sendiri, larangan akses pelanggan lain, dan eskalasi saat dependensi gagal. Penutupan tiket dan keputusan retur atau refund masing-masing perlu bukti dari pemiliknya.
