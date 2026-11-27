# Daftar slip gaji mandiri

Chrome headless diuji lewat HTTPS development. Permintaan anonim ke halaman akun diarahkan ke login. Sesi superadmin dengan izin baca pekerja dan badan hukum, yang diberikan sementara melalui GraphQL, ditolak oleh query slip karena tidak memiliki akses mandiri.

Respons daftar berhasil disimulasikan di browser. Uji memeriksa muat ulang setelah kegagalan, paginasi, label status Terjadwal, tanggal penerbitan dan bayar yang berbeda, serta ketiadaan `gid://` di area utama meski respons berisi ID internal. Tampilan metadata berhasil diuji dengan fixture. Tes belum mencakup query sukses terhadap data slip riil atau pembukaan PDF dari DRM.

## Sesi karyawan dengan scope sendiri

Pada stack HTTPS terisolasi, superadmin membuat pekerja dan mengajukan akun. IAM memproses `createEmployeeUser` setelah MFA dan HRM menyimpan tautan principal. Superadmin memberikan izin slip lewat mutasi GraphQL. IAM menolak assignment SELF dengan ID scope yang berbeda. Superadmin yang hanya memiliki grant GLOBAL ditolak Gateway. Karyawan tertaut dengan assignment SELF yang benar lolos otorisasi dan mencapai HRM. Chrome membuka `/account` dengan sesi karyawan tersebut tanpa izin operator, lalu menampilkan pesan yang bisa dicoba ulang karena verifier slip belum tersedia. Area utama tidak menampilkan `gid://`. Data uji IAM dan HRM dibersihkan setelah runner selesai.

Uji penugasan SELF dan halaman karyawan diulang setelah kedua replika Gateway memakai build dengan validasi metadata. Keduanya sehat. Uji baca pekerja juga berhasil melalui HTTPS development. Penolakan metadata rusak dibuktikan oleh tes Gateway dengan gRPC fixture.

Regresi akses diulang setelah perbaikan kecocokan periode dipasang pada image HRM development. Uji baca pekerja pada development lulus. Pada stack terisolasi dengan image HRM baru, akun karyawan tertaut dengan grant SELF mencapai halaman akun dan menerima pesan kegagalan verifier yang bisa dicoba ulang. Periode yang berbeda ditolak oleh regresi gRPC dan PostgreSQL menggunakan verifier fixture.

Regresi tampilan waktu memakai fixture `2026-11-24T23:58:00Z` dan tanggal bayar terjadwal 25 November. Chrome memastikan kolom **Waktu terbit (UTC)** menampilkan 24 November pukul 23:58, nilai waktu HTML mempertahankan timestamp sumber, dan tanggal bayar tetap 25 November. Kedua replika frontend sehat.

## Menyiapkan PDF untuk dicetak

Chrome lewat HTTPS memakai respons GraphQL simulasi untuk daftar dan isi PDF. Percobaan pertama menampilkan pesan coba ulang. Percobaan kedua membentuk dokumen `application/pdf` dengan byte yang sama dengan fixture dan menampilkan tautan **Buka PDF untuk dicetak**. Menutup panel mencabut URL dokumen dari browser. Fixture dibaca dengan `pdftotext` dan area utama tidak menampilkan ID internal. Uji ini memeriksa penyajian sampai tautan cetak. Pembukaan dokumen riil dari DRM dan pencetakan masih terbuka.

Regresi berikutnya mengeklik tautan **Buka PDF untuk dicetak** di Chrome lewat HTTPS. Browser membuka tab terpisah pada URL Blob dari PDF fixture yang tervalidasi. Setelah tab ditutup dan panel dihentikan, URL Blob tidak dapat diakses lagi. Tes ini membuktikan pembukaan tab dokumen. Uji pencetakan fisik dan pembacaan data DRM produksi masih terbuka.

Setelah batas ukuran PDF disamakan di HRM dan Gateway, kedua replika Gateway dan HRM development sehat. Tes guard payroll melalui HTTPS lulus. Chrome juga mengulang pembukaan tab dan pencabutan URL dengan PDF fixture pada build aktif.
