# Privacy Policy, Terms of Service dan DPA — Kerangka

> **DRAFT UNTUK REVIEW HUKUM — bukan dokumen siap ditandatangani atau dipublikasikan.** Legal/owner/klien wajib memverifikasi regulasi Indonesia yang berlaku, peran pengendali/prosesor dan syarat marketplace resmi. Tidak ada legal review eksternal yang dijalankan pada rekonstruksi dokumentasi ini.

## Privacy Policy: wajib diisi per merchant
`Data controller identity, address, privacy contact, processor/subprocessor, purpose & legal basis, collected categories, retention, security, data subject rights, sharing/cross-border transfer, incident contact, cookie categories, effective date`.

Data internal staff/vendor/supplier: identity, status perusahaan/party, PO/kontrak, audit; customer: akun, cart, alamat, order/pembayaran sesuai kewajiban; analytics: event dengan consent atau legal basis yang benar. Jangan otomatis mengumpulkan NIK/NPWP, TOTP secret, session raw atau data pembayaran di event/log yang tidak membutuhkan. Portal support tidak boleh membaca data customer di luar assigned case.

## Terms: merchant dan pembeli
Identitas penjual, produk/varian, harga, promo/expiry, perhitungan tax/ongkir, order acceptance (PENDING ≠ ACCEPTED), payment options aktual, shipping, hak konsumen, cancel/retur/refund dan mekanisme pengaduan. Jangan menghilangkan hak wajib konsumen lewat disclaimer; semua proses dan syarat harus direview hukum sebelum publikasi.

## Data Processing Agreement (DPA)
Jika Integrione memproses PII atas nama merchant, tentukan pengendali vs prosesor, instruksi pemrosesan, kategori subjek, data lokasi/subprosesor, audit, breach notification, retensi, exit/erasure dan hak transfer. Model dedicated hosting vs shared tenancy memengaruhi kewajiban dan harus diselaraskan dengan ADR/deployment.

## Cookie/Marketplace
Cookie otentikasi opaque adalah kategori esensial; analytics/marketing cookies memerlukan tujuan dan dasar yang sesuai, tidak otomatis diaktifkan. Shopee/TikTok Shop dan kurir/payment providers mempunyai terms yang harus diperiksa, termasuk penggunaan kembali PII customer dan batas retensi.

## Status penetapan
`legal reviewer [ ] | merchant legal identity [ ] | privacy contact [ ] | version [ ] | effectivity [ ] | approval/signatures [ ]`. Sampai terisi dan ditinjau, dokumen ini **bukan** kebijakan publik.
