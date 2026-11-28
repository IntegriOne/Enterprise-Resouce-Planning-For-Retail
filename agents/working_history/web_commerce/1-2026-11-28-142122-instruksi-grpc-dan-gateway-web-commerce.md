# Handoff rancangan Web Commerce

Disusun [instruksi DB hingga Gateway](../../instructions/web_commerce/1-leader-implementasi-grpc-dan-gateway-web-commerce-instructions.md) dan [inventaris 27 Query serta 27 Mutation](../../grpc/web_commerce/1-inventaris-operasi-grpc-dan-graphql-web-commerce.md). Semua 43 tabel dari sembilan kelompok DBML dipetakan ke pemilik, alur permintaan atau proyeksi. Keranjang guest memerlukan penyesuaian skema serta klaim sekali setelah login. Pricing memutuskan harga, OMS memegang order, WMS mengendalikan stok dan Finance mengeksekusi pembayaran. Publikasi checkout menunggu kontrak pemilik dan E2E bersama.

Seluruh tautan diperiksa pada checkout lokal. Nama root usulan unik dan tidak berbenturan dengan SDL saat pemeriksaan. Dokumen logika dan DBML tersedia sebagai berkas lokal yang belum dilacak Git. Implementasi runtime dan E2E diserahkan melalui instruksi.
