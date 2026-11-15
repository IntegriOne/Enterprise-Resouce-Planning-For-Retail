# Register Kewajiban Hukum dan Compliance

> **DRAFT UNTUK REVIEW HUKUM — bukan dokumen siap ditandatangani atau dipublikasikan.** Legal/owner/klien wajib memverifikasi regulasi Indonesia yang berlaku, peran pengendali/prosesor dan syarat marketplace resmi. Tidak ada legal review eksternal yang dijalankan pada rekonstruksi dokumentasi ini.

## Cara memakai
Lakukan penelaahan legal untuk applicability berdasarkan jenis merchant, data yang diproses, lokasi hosting, kategori produk, dan layanan marketplace/payment. Daftar berikut **calon area pemeriksaan**, bukan pernyataan produk sudah compliant.

| ID | Area | Bukti/periksa ke pihak berwenang |
|---|---|---|
| LC-01 | UU PDP | Peran controller/processor, lawful purpose, subject rights, retention, breach response dan transfer lintas negara |
| LC-02 | Perlindungan konsumen/perdagangan online | Identitas seller, informasi produk, price, order acceptance, return/refund, keluhan |
| LC-03 | PSE dan sistem elektronik | Applicability pendaftaran/keamanan sesuai model layanan yang dipilih |
| LC-04 | Pajak dan pembukuan | Invoice/PPN/e-faktur dan kewajiban merchant menurut karakter usahanya |
| LC-05 | Shopee/TikTok Shop developer terms | App access, APIs, signature, webhooks, data license, rate limits, trademark, retention |
| LC-06 | Kurir/payment integrasi | Izin/kontrak, bukti transaksi, escrow/settlement, refund dan perlindungan PII |
| LC-07 | Intellectual property | Foto produk, brand, ikon, font, source code, asset license |
| LC-08 | Analytics, cookies dan marketing | Consent/legal basis, segmentation, retargeting, cross-channel data reuse |
| LC-09 | Produk teregulasi | BPOM/halal/kategori barang berbahaya/expired lots bila merchant relevan |

## Kolom register
`Regulation source official URL/number/version | date checked | applicability | requirement | data owner | control/test | evidence | risk | legal reviewer | approved status`. Status `OPEN|REVIEW|VERIFIED|NOT_APPLICABLE_WITH_REASON|BLOCKED`.

## Launch gate
Privacy/Terms/DPA/contract review selesai, consent/purpose jelas, provider terms valid, backup/incident/security/rights flow berjalan, owner sign-off. Jangan menyebut software “bersertifikat” hanya karena memakai standar referensi di DBML.
