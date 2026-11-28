# Konteks CRM & Sales Management

## Apa yang sudah dikerjakan

- Scope produk menetapkan CRM sebagai pemilik lead, prospek, pipeline, hubungan pelanggan, dan kampanye dengan opt-in. Akun kanal dikelola Web Commerce, identitas autentikasi dikelola IAM, dan order serta piutang dikelola OMS dan Finance. [Batas management](../../boundary.md#crm_sales_management) · [Desain CRM dan Customer Service](../../../docs/03-managements/crm-and-customer-service.md)
- Pada source lokal Web Commerce yang diperiksa 29 November 2026, `RequestSalesQuotation` menerima `crm_opportunity_reference_id` opsional di kontrak `quotation_order_aftersales.proto`. Handler mengurai UUID dan service menyimpannya pada quote draf. Ini adalah titik referensi untuk integrasi mendatang. Jalur yang dibaca belum meminta verifikasi opportunity ke CRM. Bukti lokal ada di `backend/services/web_commerce/contract/proto/web_commerce/v1/quotation_order_aftersales.proto`, `transport/grpc/handler/quotation_order_aftersales.go`, dan `services/quotation/service.go` dalam modul yang sama.

## Apa yang belum dikerjakan

- Checkout belum memuat modul CRM di `backend/services/` atau entri CRM di `backend/go.work`. Penyimpanan lead dan opportunity, pengelolaan consent, kontrak verifikasi untuk konsumen, wiring Gateway, katalog izin, dan bukti E2E CRM belum tersedia dalam cakupan yang diperiksa.

## Apa yang harus dikerjakan selanjutanya

- Saat tahap CRM dibuka, tetapkan model lead dan opportunity, sumber identitas, scope badan hukum atau party, dan aturan consent. Sepakati dengan Web Commerce cara memverifikasi ID opportunity serta menangani ID tidak dikenal, scope salah, dan CRM tidak tersedia.
- Uji isolasi scope dan perubahan atau penarikan consent sebelum mengaktifkan kampanye atau menggunakan kembali data yang berasal dari marketplace.
