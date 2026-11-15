# Enterprise Foundation — Legal Entity, Supplier/Vendor Party dan Lokasi

> **Dokumen management retail · 12 November 2026.** Kepemilikan bisnis normatif mengikuti [agents/boundary.md](../../agents/boundary.md). Ini ringkasan, bukan instruksi melakukan perubahan lintas management atau seed IAM otomatis.

## Satu sumber data organisasi
Foundation memiliki legal entities, organization units, locations/addresses dan party identity (supplier, vendor, customer maupun mitra lain). Ia **tidak** menulis hubungan kerja HRM, role IAM, PO Procurement, saldo WMS, atau GL Finance.

## Internal F1
Tentukan badan hukum penjual dan pemilik warehouse, unit organisasi, alamat billing/pickup, serta party/identifier resmi supplier. Hubungkan ID stabil dan source; kontrak lintas domain tidak boleh membuat tabel FK ke database Foundation.

## Supplier/vendor F2
Party adalah bisnis/pihak; representative adalah **manusia principal IAM**. Procurement dapat menyimpan status kualifikasi dan kontrak sourcing; IAM menyetujui akses. Satu orang mewakili beberapa party hanya dengan mandat dan grant berbeda; pemilihan company UI bukan bukti hak.

## Validasi transaksi
Saat PO/ASN/quote/stock diproses, baca relasi legal entity/party yang valid dari fakta owner, verifikasi status dan versi sesuai kontrak. Jika Foundation down ketika verifikasi wajib, status menjadi unavailable/pending, bukan asal menerima UUID client.

## Source
[Enterprise Foundation README](../../backend/services/enterprise_foundation/README.md), [boundary](../../agents/boundary.md#enterprise_foundation). Status runtime/GraphQL dibuktikan per domain dan tes; data source lama pada docs historis tidak mengesahkan readiness otomatis.
